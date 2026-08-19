package onepassword

type Config struct {
	// Account this provider signs in to and resolves every secret against, as a
	// sign-in subdomain or account UUID. `register` appends `.1password.eu` to it,
	// so that verb assumes an EU-hosted account.
	Account string `json:"account" yaml:"account"`
	// Optional file the session token is written to and reloaded from, so a session
	// survives restarting the shell. Written mode 0600 and holding a live
	// credential, so keep it out of version control; leave unset to keep the
	// session in the process environment only.
	TokenFilename string `json:"tokenFilename" yaml:"tokenFilename"`
}
