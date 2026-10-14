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
	// Zip creates and extracts zip archives, optionally password-protected
	// using a secret fetched from 1Password.
	Zip struct {
		l         log.Logger
		cfg       Config
		configKey string
		op        *onepassword.OnePassword
	}
	// Option configures a Zip in New.
	Option func(*Zip) error
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

// WithConfigKey sets the viper key the Config is unmarshalled from,
// overriding the default "zip".
func WithConfigKey(v string) Option {
	return func(o *Zip) error {
		o.configKey = v
		return nil
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

// New builds a Zip, unmarshalling its Config from the configured viper key.
// It returns an error if that unmarshal fails or an Option fails.
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

// Config returns the loaded configuration.
func (c *Zip) Config() Config {
	return c.cfg
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

// Create zips filename's basename into "<basename>.zip" alongside it.
func (c *Zip) Create(ctx context.Context, filename string) error {
	basename := path.Base(filename)
	if out, err := shell.New(ctx, c.l, "zip", basename+".zip", basename).
		Dir(path.Dir(filename)).
		Output(); err != nil {
		return errors.Wrap(err, string(out))
	}

	return nil
}

// Extract unzips filename into the directory containing it.
func (c *Zip) Extract(ctx context.Context, filename string) error {
	basename := path.Base(filename)
	if out, err := shell.New(ctx, c.l, "unzip", basename).
		Dir(path.Dir(filename)).
		Output(); err != nil {
		return errors.Wrap(err, string(out))
	}

	return nil
}

// CreateWithPassword zips filename's basename into "<basename>.zip" alongside
// it, encrypted with the named credential's password from 1Password.
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

// ExtractWithPassword unzips filename into the directory containing it,
// decrypting it with the named credential's password from 1Password.
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

// GetCredentialPassword fetches the named credential's password from
// 1Password. It returns an error if the credential is not configured.
func (c *Zip) GetCredentialPassword(ctx context.Context, name string) (string, error) {
	secret, ok := c.cfg.Credential(name)
	if !ok {
		return "", fmt.Errorf("credential %s not found", name)
	}

	return c.op.Get(ctx, secret)
}
