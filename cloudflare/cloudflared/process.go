package cloudflared

// Process identifies a running cloudflared process.
type Process struct {
	PID     string
	Exe     string
	Cmdline string
}
