package rclone

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/env"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/spf13/viper"
)

type (
	Command struct {
		l           log.Logger
		cfg         Config
		name        string
		cache       cache.Namespace
		configKey   string
		commandTree tree.Root
	}
	StackNameProvider func(path string) string
	CommandOption     func(*Command)
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func CommandWithName(v string) CommandOption {
	return func(o *Command) {
		o.name = v
	}
}

func WithConfigKey(v string) CommandOption {
	return func(o *Command) {
		o.configKey = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, cache cache.Cache, opts ...CommandOption) (*Command, error) {
	inst := &Command{
		l:         l.Named("rclone"),
		name:      "rclone",
		cache:     cache.Get("rclone"),
		configKey: "rclone",
	}
	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}
	if err := viper.UnmarshalKey(inst.configKey, &inst.cfg); err != nil {
		return nil, err
	}

	if err := os.Setenv("RCLONE_CONFIG", env.Path(inst.cfg.Path)); err != nil {
		return nil, err
	}

	remoteArg := &tree.Arg{
		Name:        "remote",
		Description: "Configure remote",
		Repeat:      true,
		Optional:    true,
		Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
			return inst.cache.GetSuggests("remotes", func() any {
				out, _ := exec.CommandContext(ctx, "rclone", "listremotes").Output()
				ret := strings.Split(strings.Trim(string(out), "\n"), "\n")
				return suggests.List(ret)
			})
		},
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Run rclone commands",
		Nodes: tree.Nodes{
			{
				Name:        "init",
				Description: "Generate config file",
				Execute:     inst.init,
			},
			{
				Name:        "cat",
				Description: "Concatenates any files and sends them to stdout",
				Args:        tree.Args{remoteArg},
				Execute:     inst.execute,
			},
			{
				Name:        "check",
				Description: "Checks the files in the source and destination match",
				Args:        tree.Args{remoteArg},
				Execute:     inst.execute,
			},
			{
				Name:        "checksum",
				Description: "Checks the files in the destination against a SUM file",
				Args:        tree.Args{remoteArg},
				Execute:     inst.execute,
			},
			{
				Name:        "copy",
				Description: "Copy files from source to dest, skipping identical files",
				Args:        tree.Args{remoteArg},
				Execute:     inst.execute,
			},
			{
				Name:        "copyto",
				Description: "Copy files from source to dest, skipping identical files",
				Args:        tree.Args{remoteArg},
				Execute:     inst.execute,
			},
			{
				Name:        "copyurl",
				Description: "Copy the contents of the URL supplied content to dest:path",
				Args:        tree.Args{remoteArg},
				Execute:     inst.execute,
			},
			{
				Name:        "dedupe",
				Description: "Interactively find duplicate filenames and delete/rename them",
				Args:        tree.Args{remoteArg},
				Execute:     inst.execute,
			},
			{
				Name:        "delete",
				Description: "Remove the files in path",
				Args:        tree.Args{remoteArg},
				Execute:     inst.execute,
			},
			{
				Name:        "deletefile",
				Description: "Remove a single file from remote",
				Args:        tree.Args{remoteArg},
				Execute:     inst.execute,
			},
			{
				Name:        "ls",
				Description: "List the objects in the path with size and path",
				Args:        tree.Args{remoteArg},
				Execute:     inst.execute,
			},
			{
				Name:        "lsd",
				Description: "List all directories/containers/buckets in the path",
				Args:        tree.Args{remoteArg},
				Execute:     inst.execute,
			},
			{
				Name:        "lsf",
				Description: "List directories and objects in remote:path formatted for parsing",
				Args:        tree.Args{remoteArg},
				Execute:     inst.execute,
			},
			{
				Name:        "lsl",
				Description: "List the objects in path with modification time, size and path",
				Args:        tree.Args{remoteArg},
				Execute:     inst.execute,
			},
			{
				Name:        "mkdir",
				Description: "Make the path if it doesn't already exist",
				Args:        tree.Args{remoteArg},
				Execute:     inst.execute,
			},
			{
				Name:        "mount",
				Description: "Mount the remote as file system on a mountpoint",
				Args:        tree.Args{remoteArg},
				Execute:     inst.execute,
			},
			{
				Name:        "move",
				Description: "Move files from source to dest",
				Args:        tree.Args{remoteArg},
				Execute:     inst.execute,
			},
			{
				Name:        "moveto",
				Description: "Move file or directory from source to dest",
				Args:        tree.Args{remoteArg},
				Execute:     inst.execute,
			},
			{
				Name:        "nfsmount",
				Description: "Mount the remote as file system on a mountpoint",
				Args:        tree.Args{remoteArg},
				Execute:     inst.execute,
			},
			{
				Name:        "purge",
				Description: "Remove the path and all of its contents",
				Args:        tree.Args{remoteArg},
				Execute:     inst.execute,
			},
			{
				Name:        "rmdir",
				Description: "Remove the empty directory at path",
				Args:        tree.Args{remoteArg},
				Execute:     inst.execute,
			},
			{
				Name:        "rmdirs",
				Description: "Remove empty directories under the path",
				Args:        tree.Args{remoteArg},
				Execute:     inst.execute,
			},
			{
				Name:        "size",
				Description: "Prints the total size and number of objects in remote:path",
				Args:        tree.Args{remoteArg},
				Execute:     inst.execute,
			},
			{
				Name:        "sync",
				Description: "Make source and dest identical, modifying destination only",
				Args:        tree.Args{remoteArg},
				Execute:     inst.execute,
			},
			{
				Name:        "touch",
				Description: "Create new file or change file modification time",
				Args:        tree.Args{remoteArg},
				Execute:     inst.execute,
			},
			{
				Name:        "tree",
				Description: "List the contents of the remote in a tree like fashion",
				Args:        tree.Args{remoteArg},
				Execute:     inst.execute,
			},
		},
		Execute: inst.execute,
	})

	return inst, nil
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

func (c *Command) Validate(ctx context.Context, r *readline.Readline) error {
	if _, err := exec.LookPath("rclone"); err != nil {
		c.l.Print()
		return errors.New("missing rclone executable")
	}
	return nil
}

func (c *Command) Help(ctx context.Context, r *readline.Readline) string {
	return c.commandTree.Help(ctx, r)
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) init(ctx context.Context, r *readline.Readline) error {
	config, err := c.cfg.RenderConfig(ctx)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(path.Dir(env.Path(c.cfg.Path)), 0700); err != nil {
		return err
	}

	return os.WriteFile(env.Path(c.cfg.Path), config, 0600)
}

func (c *Command) execute(ctx context.Context, r *readline.Readline) error {
	return shell.New(ctx, c.l, "rclone").
		Args(r.Args()...).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Run()
}
