package gocontentful

type Config struct {
	SpaceID      string   `yaml:"spaceId"`
	CMAKey       string   `yaml:"cmaKey"`
	Environment  string   `yaml:"environment,omitempty"`
	ContentTypes []string `yaml:"contentTypes"`
}
