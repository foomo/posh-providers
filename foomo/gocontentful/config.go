package gocontentful

type Config struct {
	SpaceID      string   `yaml:"spaceId"`
	CMAKey       string   `yaml:"cmaKey"`
	ContentTypes []string `yaml:"contentTypes"`
}
