package slack

import (
	"github.com/foomo/posh-providers/onepassword"
)

type Config struct {
	Token    onepassword.Secret `yaml:"token"`
	Channels map[string]string  `yaml:"channels"`
}
