package beam

import (
	"fmt"
	"net"
	"os/exec"
	"time"

	"github.com/foomo/posh-providers/onepassword"
	"github.com/foomo/posh/pkg/log"
	"github.com/spf13/viper"
)

type (
	Beam struct {
		l         log.Logger
		cfg       Config
		op        *onepassword.OnePassword
		configKey string
	}
	Option func(*Beam) error
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func CommandWithConfigKey(v string) Option {
	return func(o *Beam) error {
		o.configKey = v
		return nil
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

// NewBeam command
func NewBeam(l log.Logger, op *onepassword.OnePassword, opts ...Option) (*Beam, error) {
	inst := &Beam{
		l:         l,
		op:        op,
		configKey: "beam",
	}
	for _, opt := range opts {
		if opt != nil {
			if err := opt(inst); err != nil {
				return nil, err
			}
		}
	}
	if err := viper.UnmarshalKey(inst.configKey, &inst.cfg); err != nil {
		return nil, err
	}

	return inst, nil
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (t *Beam) Config() Config {
	return t.cfg
}

func (t *Beam) Start() {
	t.l.Info("Starting beam tunnels")
	for _, tunnel := range t.cfg {
		for _, cluster := range tunnel.Clusters {
			go t.tunnel(cluster.Hostname, cluster.Port)
		}
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (t *Beam) tunnel(hostname string, port int) {
	for {
		addr := fmt.Sprintf("127.0.0.1:%d", port)
		if _, err := net.DialTimeout("tcp", addr, time.Second); err == nil {
			t.l.Debug("tunnel/port already exists", "addr", addr, "err", err)
			time.Sleep(10 * time.Second)
			continue
		}
		cmd := exec.Command("cloudflared", "access", "tcp", "--hostname", hostname, "--url", fmt.Sprintf("127.0.0.1:%d", port))
		t.l.Info("started tunnel", "addr", addr)
		if err := cmd.Run(); err != nil {
			t.l.Warn("failed to start tunnel", "error", err)
			time.Sleep(time.Second)
			continue
		}
		t.l.Info("done?", "addr", addr)
	}
}
