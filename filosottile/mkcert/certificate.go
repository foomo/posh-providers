package mkcert

type Certificate struct {
	// Basename of the generated pair, written as `<name>.pem` and
	// `<name>-key.pem`. Does not need to be one of `names`.
	Name string `json:"name" yaml:"name"`
	// Hostnames, wildcards, IPs, URLs or emails the certificate is valid for.
	Names []string `json:"names" yaml:"names"`
}
