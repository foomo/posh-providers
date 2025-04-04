package squadron

type Cluster struct {
	// Cluser name
	Name string `json:"name" yaml:"name"`
	// Enable notification by default
	Notify bool `json:"notify" yaml:"notify"`
	// Enable confirmation
	Confirm bool `json:"confirm" yaml:"confirm"`
	// Cluster fleet names
	Fleets []string `json:"fleets" yaml:"fleets"`
}
