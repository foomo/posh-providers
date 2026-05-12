package gokazi

type Config struct {
	// Cleanup will stop all processes if last posh is closed
	Cleanup bool `json:"cleanup" yaml:"cleanup"`
}
