package docusaurus

type Config struct {
	SourcePath string   `json:"sourcePath" yaml:"sourcePath"`
	LocalPort  string   `json:"localPort" yaml:"localPort"`
	NodeTag    string   `json:"nodeTag" yaml:"nodeTag"`
	ImageTag   string   `json:"imageTag" yaml:"imageTag"`
	ImageName  string   `json:"imageName" yaml:"imageName"`
	Volumes    []string `json:"volumes" yaml:"volumes"`
}
