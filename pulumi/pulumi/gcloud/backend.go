package pulumi

import (
	"github.com/foomo/posh-providers/onepassword"
)

type Backend struct {
	Location   string             `json:"location" yaml:"location"`
	Bucket     string             `json:"bucket" yaml:"bucket"`
	Project    string             `json:"project" yaml:"project"`
	Passphrase onepassword.Secret `json:"passphrase" yaml:"passphrase"`
}
