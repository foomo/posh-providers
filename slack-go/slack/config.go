package slack

import (
	"github.com/foomo/posh-providers/onepassword"
)

// Config holds the slack provider's configuration.
type Config struct {
	Token    onepassword.Secret            `json:"token" yaml:"token"`
	Channels map[string]string             `json:"channels" yaml:"channels"`
	Webhooks map[string]onepassword.Secret `json:"webhooks" yaml:"webhooks"`
}
