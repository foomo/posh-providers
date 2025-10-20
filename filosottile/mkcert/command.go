package mkcert

import (
	"context"
	"fmt"

	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/files"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type (
	Command struct {
		l           log.Logger
		cfg         Config
		name        string
		configKey   string
		commandTree tree.Root
	}
	Option func(*Command) error
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func CommandWithName(v string) Option {
	return func(o *Command) error {
		o.name = v
		return nil
	}
}

func WithConfigKey(v string) Option {
	return func(o *Command) error {
		o.configKey = v
		return nil
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, opts ...Option) (*Command, error) {
	inst := &Command{
		l:         l.Named("mkcert"),
		name:      "mkcert",
		configKey: "mkcert",
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

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Run mkcert",
		Nodes: tree.Nodes{
			{
				Name:        "install",
				Description: "Install the local CA in the system trust store",
				Execute:     inst.install,
			},
			{
				Name:        "caroot",
				Description: "Print the CA certificate and key storage location",
				Execute:     inst.caroot,
			},
			{
				Name:        "uninstall",
				Description: "Uninstall the local CA (but do not delete it)",
				Execute:     inst.uninstall,
			},
			{
				Name:        "generate",
				Description: "Generate configured certificates",
				Execute:     inst.generate,
			},
			{
				Name:        "create",
				Description: "Creat a new certificate for the given names",
				Args: []*tree.Arg{
					{
						Name:        "names",
						Description: "List of names including wildcard",
						Repeat:      true,
					},
				},
				Execute: inst.create,
			},
		},
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

func (c *Command) Validate(ctx context.Context, r *readline.Readline) error {
	switch {
	case r.Args().LenIs(0):
		return errors.New("missing [CMD] argument")
	}

	return nil
}

func (c *Command) Execute(ctx context.Context, r *readline.Readline) error {
	return c.commandTree.Execute(ctx, r)
}

func (c *Command) Help(ctx context.Context, r *readline.Readline) string {
	return c.commandTree.Help(ctx, r)
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) install(ctx context.Context, r *readline.Readline) error {
	return shell.New(ctx, c.l, "mkcert", "-install").
		Args(r.Args()...).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Run()
}

func (c *Command) generate(ctx context.Context, r *readline.Readline) error {
	if err := files.MkdirAll(c.cfg.CertificatePath); err != nil {
		return err
	}

	c.l.Info("Generating certificates:")

	for _, certificate := range c.cfg.Certificates {
		c.l.Info("└  " + certificate.Name)

		if err := shell.New(ctx, c.l, "mkcert").
			Args("-key-file", fmt.Sprintf("%s-key.pem", certificate.Name)).
			Args("-cert-file", fmt.Sprintf("%s.pem", certificate.Name)).
			Dir(c.cfg.CertificatePath).
			Args(certificate.Names...).
			Args(r.Flags()...).
			Args(r.AdditionalArgs()...).
			Run(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Command) caroot(ctx context.Context, r *readline.Readline) error {
	return shell.New(ctx, c.l, "mkcert", "-CAROOT").
		Args(r.Args()...).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Run()
}

func (c *Command) uninstall(ctx context.Context, r *readline.Readline) error {
	return shell.New(ctx, c.l, "mkcert", "-uninstall").
		Args(r.Args()...).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Run()
}

func (c *Command) create(ctx context.Context, r *readline.Readline) error {
	if err := files.MkdirAll(c.cfg.CertificatePath); err != nil {
		return err
	}

	return shell.New(ctx, c.l, "mkcert").
		Dir(c.cfg.CertificatePath).
		Args(r.Args().From(1)...).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Run()
}
