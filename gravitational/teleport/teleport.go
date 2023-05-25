package teleport

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/env"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/shell"
	"github.com/pterm/pterm"
	"github.com/spf13/viper"
)

// Teleport command
type (
	Teleport struct {
		l            log.Logger
		cfg          Config
		cache        cache.Namespace
		configKey    string
		signedIn     bool
		signedInTime time.Time
	}
	Option func(*Teleport) error
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func CommandWithConfigKey(v string) Option {
	return func(o *Teleport) error {
		o.configKey = v
		return nil
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

// NewTeleport command
func NewTeleport(l log.Logger, cache cache.Cache, opts ...Option) (*Teleport, error) {
	inst := &Teleport{
		l:         l,
		cache:     cache.Get("teleport"),
		configKey: "teleport",
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

	if err := os.Setenv("TELEPORT_HOME", env.Path(inst.cfg.Path)); err != nil {
		return nil, err
	}

	return inst, nil
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (t *Teleport) Config() Config {
	return t.cfg
}

func (t *Teleport) IsAuthenticated(ctx context.Context) bool {
	if t.signedIn && time.Since(t.signedInTime) < 12*time.Hour {
		return true
	} else if _, err := shell.New(ctx, t.l, "tsh", "status").Quiet().Output(); err != nil {
		t.signedIn = false
		return false
	} else {
		t.signedIn = true
		t.signedInTime = time.Now()
		return true
	}
}

// Clusters returns a list of cluster
//
//nolint:forcetypeassert
func (t *Teleport) Clusters(ctx context.Context) []string {
	if !t.IsAuthenticated(ctx) {
		return nil
	}
	return t.cache.Get("clusters", func() interface{} {
		ret := []string{}

		type cluster struct {
			KubeClusterName string `json:"kube_cluster_name"`
		}

		value, err := shell.New(ctx, t.l, "tsh", "kube", "ls",
			fmt.Sprintf("--query='%s'", t.cfg.Query()),
			"--format", "json",
		).
			Output()
		if err != nil {
			pterm.Error.Println(err.Error())
			return ret
		}

		var clusters []cluster
		if err := json.Unmarshal(value, &clusters); err != nil {
			pterm.Error.Println(err.Error())
			return ret
		}

		for _, s := range clusters {
			ret = append(ret, s.KubeClusterName)
		}
		return ret
	}).([]string)
}

// Databases returns a list of cluster
//
//nolint:forcetypeassert
func (t *Teleport) Databases(ctx context.Context) []string {
	if !t.IsAuthenticated(ctx) {
		return nil
	}
	return t.cache.Get("databases", func() interface{} {
		ret := []string{}

		type (
			metadata struct {
				Name string `json:"name"`
			}
			db struct {
				Metadata metadata `json:"metadata"`
			}
		)
		value, err := shell.New(ctx, t.l, "tsh", "db", "ls",
			fmt.Sprintf("--query='%s'", t.cfg.Query()),
			"--format", "json",
		).
			Output()
		if err != nil {
			pterm.Error.Println(err.Error())
			return ret
		}

		var dbs []db
		if err := json.Unmarshal(value, &dbs); err != nil {
			pterm.Error.Println(err.Error())
			return ret
		}

		for _, s := range dbs {
			ret = append(ret, s.Metadata.Name)
		}

		return ret
	}).([]string)
}
