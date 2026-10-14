package gokazi

// Config holds the gokazi provider's configuration.
type Config struct {
	// Cleanup will stop all processes if last posh is closed
	Cleanup bool `json:"cleanup" yaml:"cleanup"`
}
