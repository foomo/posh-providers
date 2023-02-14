package gcloud

import (
	"github.com/foomo/posh-providers/onepassword"
)

type Account struct {
	Name  string              `json:"name" yaml:"name"`
	Email string              `json:"email" yaml:"email"`
	Key   *onepassword.Secret `json:"key" yaml:"key"`
}
