package zip

import (
	"context"
	"fmt"
	"path"

	"github.com/foomo/posh-providers/onepassword"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/shell"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type (
	Zip struct {
		l         log.Logger
		cfg       Config
		configKey string
		op        *onepassword.OnePassword
	}
	Option func(*Zip) error
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func WithConfigKey(v string) Option {
	return func(o *Zip) error {
		o.configKey = v
		return nil
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func New(l log.Logger, op *onepassword.OnePassword, opts ...Option) (*Zip, error) {
	inst := &Zip{
		l:         l,
		op:        op,
		configKey: "zip",
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

	return inst, nil
}

// ------------------------------------------------------------------------------------------------
// ~ Getter
// ------------------------------------------------------------------------------------------------

func (c *Zip) Config() Config {
	return c.cfg
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (c *Zip) Create(ctx context.Context, filename string) error {
	basename := path.Base(filename)
	if out, err := shell.New(ctx, c.l, "zip", basename+".zip", basename).
		Dir(path.Dir(filename)).
		Output(); err != nil {
		return errors.Wrap(err, string(out))
	}
	return nil
}

func (c *Zip) Extract(ctx context.Context, filename string) error {
	basename := path.Base(filename)
	if out, err := shell.New(ctx, c.l, "unzip", basename).
		Dir(path.Dir(filename)).
		Output(); err != nil {
		return errors.Wrap(err, string(out))
	}
	return nil
}

func (c *Zip) CreateWithPassword(ctx context.Context, filename, credential string) error {
	basename := path.Base(filename)

	password, err := c.GetCredentialPassword(ctx, credential)
	if err != nil {
		return err
	}

	if out, err := shell.New(ctx, c.l, "zip", "-e", "-P", password, basename+".zip", basename).
		Dir(path.Dir(filename)).
		Output(); err != nil {
		return errors.Wrap(err, string(out))
	}
	return nil
}

func (c *Zip) ExtractWithPassword(ctx context.Context, filename, credential string) error {
	basename := path.Base(filename)

	password, err := c.GetCredentialPassword(ctx, credential)
	if err != nil {
		return err
	}

	if out, err := shell.New(ctx, c.l, "unzip", "-P", password, basename).
		Dir(path.Dir(filename)).
		Output(); err != nil {
		return errors.Wrap(err, string(out))
	}
	return nil
}

func (c *Zip) GetCredentialPassword(ctx context.Context, name string) (string, error) {
	secret, ok := c.cfg.Credential(name)
	if !ok {
		return "", fmt.Errorf("credential %s not found", name)
	}
	return c.op.Get(ctx, secret)
}
