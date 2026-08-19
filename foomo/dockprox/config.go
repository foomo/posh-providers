package dockprox

type Config struct {
	// Path to the dockprox config file, passed to `dockprox serve --config`.
	// Used verbatim, without environment expansion, and it is also what makes a
	// started process findable again by stop.
	Config string `json:"config" yaml:"config"`
}
