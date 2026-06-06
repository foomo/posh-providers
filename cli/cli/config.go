package cli

type Config struct {
	// Required OAuth scopes per gh hostname.
	Scopes map[string][]string `json:"scopes" yaml:"scopes"`
}

func (c Config) RequiredScopes(host string) []string {
	return c.Scopes[host]
}
