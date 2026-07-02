package gocontentful

type Config struct {
	SpaceID      string   `json:"spaceId" yaml:"spaceId"`
	CMAKey       string   `json:"cmaKey" yaml:"cmaKey"`
	Environment  string   `json:"environment,omitempty" yaml:"environment,omitempty"`
	Region       string   `json:"region,omitempty" yaml:"region,omitempty"`
	ContentTypes []string `json:"contentTypes" yaml:"contentTypes"`
}
