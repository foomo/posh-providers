package hygen

type Config struct {
	// Directory holding the template directories; its parent becomes HYGEN_TMPLS
	TemplatePath string `json:"templatePath" yaml:"templatePath"`
}
