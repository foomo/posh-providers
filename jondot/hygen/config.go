package hygen

// Config holds the hygen provider's settings.
type Config struct {
	// Directory holding the template directories; its parent becomes HYGEN_TMPLS
	TemplatePath string `json:"templatePath" yaml:"templatePath"`
}
