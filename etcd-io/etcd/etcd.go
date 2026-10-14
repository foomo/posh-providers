package etcd

import (
	"context"
	"fmt"
	"strings"

	"github.com/foomo/posh-providers/kubernetes/kubectl"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/files"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type (
	// ETCD reads the provider's configuration and executes etcdctl inside a
	// configured cluster's etcd pod via kubectl.
	ETCD struct {
		l         log.Logger
		cfg       Config
		configKey string
		kubectl   *kubectl.Kubectl
	}
	// Option configures an ETCD built by New.
	Option func(*ETCD) error
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

// CommandWithConfigKey overrides the viper key the Config is unmarshalled
// from, which defaults to "etcd".
func CommandWithConfigKey(v string) Option {
	return func(o *ETCD) error {
		o.configKey = v
		return nil
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

// New reads the etcd Config from viper and creates its config directory. It
// returns an error if the config cannot be unmarshalled or the directory
// cannot be created.
func New(l log.Logger, kubectl *kubectl.Kubectl, opts ...Option) (*ETCD, error) {
	inst := &ETCD{
		l:         l,
		kubectl:   kubectl,
		configKey: "etcd",
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

	if err := files.MkdirAll(inst.cfg.ConfigPath); err != nil {
		return nil, errors.Wrapf(err, "failed to create config path: %s", inst.cfg.ConfigPath)
	}

	return inst, nil
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

// GetPath execs into the cluster's etcd pod and returns the value stored at
// path, via `etcdctl get --print-value-only`.
func (c *ETCD) GetPath(ctx context.Context, cluster Cluster, profile, path string) (string, error) {
	if out, err := shell.New(ctx, c.l, "kubectl", "exec",
		"-it", cluster.PodName,
		"--namespace", cluster.Namespace,
		"--", "/bin/sh", "-c", "'etcdctl get "+path+" --print-value-only'",
	).
		Env(c.kubectl.Cluster(cluster.Name).Env(profile)).
		Output(); err != nil {
		return "", err
	} else {
		return strings.TrimSpace(string(out)), nil
	}
}

// SetPath execs into the cluster's etcd pod and writes value to path via
// `etcdctl put`. value is interpolated into the shell command unescaped, so
// values containing quotes or backticks will corrupt the command.
func (c *ETCD) SetPath(ctx context.Context, cluster Cluster, profile, path, value string) (string, error) {
	if out, err := shell.New(ctx, c.l, "kubectl", "exec",
		"-it", cluster.PodName,
		"--namespace", cluster.Namespace,
		"--", "/bin/sh", "-c", fmt.Sprintf("'echo \"%s\" | etcdctl put "+path+"'", strings.ReplaceAll(value, "\n", "\\n")),
	).
		Env(c.kubectl.Cluster(cluster.Name).Env(profile)).
		Output(); err != nil {
		return "", err
	} else {
		return strings.TrimSpace(string(out)), nil
	}
}
