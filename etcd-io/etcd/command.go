package etcd

import (
	"bytes"
	"context"
	_ "embed"
	"os"
	"os/exec"
	"path"
	"slices"
	"strings"

	prompt2 "github.com/c-bata/go-prompt"
	"github.com/foomo/posh-providers/kubernetes/kubectl"
	"github.com/foomo/posh/pkg/command"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/env"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/util/prints"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/pkg/errors"
)

//go:embed SKILL.md
var skill string

type Command struct {
	l           log.Logger
	etcd        *ETCD
	kubectl     *kubectl.Kubectl
	commandTree tree.Root
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, etcd *ETCD, kubectl *kubectl.Kubectl, opts ...Option) *Command {
	inst := &Command{
		l:       l.Named("etcd"),
		etcd:    etcd,
		kubectl: kubectl,
	}

	args := tree.Args{
		{
			Name:        "path",
			Description: "etcd key to read or write, from the cluster's configured paths",
			Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []prompt2.Suggest {
				if value, ok := inst.etcd.cfg.Cluster(r.Args().At(0)); ok {
					return suggests.List(value.Paths)
				}

				return nil
			},
		},
	}
	flags := func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
		if r.Args().HasIndex(0) {
			fs.Internal().String("profile", "", "Kubectl config profile holding this cluster's kubeconfig")

			if err := fs.Internal().SetValues("profile", inst.kubectl.Cluster(r.Args().At(0)).Profiles(ctx)...); err != nil {
				return err
			}
		}

		return nil
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        "etcd",
		Description: "Read and write etcd keys inside a cluster's etcd pod",
		Nodes: tree.Nodes{
			{
				Name:        "cluster",
				Description: "Cluster to connect to; must be configured here and have a kubeconfig",
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
						Name:        "get",
						Description: "Print the value at the given key",
						Args:        args,
						Flags:       flags,
						Execute:     inst.get,
					},
					{
						Name:        "edit",
						Description: "Open the value in $EDITOR and write changes back to etcd",
						Args:        args,
						Flags:       flags,
						Execute:     inst.edit,
					},
				},
			},
		},
	})

	return inst
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (c *Command) Name() string {
	return c.commandTree.Node().Name
}

func (c *Command) Description() string {
	return c.commandTree.Node().Description
}

func (c *Command) Complete(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
	return c.commandTree.Complete(ctx, r)
}

func (c *Command) Execute(ctx context.Context, r *readline.Readline) error {
	return c.commandTree.Execute(ctx, r)
}

func (c *Command) Help(ctx context.Context, r *readline.Readline) string {
	return c.commandTree.Help(ctx, r)
}

// Describe implements the optional command.Describer interface, letting
// `posh agent catalog` describe this command's subtree.
func (c *Command) Describe(ctx context.Context) command.CommandInfo {
	return c.commandTree.Describe(ctx)
}

// Skill implements the optional command.Skiller interface. The catalog cannot
// show that `edit` writes straight into a live cluster's etcd behind an
// interactive editor, that the value is interpolated into a shell command inside
// the pod and so corrupts anything containing quotes or backticks, or that both
// verbs depend on an exact pod name that changes when the pod is rescheduled.
func (c *Command) Skill(ctx context.Context) string {
	return skill
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) get(ctx context.Context, r *readline.Readline) error {
	etcdPath := r.Args().At(2)
	ifs := r.FlagSets().Internal()

	profile, err := ifs.GetString("profile")
	if err != nil {
		return err
	}

	if cluster, ok := c.etcd.cfg.Cluster(r.Args().At(0)); !ok {
		return errors.New("invalid cluster")
	} else if out, err := c.etcd.GetPath(ctx, cluster, profile, etcdPath); err != nil {
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
	ifs := r.FlagSets().Internal()
	filename := env.Path(c.etcd.cfg.ConfigPath, etcdPath)

	profile, err := ifs.GetString("profile")
	if err != nil {
		return err
	}

	{ // retrieve data
		if value, err := c.etcd.GetPath(ctx, cluster, profile, etcdPath); err != nil {
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

		if value := os.Getenv("EDITOR"); slices.Contains([]string{"vim", "nvim", "nano", "code", "micro", "emacs"}, value) {
			d = value
		}

		editor := exec.CommandContext(ctx, d, filename) //nolint:gosec // checked above
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

	if out, err := c.etcd.SetPath(ctx, cluster, profile, etcdPath, string(next)); err != nil {
		return errors.Wrap(err, out)
	}

	return nil
}
