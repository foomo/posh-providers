package proxy

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Logger is a minimal interface satisfied by posh/pkg/log.Logger.
type Logger interface {
	Info(a ...any)
}

type Proxy struct {
	SSHHost       string `json:"sshHost"       yaml:"sshHost"`
	SSHUser       string `json:"sshUser"       yaml:"sshUser"`
	SSHPort       int    `json:"sshPort"       yaml:"sshPort"`
	IdentityFile  string `json:"identityFile"  yaml:"identityFile"`  // path to SSH private key (-i)
	IdentityAgent string `json:"identityAgent" yaml:"identityAgent"` // SSH agent socket path (-o IdentityAgent)
	LocalPort     int    `json:"localPort"     yaml:"localPort"`     // 0 = auto-assign
	NoProxy       string `json:"noProxy"       yaml:"noProxy"`
}

// Config is a named map of proxy entries, typically loaded via
// viper.UnmarshalKey("proxies", &cfg) in each provider constructor.
type Config map[string]Proxy

func (c Config) Names() []string {
	names := make([]string, 0, len(c))
	for name := range c {
		names = append(names, name)
	}

	return names
}

// Start opens an SSH SOCKS5 tunnel for the named proxy entry and returns
// environment variables to inject plus a stop function that tears everything
// down. Callers must always invoke stop() when the proxied command finishes.
//
// HTTP_PROXY and HTTPS_PROXY are set to an HTTP CONNECT proxy that wraps the
// SOCKS5 tunnel, so non-Go subprocesses (Python, Node, etc.) work without
// SOCKS5 library support. ALL_PROXY carries the raw socks5h:// address for
// tools that prefer it. DNS resolution happens on the jump host in both cases.
func (c Config) Start(ctx context.Context, l Logger, name string) ([]string, func(), error) {
	env, _, _, stop, err := c.start(ctx, l, name, false)

	return env, stop, err
}

// StartWithDockerProxy is like Start but also configures Docker Desktop to
// route daemon connections through the tunnel. Use this for commands that
// involve docker operations (e.g. az acr login) where the Docker Desktop
// daemon ignores process-level env vars.
//
// Docker Desktop's daemon runs in a Linux VM. Its connections go through an
// internal HTTP proxy on the macOS host (http.docker.internal:3128). This
// method routes that proxy's upstream through the HTTP CONNECT proxy that
// Start already binds, by updating Docker Desktop's settings via its backend
// API — no restart required.
func (c Config) StartWithDockerProxy(ctx context.Context, l Logger, name string) ([]string, func(), error) {
	env, _, _, stop, err := c.start(ctx, l, name, true)

	return env, stop, err
}

// Addr starts the named SSH SOCKS5 tunnel and returns the local address
// ("localhost:PORT"), the SSH process PID, and a stop function. Use Start
// instead when env-var injection is needed.
func (c Config) Addr(ctx context.Context, l Logger, name string) (string, int, func(), error) {
	_, addr, pid, stop, err := c.start(ctx, l, name, false)

	return addr, pid, stop, err
}

func (c Config) start(ctx context.Context, l Logger, name string, withDocker bool) ([]string, string, int, func(), error) {
	p, ok := c[name]
	if !ok {
		return nil, "", 0, func() {}, errors.Errorf("proxy not found: %s", name)
	}

	localPort := p.LocalPort
	if localPort == 0 {
		port, err := freePort(ctx)
		if err != nil {
			return nil, "", 0, func() {}, errors.Wrap(err, "find free port")
		}

		localPort = port
	}

	host := p.SSHHost
	if p.SSHUser != "" {
		host = p.SSHUser + "@" + host
	}

	args := []string{
		"-D", fmt.Sprintf("%d", localPort),
		"-N",
		"-o", "BatchMode=yes",
		"-o", "ExitOnForwardFailure=yes",
		"-o", "ServerAliveInterval=10",
		"-o", "ServerAliveCountMax=3",
	}
	if p.SSHPort > 0 && p.SSHPort != 22 {
		args = append(args, "-p", fmt.Sprintf("%d", p.SSHPort))
	}

	if p.IdentityFile != "" {
		args = append(args, "-i", expandPath(p.IdentityFile))
	}

	if p.IdentityAgent != "" {
		args = append(args, "-o", "IdentityAgent="+expandPath(p.IdentityAgent))
	}

	args = append(args, host)

	cmd := exec.CommandContext(ctx, "ssh", args...)
	cmd.Env = os.Environ()

	if err := cmd.Start(); err != nil {
		return nil, "", 0, func() {}, errors.Wrap(err, "start ssh tunnel")
	}

	proxyAddr := fmt.Sprintf("localhost:%d", localPort)

	pid := cmd.Process.Pid

	l.Info("waiting for SSH proxy tunnel at", proxyAddr, "pid", pid)

	if err := waitForPort(ctx, proxyAddr, 15*time.Second); err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()

		return nil, "", 0, func() {}, errors.Wrap(err, "ssh tunnel not ready")
	}

	l.Info("SSH proxy tunnel ready at", proxyAddr, "pid", pid)

	socks5URL := "socks5h://" + proxyAddr

	stops := []func(){
		func() {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()

			l.Info("SSH proxy tunnel closed at", proxyAddr, "pid", pid)
		},
	}

	// Always start an HTTP CONNECT proxy wrapping the SOCKS5 tunnel. Non-Go
	// subprocesses (e.g. Python-based Azure CLI) do not understand socks5h://
	// and need a plain http:// proxy URL to route HTTPS traffic correctly.
	httpAddr, stopHTTP, err := startHTTPConnectProxy(ctx, proxyAddr)
	if err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()

		return nil, "", 0, func() {}, errors.Wrap(err, "start http connect proxy")
	}

	l.Info("HTTP CONNECT proxy ready at", httpAddr, "pid", pid)

	stops = append(stops, func() {
		stopHTTP()

		l.Info("HTTP CONNECT proxy closed at", httpAddr, "pid", pid)
	})

	_, httpPort, _ := net.SplitHostPort(httpAddr)
	httpProxyURL := "http://127.0.0.1:" + httpPort

	if withDocker {
		// Docker Desktop's daemon runs in a Linux VM and only supports HTTP/HTTPS
		// proxies. Route it through the HTTP CONNECT proxy via the backend API —
		// no restart required.
		restore, err := applyDockerDesktopProxy(ctx, httpProxyURL, p.NoProxy)
		if err != nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()

			return nil, "", 0, func() {}, errors.Wrap(err, "configure docker desktop proxy")
		}

		l.Info("Docker Desktop proxy set to", httpProxyURL, "pid", pid)

		stops = append(stops, func() {
			restore()

			l.Info("Docker Desktop proxy restored", "pid", pid)
		})
	}

	stop := func() {
		for _, fn := range stops {
			fn()
		}
	}

	env := []string{
		// Use the HTTP CONNECT proxy for HTTP_PROXY/HTTPS_PROXY so that
		// non-Go subprocesses (Python, Node, etc.) can route through the
		// tunnel without needing SOCKS5 library support.
		"HTTP_PROXY=" + httpProxyURL,
		"http_proxy=" + httpProxyURL,
		"HTTPS_PROXY=" + httpProxyURL,
		"https_proxy=" + httpProxyURL,
		// Keep the raw SOCKS5 address in ALL_PROXY for tools that prefer it.
		"ALL_PROXY=" + socks5URL,
		"all_proxy=" + socks5URL,
	}
	if p.NoProxy != "" {
		env = append(env, "NO_PROXY="+p.NoProxy, "no_proxy="+p.NoProxy)
	}

	return env, proxyAddr, pid, stop, nil
}

// applyDockerDesktopProxy configures Docker Desktop's daemon to route
// connections through proxyURL by calling the Docker Desktop backend API.
// Docker Desktop applies the change immediately without requiring a restart.
// The returned function restores the original proxy settings.
func applyDockerDesktopProxy(ctx context.Context, proxyURL, noProxy string) (func(), error) {
	sockPath := filepath.Join(os.Getenv("HOME"),
		"Library/Containers/com.docker.docker/Data/backend.sock")

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, "unix", sockPath)
			},
		},
	}

	// Read current settings to save original proxy values.
	getReq, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost/app/settings", nil)
	if err != nil {
		return func() {}, errors.Wrap(err, "build docker desktop settings request")
	}

	resp, err := client.Do(getReq)
	if err != nil {
		return func() {}, errors.Wrap(err, "read docker desktop settings")
	}
	defer resp.Body.Close()

	var settings map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&settings); err != nil {
		return func() {}, errors.Wrap(err, "decode docker desktop settings")
	}

	origHTTP, origHTTPS, origMode, origExclude := dockerProxyValues(settings)

	restore := func() { //nolint:contextcheck
		payload := buildDockerProxyPayload(origHTTP, origHTTPS, origMode, origExclude)
		body, _ := json.Marshal(payload) //nolint:errchkjson
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "http://localhost/app/settings",
			bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		r, err := client.Do(req)
		if err == nil {
			r.Body.Close()
		}
	}

	exclude := noProxy
	if origExclude != "" && exclude != "" {
		exclude = origExclude + "," + exclude
	} else if origExclude != "" {
		exclude = origExclude
	}

	payload := buildDockerProxyPayload(proxyURL, proxyURL, "manual", exclude)

	body, err := json.Marshal(payload)
	if err != nil {
		return func() {}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost/app/settings",
		bytes.NewReader(body))
	if err != nil {
		return func() {}, err
	}

	req.Header.Set("Content-Type", "application/json")

	r, err := client.Do(req)
	if err != nil {
		return func() {}, errors.Wrap(err, "update docker desktop settings")
	}

	r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return restore, errors.Errorf("docker desktop settings returned %d", r.StatusCode)
	}

	return restore, nil
}

func buildDockerProxyPayload(httpVal, httpsVal, mode, exclude string) map[string]any {
	return map[string]any{
		"vm": map[string]any{
			"proxy": map[string]any{
				"http":    map[string]any{"value": httpVal},
				"https":   map[string]any{"value": httpsVal},
				"mode":    map[string]any{"value": mode},
				"exclude": map[string]any{"value": exclude},
			},
		},
	}
}

func dockerProxyValues(settings map[string]any) (string, string, string, string) {
	vm, _ := settings["vm"].(map[string]any)
	if vm == nil {
		return "", "", "", ""
	}

	proxy, _ := vm["proxy"].(map[string]any)
	if proxy == nil {
		return "", "", "", ""
	}

	return nestedValue(proxy, "http"), nestedValue(proxy, "https"), nestedValue(proxy, "mode"), nestedValue(proxy, "exclude")
}

func nestedValue(m map[string]any, key string) string {
	if v, ok := m[key].(map[string]any); ok {
		if s, ok := v["value"].(string); ok {
			return s
		}
	}

	return ""
}

// startHTTPConnectProxy starts a minimal HTTP CONNECT proxy on a free local
// port. Each CONNECT request is tunnelled through the SOCKS5 proxy at
// socks5Addr using remote hostname resolution (socks5h behaviour).
func startHTTPConnectProxy(ctx context.Context, socks5Addr string) (string, func(), error) {
	ln, err := (&net.ListenConfig{}).Listen(ctx, "tcp", "127.0.0.1:0")
	if err != nil {
		return "", func() {}, err
	}

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}

			go serveHTTPConnect(ctx, conn, socks5Addr)
		}
	}()

	return ln.Addr().String(), func() { _ = ln.Close() }, nil
}

func serveHTTPConnect(ctx context.Context, client net.Conn, socks5Addr string) {
	defer client.Close()

	br := bufio.NewReader(client)

	// Read request line: "CONNECT host:port HTTP/1.1"
	line, err := br.ReadString('\n')
	if err != nil {
		return
	}

	parts := strings.Fields(strings.TrimSpace(line))
	if len(parts) < 2 || parts[0] != "CONNECT" {
		fmt.Fprintf(client, "HTTP/1.1 405 Method Not Allowed\r\n\r\n")
		return
	}

	target := parts[1]

	// Drain remaining headers
	for {
		l, err := br.ReadString('\n')
		if err != nil || strings.TrimSpace(l) == "" {
			break
		}
	}

	tunnel, err := dialSocks5h(ctx, socks5Addr, target)
	if err != nil {
		fmt.Fprintf(client, "HTTP/1.1 502 Bad Gateway\r\n\r\n")
		return
	}

	defer tunnel.Close()

	fmt.Fprintf(client, "HTTP/1.1 200 Connection Established\r\n\r\n")

	done := make(chan struct{}, 1)

	go func() {
		_, _ = io.Copy(tunnel, io.MultiReader(br, client))

		done <- struct{}{}
	}()

	_, _ = io.Copy(client, tunnel)

	<-done
}

// dialSocks5h connects to target ("host:port") through the SOCKS5 proxy at
// proxyAddr, asking the proxy to resolve the hostname (socks5h / atyp 0x03).
func dialSocks5h(ctx context.Context, proxyAddr, target string) (net.Conn, error) {
	host, portStr, err := net.SplitHostPort(target)
	if err != nil {
		return nil, err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}

	if len(host) > 255 {
		return nil, errors.New("hostname too long for socks5")
	}

	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", proxyAddr)
	if err != nil {
		return nil, err
	}

	// Greeting: version 5, one method, no-auth
	if _, err = conn.Write([]byte{0x05, 0x01, 0x00}); err != nil {
		conn.Close()

		return nil, err
	}

	resp := make([]byte, 2)
	if _, err = io.ReadFull(conn, resp); err != nil || resp[0] != 0x05 || resp[1] != 0x00 {
		conn.Close()

		return nil, errors.New("socks5 handshake failed")
	}

	// CONNECT request: ver=5, cmd=CONNECT, rsv=0, atyp=domain(3)
	hostB := []byte(host)
	req := make([]byte, 0, 6+len(hostB))
	req = append(req, 0x05, 0x01, 0x00, 0x03, byte(len(hostB))) //nolint:gosec
	req = append(req, hostB...)
	req = append(req, byte(port>>8), byte(port)) //nolint:gosec

	if _, err = conn.Write(req); err != nil {
		conn.Close()

		return nil, err
	}

	// Response: ver, rep, rsv, atyp, ...
	head := make([]byte, 4)
	if _, err = io.ReadFull(conn, head); err != nil {
		conn.Close()

		return nil, err
	}

	if head[1] != 0x00 {
		conn.Close()

		return nil, fmt.Errorf("socks5 connect failed: code %d", head[1])
	}

	// Drain bound address from response
	switch head[3] {
	case 0x01:
		_, _ = io.ReadFull(conn, make([]byte, 6)) // IPv4 (4) + port (2)
	case 0x03:
		n := make([]byte, 1)
		_, _ = io.ReadFull(conn, n)
		_, _ = io.ReadFull(conn, make([]byte, int(n[0])+2)) // domain + port
	case 0x04:
		_, _ = io.ReadFull(conn, make([]byte, 18)) // IPv6 (16) + port (2)
	}

	return conn, nil
}

// expandPath expands a leading ~ to the user's home directory.
// The shell does not expand ~ when Go calls exec directly.
func expandPath(p string) string {
	if strings.HasPrefix(p, "~/") {
		home, _ := os.UserHomeDir()

		return filepath.Join(home, p[2:])
	}

	return p
}

func freePort(ctx context.Context) (int, error) {
	ln, err := (&net.ListenConfig{}).Listen(ctx, "tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}

	defer ln.Close()

	return ln.Addr().(*net.TCPAddr).Port, nil
}

func waitForPort(ctx context.Context, addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		conn, err := (&net.Dialer{Timeout: time.Second}).DialContext(ctx, "tcp", addr)
		if err == nil {
			conn.Close()

			return nil
		}

		time.Sleep(300 * time.Millisecond)
	}

	return errors.Errorf("timeout waiting for %s", addr)
}
