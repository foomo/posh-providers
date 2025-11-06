package kubeconfig

import (
	"os"
	"path/filepath"

	"github.com/foomo/posh/pkg/util/files"
	"gopkg.in/yaml.v3"
)

type KubeConfig struct {
	APIVersion     string    `yaml:"apiVersion"`
	Clusters       []Cluster `yaml:"clusters"`
	Contexts       []Context `yaml:"contexts"`
	CurrentContext string    `yaml:"current-context"`
	Kind           string    `yaml:"kind"`
	Preferences    any       `yaml:"preferences"`
	Users          []User    `yaml:"users"`
}

func Read(path, cluster string) (*KubeConfig, error) {
	var c *KubeConfig

	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(bytes, &c); err != nil {
		return nil, err
	}

	c.filter(cluster)

	return c, nil
}

func (t *KubeConfig) Write(dirname string) error {
	bytes, err := yaml.Marshal(t)
	if err != nil {
		return err
	}

	if err := files.MkdirAll(dirname); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dirname, t.CurrentContext+".yaml"), bytes, 0600)
}

func (t *KubeConfig) filter(cluster string) {
	t.CurrentContext = cluster

	var clusters []Cluster

	for _, c := range t.Clusters {
		if c.Name == cluster {
			clusters = append(clusters, c)
		}
	}

	t.Clusters = clusters

	var contexts []Context

	for _, c := range t.Contexts {
		if c.Name == cluster {
			contexts = append(contexts, c)
		}
	}

	t.Contexts = contexts

	var users []User

	for _, u := range t.Users {
		if u.Name == cluster {
			users = append(users, u)
		}
	}

	t.Users = users
}
