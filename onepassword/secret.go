package onepassword

// Secret addresses a single field of a 1Password item, in the same terms as an
// `op://` reference: `op://<vault>/<item>/<field>`, resolved against Account.
type Secret struct {
	// Account is the 1Password account the vault belongs to, as a sign-in
	// address or account UUID.
	Account string `json:"account" yaml:"account"`
	// Vault is the name or UUID of the vault holding the item.
	Vault string `json:"vault" yaml:"vault"`
	// Item is the name or UUID of the item within the vault.
	Item string `json:"item" yaml:"item"`
	// Field is the name of the field to read from the item, e.g. `password`.
	Field string `json:"field" yaml:"field"`
}
