package gcloud

const (
	ClusterNameDefault string = "default"
)

type Cluster struct {
	Name    string `json:"name" yaml:"name"`
	Project string `json:"project" yaml:"project"`
	Region  string `json:"region" yaml:"region"`
	Account string `json:"account" yaml:"account"`
}
