package zip

import (
	"github.com/foomo/posh-providers/onepassword"
	"github.com/samber/lo"
)

type Config struct {
	Credentials map[string]onepassword.Secret `json:"credentials" yaml:"credentials"`
}

func (c Config) Credential(name string) (onepassword.Secret, bool) {
	value, ok := c.Credentials[name]
	return value, ok
}

func (c Config) CredentialNames() []string {
	return lo.Keys(c.Credentials)
}
