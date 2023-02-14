package kubectl

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/foomo/posh/pkg/env"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/shell"
)

const None string = "_"

type Cluster struct {
	l       log.Logger
	name    string
	kubectl *Kubectl
}

type Clusters []*Cluster

func NewCluster(kubectl *Kubectl, name string) *Cluster {
	return &Cluster{
		l:       kubectl.l.Named("kubectl:" + name),
		name:    name,
		kubectl: kubectl,
	}
}

func (c *Cluster) IsNone() bool {
	return c.name == None
}

func (c *Cluster) Name() string {
	return c.name
}

func (c *Cluster) String() string {
	return c.name
}

func (c *Cluster) Env() string {
	return fmt.Sprintf("KUBECONFIG=%s", c.Config())
}

func (c *Cluster) Config() string {
	return path.Join(os.Getenv(env.ProjectRoot), c.kubectl.cfg.ConfigPath, c.Name()+".yaml")
}

func (c *Cluster) ConfigExists() bool {
	if _, err := os.Stat(c.Config()); err == nil {
		return true
	}
	return false
}

func (c *Cluster) DeleteConfig() error {
	if !c.ConfigExists() {
		return nil
	}
	return os.Remove(c.Config())
}

//nolint:forcetypeassert
func (c *Cluster) Namespaces(ctx context.Context) []string {
	return c.kubectl.cache.Get(c.name+"-namespaces", func() any {
		if sh, err := c.shell(ctx,
			"get", "namespaces",
			"-o", "jsonpath='{.items[*].metadata.name}'",
		); err != nil {
			c.l.Debug(err.Error())
			return []string{}
		} else if out, err := sh.Output(); err != nil {
			c.l.Debug(err.Error())
			return []string{}
		} else {
			return strings.Split(string(out), " ")
		}
	}).([]string)
}

//nolint:forcetypeassert
func (c *Cluster) Pods(ctx context.Context, namespace string) []string {
	return c.kubectl.cache.Get(c.name+"-"+namespace+"-pods", func() any {
		if sh, err := c.shell(ctx,
			"get", "pods",
			"-n", namespace,
			"-o", "jsonpath='{.items[*].metadata.name}'",
		); err != nil {
			c.l.Debug(err.Error())
			return []string{}
		} else if out, err := sh.Output(); err != nil {
			c.l.Warn(err.Error())
			return []string{}
		} else {
			return strings.Split(string(out), " ")
		}

	}).([]string)
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Cluster) shell(ctx context.Context, args ...string) (*shell.Shell, error) {
	sh := shell.New(ctx, c.l, "kubectl").Args(args...)
	if c.kubectl.authTokenProvider != nil {
		if token, err := c.kubectl.authTokenProvider(ctx, c.Name()); err != nil {
			return nil, err
		} else {
			sh.Args("--token", token)
		}
	}
	return sh.Env(c.Env()), nil
}
