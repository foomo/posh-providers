package beam

type Database struct {
	Port     int    `json:"port" yaml:"port"`
	Hostname string `json:"hostname" yaml:"hostname"`
}
