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
		sh := shell.New(ctx, c.l, "kubectl",
			"get", "namespaces",
			"-o", "jsonpath='{.items[*].metadata.name}'").
			Env(c.Env())

		if c.kubectl.authProvider != nil {
			err := c.kubectl.authProvider(ctx, c.Name(), sh)
			if err != nil {
				c.l.Warn(err.Error())
				return []string{}
			}
		}

		out, err := sh.Output()
		if err != nil {
			c.l.Warn(err.Error())
			return []string{}
		}

		return strings.Split(string(out), " ")
	}).([]string)
}

//nolint:forcetypeassert
func (c *Cluster) Pods(ctx context.Context, namespace string) []string {
	return c.kubectl.cache.Get(c.name+"-"+namespace+"-pods", func() any {
		if out, err := shell.New(ctx, c.l, "kubectl",
			"get", "pods",
			"-n", namespace,
			"-o", "jsonpath='{.items[*].metadata.name}'",
		).
			Env(c.Env()).
			Output(); err != nil {
			c.l.Warn(err.Error())
			return []string{}
		} else {
			return strings.Split(string(out), " ")
		}
	}).([]string)
}
