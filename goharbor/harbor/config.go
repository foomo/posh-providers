package harbor

type Config struct {
	URL     string `json:"url" yaml:"url"`
	AuthURL string `json:"authUrl" yaml:"authUrl"`
}
