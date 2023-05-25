package etcd

import (
	"context"
	"fmt"
	"strings"

	"github.com/foomo/posh-providers/kubernets/kubectl"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/files"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type (
	ETCD struct {
		l         log.Logger
		cfg       Config
		configKey string
		kubectl   *kubectl.Kubectl
	}
	Option func(*ETCD) error
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func CommandWithConfigKey(v string) Option {
	return func(o *ETCD) error {
		o.configKey = v
		return nil
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

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
