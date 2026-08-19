package gocontentful

// Config is the schema of a `gocontentful.yaml` file discovered in the project,
// not of a key in the posh config. It is rendered through 1Password before being
// unmarshalled, so its values may be secret references.
type Config struct {
	// Contentful space id to generate against
	SpaceID string `json:"spaceId" yaml:"spaceId"`
	// Contentful management API key, typically a 1Password secret reference
	CMAKey string `json:"cmaKey" yaml:"cmaKey"`
	// Contentful environment; defaults to "master" when empty
	Environment string `json:"environment,omitempty" yaml:"environment,omitempty"`
	// Contentful API region
	Region string `json:"region,omitempty" yaml:"region,omitempty"`
	// Content types to generate code for
	ContentTypes []string `json:"contentTypes" yaml:"contentTypes"`
}
