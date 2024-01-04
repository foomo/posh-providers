package squadron

type Cluster struct {
	Name   string   `json:"name" yaml:"name"`
	Fleets []string `json:"fleets" yaml:"fleets"`
}
