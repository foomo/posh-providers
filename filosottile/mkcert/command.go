package mkcert

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/foomo/posh/pkg/command"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/files"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

//go:embed SKILL.md
var skill string

// skillName is the placeholder the embedded SKILL.md uses wherever the command's
// own name appears. Skill substitutes the name the command is registered under,
// which is not necessarily the default: a fragment hardcoding the default tells
// an agent to run a command the project may not have.
const skillName = "{{cmd}}"

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
				Description: "Generate every configured certificate, overwriting existing files",
				Execute:     inst.generate,
			},
			{
				Name:        "create",
				Description: "Create a new certificate for the given names into the certificate path",
				Args: []*tree.Arg{
					{
						Name:        "names",
						Description: "Hostnames, wildcards, IPs, URLs or emails to include in the certificate",
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

// Describe implements the optional command.Describer interface, letting
// `posh agent catalog` describe this command's subtree.
func (c *Command) Describe(ctx context.Context) command.CommandInfo {
	return c.commandTree.Describe(ctx)
}

// Skill implements the optional command.Skiller interface. The catalog cannot
// show that `install` touches the system trust store rather than project state;
// that `generate` and `create` overwrite unencrypted keys without prompting; or
// that the two verbs derive output filenames differently for the same
// certificate.
func (c *Command) Skill(ctx context.Context, name string) string {
	return strings.ReplaceAll(skill, skillName, name)
}

// SkillMetadata implements the optional command.SkillMetadataer interface,
// supplying the frontmatter of this command's generated skill. The description
// names the browser-level TLS symptoms, since that is what sends someone looking
// for a local certificate rather than the name of the tool that mints one.
func (c *Command) SkillMetadata(ctx context.Context, name string) command.SkillMetadata {
	return command.SkillMetadata{
		Description: "Use when setting up HTTPS for local development - issuing a trusted " +
			"certificate for a local hostname or wildcard domain, regenerating this " +
			"project's certificates, or installing and removing the local CA from the " +
			"system trust store. Also when a local site shows a certificate warning, " +
			"ERR_CERT_AUTHORITY_INVALID, or an x509 unknown-authority error.",
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) install(ctx context.Context, r *readline.Readline) error {
	return shell.New(ctx, c.l, "mkcert", "-install").
		Args(r.Args().From(1)...).
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
		Args(r.Args().From(1)...).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Run()
}

func (c *Command) uninstall(ctx context.Context, r *readline.Readline) error {
	return shell.New(ctx, c.l, "mkcert", "-uninstall").
		Args(r.Args().From(1)...).
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
