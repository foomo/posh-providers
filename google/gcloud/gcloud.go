package gcloud

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/foomo/posh/pkg/shell"

	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/util/files"
	"github.com/pkg/errors"
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

	if err := files.MkdirAll(inst.cfg.ConfigDir); err != nil {
		return nil, errors.Wrapf(err, "failed to create directory %q", inst.cfg.ConfigDir)
	}

	return inst, nil
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (gc *GCloud) ParseAccounts(ctx context.Context) ([]Account, error) {
	accountFiles, err := files.Find(ctx, gc.cfg.ConfigDir, "*.json")
	if err != nil {
		return nil, err
	}

	var accounts []Account
	for _, f := range accountFiles {
		matchString := gc.accountFileNameRegex.FindAllStringSubmatch(filepath.Base(f), 1)
		if len(matchString) == 0 {
			continue
		}
		match := matchString[0]
		acc := Account{
			Role:        match[1],
			Environment: match[2],
			Cluster:     match[3],
			Path:        f,
		}
		accounts = append(accounts, acc)
	}

	return accounts, err
}

func (gc *GCloud) FindAccounts(ctx context.Context, env, cluster string) ([]Account, error) {
	accounts, err := gc.ParseAccounts(ctx)
	if err != nil {
		return nil, err
	}

	filtered := accounts[:0]
	for _, acc := range accounts {
		if acc.Environment == env && acc.Cluster == cluster {
			filtered = append(filtered, acc)
		}
	}

	if len(filtered) == 0 {
		return nil, fmt.Errorf("account not found for cluster %q and env %q", cluster, env)
	}

	return filtered, nil
}

func (gc *GCloud) GenerateToken(ctx context.Context, env, cluster string) (string, error) {
	accounts, err := gc.FindAccounts(ctx, env, cluster)
	if err != nil {
		return "", err
	}
	if len(accounts) > 1 {
		gc.l.Warnf("multiple accounts found for env %q and cluster %q", env, cluster)
	}
	account := accounts[0]

	out, err := shell.New(ctx, gc.l,
		"gcloud", "auth", "application-default", "print-access-token").
		Env("CLOUDSDK_AUTH_CREDENTIAL_FILE_OVERRIDE=" + account.Path).
		Env("GOOGLE_APPLICATION_CREDENTIALS=" + account.Path).
		Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
