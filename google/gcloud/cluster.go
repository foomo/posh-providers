package gcloud

const (
	// ClusterNameDefault is the conventional GKE cluster name "default".
	ClusterNameDefault string = "default"
)

// Cluster is a GKE cluster configuration: its name, project, region, and the
// optional account to authenticate as when fetching credentials.
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
