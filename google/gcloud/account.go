package gcloud

import (
	"github.com/foomo/posh-providers/onepassword"
)

type Account struct {
	// Account identity passed to gcloud as --account
	Name string `json:"name" yaml:"name"`
	// Email of the account, for reference only
	Email string `json:"email" yaml:"email"`
	// 1Password document holding the service account key; without it, login falls back to an interactive browser prompt
	Key *onepassword.Secret `json:"key" yaml:"key"`
}
