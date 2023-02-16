package etcd

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path"
	"strings"

	prompt2 "github.com/c-bata/go-prompt"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/util/prints"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/pkg/errors"
)

type Command struct {
	l           log.Logger
	etcd        *ETCD
	commandTree *tree.Root
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, etcd *ETCD, opts ...Option) *Command {
	inst := &Command{
		l:    l.Named("etcd"),
		etcd: etcd,
	}

	pathArg := &tree.Arg{
		Name: "path",
		Suggest: func(ctx context.Context, t *tree.Root, r *readline.Readline) []prompt2.Suggest {
			if value, ok := inst.etcd.cfg.Cluster(r.Args().At(0)); ok {
				return suggests.List(value.Paths)
			}
			return nil
		},
	}

	inst.commandTree = &tree.Root{
		Name:        "etcd",
		Description: "read and write to etcd",
		Nodes: tree.Nodes{
			{
				Name: "cluster",
				Values: func(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
					var ret []string
					for _, cluster := range inst.etcd.kubectl.Clusters() {
						if _, ok := inst.etcd.cfg.Cluster(cluster.Name()); ok {
							ret = append(ret, cluster.Name())
						}
					}
					return suggests.List(ret)
				},
				Nodes: tree.Nodes{
					{
						Name:    "get",
						Args:    tree.Args{pathArg},
						Execute: inst.get,
					},
					{
						Name:    "edit",
						Args:    tree.Args{pathArg},
						Execute: inst.edit,
					},
				},
			},
		},
	}

	return inst
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (c *Command) Name() string {
	return c.commandTree.Name
}

func (c *Command) Description() string {
	return c.commandTree.Description
}

func (c *Command) Complete(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
	return c.commandTree.Complete(ctx, r)
}

func (c *Command) Execute(ctx context.Context, r *readline.Readline) error {
	return c.commandTree.Execute(ctx, r)
}

func (c *Command) Help(ctx context.Context, r *readline.Readline) string {
	return `Read and write to etcd.

Usage:
  etcd [cluster] [cmd]

Available commands:
  get [path]   Prints the value
  edit [path]  Edit the given value

Examples:
  etcd example-cluster get config.yaml
  etcd example-cluster edit config.yaml
`
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) get(ctx context.Context, r *readline.Readline) error {
	etcdPath := r.Args().At(2)
	if cluster, ok := c.etcd.cfg.Cluster(r.Args().At(0)); !ok {
		return errors.New("invalid cluster")
	} else if out, err := c.etcd.GetPath(ctx, cluster, etcdPath); err != nil {
		return errors.Wrap(err, out)
	} else {
		prints.Code(c.l, etcdPath, out+"\n", "yaml")
	}
	return nil
}

func (c *Command) edit(ctx context.Context, r *readline.Readline) error {
	var (
		prev []byte
		next []byte
	)
	cluster, ok := c.etcd.cfg.Cluster(r.Args().At(0))
	if !ok {
		return errors.New("invalid cluster")
	}

	etcdPath := r.Args().At(2)
	filename := path.Join(os.Getenv("PROJECT_ROOT"), c.etcd.cfg.ConfigPath, etcdPath)

	{ // retrieve data
		if value, err := c.etcd.GetPath(ctx, cluster, etcdPath); err != nil {
			return err
		} else {
			prev = []byte(strings.ReplaceAll(value, "\r\r\n", "\n"))
		}
	}

	{ // write to file
		if err := os.MkdirAll(path.Dir(filename), 0700); err != nil {
			return err
		} else if err := os.WriteFile(filename, prev, 0600); err != nil {
			return err
		}
	}

	{ // edit file
		d := "vim"
		if value := os.Getenv("EDITOR"); value != "" {
			d = value
		}
		editor := exec.Command(d, filename)
		editor.Stdin = os.Stdin
		editor.Stdout = os.Stdout
		editor.Stderr = os.Stderr
		if err := editor.Run(); err != nil {
			return err
		}
	}

	{ // read in file
		if value, err := os.ReadFile(filename); err != nil {
			return err
		} else if bytes.Equal(prev, value) {
			c.l.Info("no changes")
			return nil
		} else {
			next = value
		}
	}

	c.l.Info("updating config")
	if out, err := c.etcd.SetPath(ctx, cluster, etcdPath, string(next)); err != nil {
		return errors.Wrap(err, out)
	}
	return nil
}
