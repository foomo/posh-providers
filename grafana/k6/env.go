package k6

// Env holds environment variables passed to a k6 scenario, upper-cased and
// exposed to the script via __ENV.
type Env map[string]string
