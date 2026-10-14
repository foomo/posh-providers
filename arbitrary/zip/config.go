package zip

import (
	"github.com/foomo/posh-providers/onepassword"
	"github.com/samber/lo"
)

// Config holds the named 1Password secrets usable as archive passwords.
type Config struct {
	// Named 1Password secrets holding archive passwords, keyed by the name accepted by --cred
	Credentials map[string]onepassword.Secret `json:"credentials" yaml:"credentials"`
}

// Credential returns the named credential's 1Password secret and whether it exists.
func (c Config) Credential(name string) (onepassword.Secret, bool) {
	value, ok := c.Credentials[name]
	return value, ok
}

// CredentialNames returns the configured credential names.
func (c Config) CredentialNames() []string {
	return lo.Keys(c.Credentials)
}
