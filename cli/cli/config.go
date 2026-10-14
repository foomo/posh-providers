package cli

// Config holds the required gh OAuth scopes per hostname, checked by
// ScopesChecker.
type Config struct {
	// Required OAuth scopes per gh hostname.
	Scopes map[string][]string `json:"scopes" yaml:"scopes"`
}

// RequiredScopes returns the configured OAuth scopes required for host.
func (c Config) RequiredScopes(host string) []string {
	return c.Scopes[host]
}
