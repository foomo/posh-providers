package chrome

type Profile struct {
	// App indicates whether the application mode is enabled for the profile.
	App bool `json:"app" yaml:"app"`
	// Default URL to open when no URL argument is given
	URL string `json:"url" yaml:"url"`
	// Proxy passed verbatim to Chrome as --proxy-server, e.g. socks5://localhost:1080
	Proxy string `json:"proxy" yaml:"proxy"`
	// Open in incognito mode
	Incognito bool `json:"incognito" yaml:"incognito"`
}
