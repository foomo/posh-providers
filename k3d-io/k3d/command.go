package k3d

import (
	"context"
	"fmt"

	"github.com/foomo/posh-providers/kubernetes/kubectl"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/env"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/files"
	"github.com/foomo/posh/pkg/util/suggests"
)

type (
	Command struct {
		l           log.Logger
		k3d         *K3d
		name        string
		cache       cache.Cache
		kubectl     *kubectl.Kubectl
		commandTree tree.Root
	}
	CommandOption func(*Command) error
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func CommandWithName(v string) CommandOption {
	return func(o *Command) error {
		o.name = v
		return nil
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, k3d *K3d, cache cache.Cache, kubectl *kubectl.Kubectl, opts ...CommandOption) (*Command, error) {
	inst := &Command{
		l:       l.Named("k3d"),
		k3d:     k3d,
		cache:   cache,
		kubectl: kubectl,
		name:    "k3d",
	}

	for _, opt := range opts {
		if opt != nil {
			if err := opt(inst); err != nil {
				return nil, err
			}
		}
	}

	nameArg := &tree.Arg{
		Name:        "name",
		Description: "Cluster name",
		Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
			return suggests.List(inst.k3d.Config().ClusterNames())
		},
	}
	chartArg := &tree.Arg{
		Name:        "chart",
		Description: "Chart name",
		Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
			if value, err := inst.k3d.Config().Charts.Names(); err != nil {
				inst.l.Debug(err.Error())
				return nil
			} else {
				return suggests.List(value)
			}
		},
	}
	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Execute K3d commands",
		Nodes: tree.Nodes{
			{
				Name:        "up",
				Description: "Spin up configured cluster",
				Args:        tree.Args{nameArg},
				Execute:     inst.up,
			},
			{
				Name:        "pause",
				Description: "Stop existing k3d cluster",
				Args:        tree.Args{nameArg},
				Execute:     inst.pause,
			},
			{
				Name:        "resume",
				Description: "Start existing k3d cluster",
				Args:        tree.Args{nameArg},
				Execute:     inst.resume,
			},
			{
				Name:        "kubeconfig",
				Description: "Retrieve kubeconfig from running cluster",
				Args:        tree.Args{nameArg},
				Execute:     inst.kubeconfig,
			},
			{
				Name:        "install",
				Description: "Install predified charts",
				Args:        tree.Args{nameArg, chartArg},
				Execute:     inst.install,
			},
			{
				Name:        "uninstall",
				Description: "Uninstall predefined charts",
				Args:        tree.Args{nameArg, chartArg},
				Execute:     inst.uninstall,
			},
			{
				Name:        "down",
				Description: "Shut down configured cluster",
				Args:        tree.Args{nameArg},
				Execute:     inst.down,
			},
		},
	})

	return inst, nil
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (c *Command) Name() string {
	return c.commandTree.Node().Name
}

func (c *Command) Description() string {
	return c.commandTree.Node().Description
}

func (c *Command) Complete(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
	return c.commandTree.Complete(ctx, r)
}

func (c *Command) Execute(ctx context.Context, r *readline.Readline) error {
	return c.commandTree.Execute(ctx, r)
}

func (c *Command) Help(ctx context.Context, r *readline.Readline) string {
	return c.commandTree.Help(ctx, r)
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

/*
@k3d registry create $(CLUSTER_NAME)-registry --port 12345

	@k3d cluster create $(CLUSTER_NAME) \
		--image=rancher/k3s:v1.23.16-k3s1 \
		--kubeconfig-update-default=false \
		--kubeconfig-switch-context=false \
		--registry-use k3d-$(CLUSTER_NAME)-registry:12345 \
		-p "$(CLUSTER_PORT):443@loadbalancer" \
		--k3s-arg "--disable=traefik@server:0"
		--agents 1

@bin/helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
@bin/helm dependency update devops/helm/charts/base
@bin/helm dependency build devops/helm/charts/base
@bin/helm upgrade --install base devops/helm/charts/base
@kubectl create ns catalogue
@kubectl create ns checkout
@kubectl create ns general
@kubectl create ns msdconnect
@kubectl create ns profile
@kubectl create ns site
@kubectl create ns stock
@kubectl create ns toolbox
*/
func (c *Command) up(ctx context.Context, r *readline.Readline) error {
	cfg := c.k3d.Config()
	_, args := r.Args().Shift()
	name, args := args.Shift()

	// ensure registry
	registry, err := c.k3d.Registry(ctx, cfg.Registry.Name)
	if err != nil {
		return err
	}

	if registry == nil {
		// create registry
		if err := shell.New(ctx, c.l, "k3d", "registry", "create", cfg.Registry.Name,
			"--port", cfg.Registry.Port,
			"--no-help",
		).Run(); err != nil {
			return err
		}
	}

	clusterCfg, err := cfg.Cluster(name)
	if err != nil {
		return err
	}

	// ensure cluster
	cluster, err := c.k3d.Cluster(ctx, clusterCfg.AliasName())
	if err != nil {
		return err
	} else if cluster != nil {
		c.l.Info("cluster already exists")
		return nil
	}

	k3sArg := ""
	if !clusterCfg.EnableTraefikRouter {
		k3sArg = "--k3s-arg \"--disable=traefik@server:0\""
	}

	return shell.New(ctx, c.l, "k3d", "cluster", "create", clusterCfg.AliasName(),
		"--image", clusterCfg.Image,
		"--registry-use", fmt.Sprintf("%s:%s", cfg.Registry.Name, cfg.Registry.Port),
		"--port", fmt.Sprintf("%s:443@loadbalancer", clusterCfg.Port),
		k3sArg,
		"--agents", "1",
	).
		Env(c.kubectl.Cluster(name).Env("")).
		Args(args...).
		Args(clusterCfg.Args...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run()
}

func (c *Command) pause(ctx context.Context, r *readline.Readline) error {
	cfg := c.k3d.Config()
	_, args := r.Args().Shift()
	name, args := args.Shift()

	clusterCfg, err := cfg.Cluster(name)
	if err != nil {
		return err
	}

	cluster, err := c.k3d.Cluster(ctx, clusterCfg.AliasName())
	if err != nil {
		return err
	} else if cluster == nil {
		c.l.Info("cluster does not exists")
		return nil
	}

	return shell.New(ctx, c.l, "k3d", "cluster", "stop", clusterCfg.AliasName()).
		Env(c.kubectl.Cluster(name).Env("")).
		Args(args...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run()
}

func (c *Command) resume(ctx context.Context, r *readline.Readline) error {
	cfg := c.k3d.Config()
	_, args := r.Args().Shift()
	name, args := args.Shift()

	clusterCfg, err := cfg.Cluster(name)
	if err != nil {
		return err
	}

	cluster, err := c.k3d.Cluster(ctx, clusterCfg.AliasName())
	if err != nil {
		return err
	} else if cluster == nil {
		c.l.Info("cluster does not exists")
		return nil
	}

	return shell.New(ctx, c.l, "k3d", "cluster", "start", clusterCfg.AliasName()).
		Env(c.kubectl.Cluster(name).Env("")).
		Args(args...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run()
}

func (c *Command) install(ctx context.Context, r *readline.Readline) error {
	var args []string

	cfg := c.k3d.Config()
	fs := r.FlagSets().Default()
	cluster, name := r.Args().At(1), r.Args().At(2)

	// allow values.override.yaml files
	if err := files.Exists(env.Path(cfg.Charts.Path, name, "values.override.yaml")); err == nil {
		args = append(args, "--values", env.Path(cfg.Charts.Path, name, "values.override.yaml"))
	}

	return shell.New(ctx, c.l, "helm",
		"upgrade", name,
		"--install",
		"--dependency-update",
		"--namespace", fmt.Sprintf("%s%s", cfg.Charts.Prefix, name),
		"--create-namespace",
		"--force",
		env.Path(cfg.Charts.Path, name),
	).
		Env(c.kubectl.Cluster(cluster).Env("")).
		Args(args...).
		Args(fs.Visited().Args()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run()
}

func (c *Command) uninstall(ctx context.Context, r *readline.Readline) error {
	fs := r.FlagSets().Default()
	cfg := c.k3d.Config()
	cluster, name := r.Args().At(1), r.Args().At(2)

	return shell.New(ctx, c.l, "helm",
		"uninstall",
		"--namespace", fmt.Sprintf("%s%s", cfg.Charts.Prefix, name),
		name,
	).
		Env(c.kubectl.Cluster(cluster).Env("")).
		Args(fs.Visited().Args()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run()
}

func (c *Command) kubeconfig(ctx context.Context, r *readline.Readline) error {
	cfg := c.k3d.Config()
	name := r.Args().At(1)

	clusterCfg, err := cfg.Cluster(name)
	if err != nil {
		return err
	}

	// ensure cluster
	cluster, err := c.k3d.Cluster(ctx, clusterCfg.AliasName())
	if err != nil {
		return err
	} else if cluster == nil {
		c.l.Info("cluster does not exist")
		return nil
	}

	// delete cluster
	if err := shell.New(ctx, c.l, "k3d", "kubeconfig", "get", clusterCfg.AliasName()).
		Args(">", c.kubectl.Cluster(name).Config("")).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run(); err != nil {
		return err
	}

	c.cache.Clear()

	return nil
}

/*
cluster.down: export KUBECONFIG=devops/config/kubectl/$(CLUSTER_NAME).yaml
cluster.down:

	@k3d cluster delete $(CLUSTER_NAME)
	@k3d registry delete $(CLUSTER_NAME)-registry
	@rm -f devops/config/kubectl/$(CLUSTER_NAME).yaml
*/
func (c *Command) down(ctx context.Context, r *readline.Readline) error {
	cfg := c.k3d.Config()
	name := r.Args().At(1)

	clusterCfg, err := cfg.Cluster(name)
	if err != nil {
		return err
	}

	cluster, err := c.k3d.Cluster(ctx, clusterCfg.AliasName())
	if err != nil {
		return err
	} else if cluster == nil {
		c.l.Info("cluster does not exist")
		return nil
	}

	// delete cluster
	if err := shell.New(ctx, c.l, "k3d", "cluster", "delete", clusterCfg.AliasName()).
		Env(c.kubectl.Cluster(name).Env("")).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run(); err != nil {
		return err
	}

	// delete config
	if err := c.kubectl.Cluster(name).DeleteConfig(""); err != nil {
		return err
	}

	// ensure registry
	registry, err := c.k3d.Registry(ctx, cfg.Registry.Name)
	if err != nil {
		return err
	}

	if registry != nil {
		// TODO check if empty
		// delete registry
		if err := shell.New(ctx, c.l, "k3d", "registry", "delete", cfg.Registry.Name).Run(); err != nil {
			return err
		}
	}

	return nil
}
