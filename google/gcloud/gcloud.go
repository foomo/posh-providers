package gcloud

import (
	"context"
	"encoding/json"
	"os"
	"path"
	"regexp"

	"github.com/foomo/posh/pkg/env"
	"github.com/foomo/posh/pkg/shell"
	"github.com/pkg/errors"

	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/util/files"
	"github.com/spf13/viper"
)

type (
	GCloud struct {
		l                    log.Logger
		cfg                  *Config
		cache                cache.Namespace
		configKey            string
		accountFileNameRegex *regexp.Regexp
	}
	Option func(*GCloud)
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func WithConfig(v *Config) Option {
	return func(o *GCloud) {
		o.cfg = v
	}
}

func WithConfigKey(v string) Option {
	return func(o *GCloud) {
		o.configKey = v
	}
}

func WithAccountFileNameRegex(v *regexp.Regexp) Option {
	return func(o *GCloud) {
		o.accountFileNameRegex = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func New(l log.Logger, cache cache.Cache, opts ...Option) (*GCloud, error) {
	inst := &GCloud{
		l:                    l,
		cache:                cache.Get("gcloud"),
		configKey:            "gcloud",
		accountFileNameRegex: regexp.MustCompile(`(\w+)@(\w+)-(\w+)\.json`),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}

	if inst.cfg == nil && inst.configKey != "" {
		if err := viper.UnmarshalKey(inst.configKey, &inst.cfg); err != nil {
			return nil, err
		}
	}

	// ensure config path
	if err := files.MkdirAll(inst.cfg.ConfigPath); err != nil {
		return nil, errors.Wrapf(err, "failed to create directory %q", inst.cfg.ConfigPath)
	}

	if err := os.Setenv("CLOUDSDK_CONFIG", env.Path(inst.cfg.ConfigPath)); err != nil {
		return nil, err
	}

	return inst, nil
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (p *GCloud) Config() *Config {
	return p.cfg
}

func (p *GCloud) ServiceAccountKeysPath() string {
	return path.Join(p.cfg.ConfigPath, "service_account_keys")
}

// ActiveAccount returns the active account email
func (p *GCloud) ActiveAccount(ctx context.Context, l log.Logger) (string, error) {
	out, err := shell.New(ctx, l, "gcloud", "auth", "list", "--format", "json").Quiet().Output()
	if err != nil {
		return "", err
	}

	var list []struct {
		Account string `json:"account"`
		Status  string `json:"status"`
	}
	if err := json.Unmarshal(out, &list); err != nil {
		return "", err
	}

	for _, item := range list {
		if item.Status == "ACTIVE" {
			return item.Account, nil
		}
	}

	return "", errors.New("no active account")
}
