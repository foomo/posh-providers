package gnat

type Profile struct {
	// URL is the NATS server URL, passed to the CLI via -url.
	URL string `json:"url" yaml:"url"`
	// Description is shown in shell autocompletion.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}
