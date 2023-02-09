package gcloud

type Environment struct {
	Name     string    `json:"name" yaml:"name"`
	Clusters []Cluster `json:"clusters" yaml:"clusters"`
}
