package tempo

type Profile struct {
	// URL is the Temporal server address (host:port), passed to the CLI via --address.
	URL string `json:"url" yaml:"url"`
	// Description is shown in shell autocompletion.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}
