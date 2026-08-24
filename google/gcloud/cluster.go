package gcloud

const (
	ClusterNameDefault string = "default"
)

type Cluster struct {
	// Name of the cluster in GKE, which may differ from the key used to select it
	Name string `json:"name" yaml:"name"`
	// Google cloud project the cluster belongs to
	Project string `json:"project" yaml:"project"`
	// Google cloud region the cluster runs in
	Region string `json:"region" yaml:"region"`
	// Account key to authenticate as; passed to gcloud as --account when set
	Account string `json:"account" yaml:"account"`
}
