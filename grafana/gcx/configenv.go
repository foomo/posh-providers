package gcx

type ConfigEnv struct {
	// URL of the grafana server
	URL string `json:"url" yaml:"url"`
	// KubeForward target config
	KubeForward *ConfigKubeForward `json:"kubeforward" yaml:"kubeforward"`
	// Env to pass to the cli
	Env []string `json:"env" yaml:"env"`
	// OrgID to set
	OrgID string `json:"orgId" yaml:"orgId"`
	// Token to set
	Token string `json:"token" yaml:"token"`
	// TokenCmd to execute returning the token
	TokenCmd []string `json:"tokenCmd" yaml:"tokenCmd"`
}
