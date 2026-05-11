package kubeconfig

import (
	"path/filepath"

	"github.com/foomo/posh/pkg/util/files"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

func LoadFromFile(path string) (*api.Config, error) {
	return clientcmd.LoadFromFile(path)
}

func Write(c *api.Config, dirname string) error {
	if err := files.MkdirAll(dirname); err != nil {
		return err
	}

	return WriteToFile(c, filepath.Join(dirname, c.CurrentContext+".yaml"))
}

func WriteToFile(c *api.Config, filename string) error {
	return clientcmd.WriteToFile(*c, filename)
}

func FilterContext(c *api.Config, name string) {
	context := c.Contexts[name]
	c.Contexts = map[string]*api.Context{
		name: context,
	}

	cluster := c.Clusters[context.Cluster]
	c.Clusters = map[string]*api.Cluster{
		name: cluster,
	}

	authInfo := c.AuthInfos[context.AuthInfo]
	c.AuthInfos = map[string]*api.AuthInfo{
		context.AuthInfo: authInfo,
	}

	c.CurrentContext = name
}
