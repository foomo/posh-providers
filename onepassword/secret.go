package onepassword

type Secret struct {
	Account string `json:"account" yaml:"account"`
	Vault   string `json:"vault" yaml:"vault"`
	Item    string `json:"item" yaml:"item"`
	Field   string `json:"field" yaml:"field"`
}
