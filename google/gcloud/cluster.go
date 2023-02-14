package gcloud

import (
	"github.com/foomo/posh-providers/onepassword"
)

const (
	ClusterNameDefault string = "default"
	ClusterRoleDefault string = "default"
)

type Cluster struct {
	Name        string              `json:"name" yaml:"name"`
	FullName    string              `json:"fullName" yaml:"fullName"`
	Region      string              `json:"region" yaml:"region"`
	Roles       []string            `json:"roles" yaml:"roles"`
	AccessToken *onepassword.Secret `json:"accessToken" yaml:"accessToken"`
}

func (c Cluster) DefaultFullName() string {
	if c.FullName != "" {
		return c.FullName
	}
	return c.Name
}

func (c Cluster) DefaultRoles() []string {
	if len(c.Roles) != 0 {
		return c.Roles
	}
	return []string{ClusterRoleDefault}
}
