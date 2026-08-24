package kubectl

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/foomo/posh/pkg/env"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/readline"
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

func (c *Cluster) Env(profile string) string {
	return fmt.Sprintf("KUBECONFIG=%s", c.Config(profile))
}

func (c *Cluster) Config(profile string) string {
	if profile != "" {
		return env.Path(c.kubectl.cfg.ConfigPath, profile, c.Name()+".yaml")
	}

	return env.Path(c.kubectl.cfg.ConfigPath, c.Name()+".yaml")
}

func (c *Cluster) ConfigExists(profile string) bool {
	if _, err := os.Stat(c.Config(profile)); err == nil {
		return true
	}

	return false
}

// ConfigExistsForFlags reports whether the cluster's kubeconfig exists, taking
// the profile from the raw --profile flag token rather than from a parsed flag
// set.
//
// It exists for Validate, which posh calls before Execute - and flag sets are
// only built inside Execute, so r.FlagSets() is nil at validation time and
// GetString("profile") is unavailable. The raw tokens are in r.Flags() by then,
// so the profile can still be honoured. Without this, a Validate can only check
// the profile-less path and will reject a cluster that exists only under a
// profile.
func (c *Cluster) ConfigExistsForFlags(flags readline.Args) bool {
	return c.ConfigExists(Profile(flags))
}

// Profile returns the value of the --profile flag in a raw, unparsed flag token
// list, or "" when it is absent. It accepts both "--profile x" and
// "--profile=x".
func Profile(flags readline.Args) string {
	for i, flag := range flags {
		if value, ok := strings.CutPrefix(flag, "--profile="); ok {
			return value
		}

		if flag == "--profile" && len(flags) > i+1 {
			return flags[i+1]
		}
	}

	return ""
}

func (c *Cluster) DeleteConfig(profile string) error {
	if !c.ConfigExists(profile) {
		return nil
	}

	return os.Remove(c.Config(profile))
}

//nolint:forcetypeassert
func (c *Cluster) Profiles(ctx context.Context) []string {
	return c.kubectl.cache.Get("profiles", func() any {
		files, err := os.ReadDir(c.kubectl.cfg.ConfigPath)
		if err != nil {
			c.l.Debug(err.Error())
			return []string{}
		}

		ret := []string{}

		for _, f := range files {
			if f.IsDir() && !strings.HasPrefix(f.Name(), ".") {
				ret = append(ret, f.Name())
			}
		}

		return ret
	}).([]string)
}

//nolint:forcetypeassert
func (c *Cluster) Namespaces(ctx context.Context, profile string) []string {
	return c.kubectl.cache.Get(profile+"-"+c.name+"-namespaces", func() any {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if sh, err := c.shell(ctx, profile,
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
func (c *Cluster) Pods(ctx context.Context, profile, namespace string) []string {
	return c.kubectl.cache.Get(profile+"-"+c.name+"-"+namespace+"-pods", func() any {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if sh, err := c.shell(ctx, profile,
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

func (c *Cluster) shell(ctx context.Context, profile string, args ...string) (*shell.Shell, error) {
	sh := shell.New(ctx, c.l, "kubectl").Args(args...)
	if c.kubectl.authTokenProvider != nil {
		if token, err := c.kubectl.authTokenProvider(ctx, c.Name()); err != nil {
			return nil, err
		} else {
			sh.Args("--token", token)
		}
	}

	return sh.Env(c.Env(profile)), nil
}
