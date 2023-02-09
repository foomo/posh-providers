package gcloud

type Cluster struct {
	Project string `json:"project" yaml:"project"`
	Region  string `json:"region" yaml:"region"`
	Name    string `json:"name" yaml:"name"`
}
