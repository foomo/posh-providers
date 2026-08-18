package docusaurus

type Config struct {
	// Path to the docusaurus site, added into the image at build time
	SourcePath string `json:"sourcePath" yaml:"sourcePath"`
	// Host port mapped to the container's port 3000
	LocalPort string `json:"localPort" yaml:"localPort"`
	// Node base image tag to build against
	NodeTag string `json:"nodeTag" yaml:"nodeTag"`
	// Tag of the image being built and run
	ImageTag string `json:"imageTag" yaml:"imageTag"`
	// Name of the image being built and run
	ImageName string `json:"imageName" yaml:"imageName"`
	// Additional docker volume mounts, environment variables are expanded
	Volumes []string `json:"volumes" yaml:"volumes"`
}
