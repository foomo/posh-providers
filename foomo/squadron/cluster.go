package squadron

type Cluster struct {
	// Cluser name
	Name string `json:"name" yaml:"name"`
	// Enable notification by default
	Notify bool `json:"notify" yaml:"notify"`
	// Cluster fleet names
	Fleets []string `json:"fleets" yaml:"fleets"`
}
