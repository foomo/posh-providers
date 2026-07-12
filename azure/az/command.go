package az

import (
	"bytes"
	"context"
	"strings"

	"github.com/foomo/posh-providers/kubernetes/kubeconfig"
	"github.com/foomo/posh-providers/kubernetes/kubectl"
	"github.com/foomo/posh/pkg/command/tree"
	pkgexec "github.com/foomo/posh/pkg/exec"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/pkg/errors"
)

type (
	Command struct {
		l             log.Logger
		name          string
		az            *AZ
		kubectl       *kubectl.Kubectl
		commandTree   tree.Root
		middlewares   []pkgexec.Middleware
		clusterNameFn ClusterNameFn
	}
	ClusterNameFn func(name string, cluster Cluster) string
	CommandOption func(*Command)
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func CommandWithName(v string) CommandOption {
	return func(o *Command) {
		o.name = v
	}
}

func CommandWithMiddlewares(v ...pkgexec.Middleware) CommandOption {
	return func(o *Command) {
		o.middlewares = append(o.middlewares, v...)
	}
}

func CommandWithClusterNameFn(v ClusterNameFn) CommandOption {
	return func(o *Command) {
		o.clusterNameFn = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, az *AZ, kubectl *kubectl.Kubectl, opts ...CommandOption) *Command {
	inst := &Command{
		l:       l.Named("az"),
		name:    "az",
		az:      az,
		kubectl: kubectl,
		clusterNameFn: func(name string, cluster Cluster) string {
			return name
		},
	}

	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Manage azure resources",
		Nodes: tree.Nodes{
			{
				Name:        "login",
				Description: "Log in to Azure",
				Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
					fs.Internal().String("service-principal", "", "Service principal to use for authentication")

					if err := fs.Internal().SetValues("service-principal", inst.az.Config().ServicePrincipalNames()...); err != nil {
						return err
					}

					return nil
				},
				Execute: inst.login,
			},
			{
				Name:        "logout",
				Description: "Log out to remove access to Azure subscriptions",
				Execute:     inst.exec,
			},
			{
				Name:        "vault",
				Description: "Manage key vault entries",
				Nodes: tree.Nodes{
					{
						Name:        "subscription",
						Description: "Name of the subscription",
						Values: func(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
							return suggests.List(inst.az.cfg.SubscriptionNames())
						},
						Nodes: tree.Nodes{
							{
								Name:        "vault",
								Description: "Name of the vault",
								Values: func(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
									ret, err := inst.az.cfg.Subscription(r.Args().At(1))
									if err != nil {
										return nil
									}

									return suggests.List(ret.VaultNames())
								},
								Nodes: tree.Nodes{
									{
										Name:        "key",
										Description: "Manage keys",
										Nodes: tree.Nodes{
											{
												Name:        "create",
												Description: "Create a new key or key version",
												Args: tree.Args{
													{
														Name:        "name",
														Description: "Name of the key",
													},
												},
												Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
													fs.Default().String("kty", "", "Type of key to create")
													fs.Default().Int("size", 0, "Key size in bits (for RSA keys)")
													fs.Default().StringArray("ops", nil, "Permitted JSON web key operations")

													return fs.Default().SetValues("kty", "RSA", "RSA-HSM", "EC", "EC-HSM", "oct", "oct-HSM")
												},
												Execute: inst.vaultKeyCreate,
											},
											{
												Name:        "list",
												Description: "List keys in the vault",
												Execute:     inst.vaultKeyList,
											},
											{
												Name:        "delete",
												Description: "Delete a key from the vault",
												Args: tree.Args{
													{
														Name:        "name",
														Description: "Name of the key",
														Suggest:     inst.completeVaultEntries,
													},
												},
												Execute: inst.vaultKeyDelete,
											},
										},
									},
									{
										Name:        "secret",
										Description: "Manage secrets",
										Nodes: tree.Nodes{
											{
												Name:        "set",
												Description: "Create or update a secret",
												Args: tree.Args{
													{
														Name:        "name",
														Description: "Name of the secret",
													},
												},
												Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
													fs.Default().String("value", "", "Plain text secret value")
													fs.Default().String("file", "", "Source file for the secret value")
													fs.Default().String("content-type", "", "Type of the secret value such as a password")

													return nil
												},
												Execute: inst.vaultSecretSet,
											},
											{
												Name:        "list",
												Description: "List secrets in the vault",
												Execute:     inst.vaultSecretList,
											},
											{
												Name:        "delete",
												Description: "Delete a secret from the vault",
												Args: tree.Args{
													{
														Name:        "name",
														Description: "Name of the secret",
														Suggest:     inst.completeVaultEntries,
													},
												},
												Execute: inst.vaultSecretDelete,
											},
										},
									},
									{
										Name:        "certificate",
										Description: "Manage certificates",
										Nodes: tree.Nodes{
											{
												Name:        "set",
												Description: "Import a certificate from a file",
												Args: tree.Args{
													{
														Name:        "name",
														Description: "Name of the certificate",
													},
												},
												Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
													fs.Default().String("file", "", "Path to the PEM or PFX certificate file")
													fs.Default().String("password", "", "Password for the certificate file, if any")

													return nil
												},
												Execute: inst.vaultCertificateSet,
											},
											{
												Name:        "list",
												Description: "List certificates in the vault",
												Execute:     inst.vaultCertificateList,
											},
											{
												Name:        "delete",
												Description: "Delete a certificate from the vault",
												Args: tree.Args{
													{
														Name:        "name",
														Description: "Name of the certificate",
														Suggest:     inst.completeVaultEntries,
													},
												},
												Execute: inst.vaultCertificateDelete,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			{
				Name:        "artifactory",
				Description: "Login into the artifactory",
				Args: tree.Args{
					{
						Name:        "subscription",
						Description: "Name of the subscription",
						Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
							return suggests.List(inst.az.cfg.SubscriptionNames())
						},
					},
					{
						Name:        "artifactory",
						Description: "Name of the artifactory",
						Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
							ret, err := inst.az.cfg.Subscription(r.Args().At(1))
							if err != nil {
								return nil
							}

							return suggests.List(ret.ArtifactoryNames())
						},
					},
				},
				Execute: inst.artifactory,
			},
			{
				Name:        "kubeconfig",
				Description: "Retrieve credentials to access remote cluster",
				Args: tree.Args{
					{
						Name:        "subscription",
						Description: "Name of the subscription",
						Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
							return suggests.List(inst.az.cfg.SubscriptionNames())
						},
					},
					{
						Name:        "cluster",
						Description: "Name of the cluster",
						Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
							ret, err := inst.az.cfg.Subscription(r.Args().At(1))
							if err != nil {
								return nil
							}

							return suggests.List(ret.ClusterNames())
						},
					},
				},
				Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
					fs.Internal().String("profile", "", "Store credentials in given profile")
					return fs.Internal().SetValues("profile", "azure")
				},
				Execute: inst.kubeconfig,
			},
			{
				Name:        "raw",
				Description: "Execute raw Azure CLI commands",
				Execute:     inst.raw,
			},
		},
	})

	return inst
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

func (c *Command) login(ctx context.Context, r *readline.Readline) error {
	fs := r.FlagSets().Default()
	ifs := r.FlagSets().Internal()

	servicePricipal, err := ifs.GetString("service-principal")
	if err != nil {
		return err
	}

	var args []string

	if servicePricipal != "" {
		sp, err := c.az.cfg.ServicePrincipal(servicePricipal)
		if err != nil {
			return err
		}

		args = append(args,
			"--service-principal",
			"--username", sp.ClientID,
			"--password", sp.ClientSecret,
			"--tenant", sp.TenantID,
		)
	} else {
		args = append(args,
			"--allow-no-subscriptions",
			"--tenant", c.az.Config().TenantID,
		)
	}

	return c.cmd(ctx, "login").
		Args(args...).
		Args(fs.Visited().Args()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run()
}

func (c *Command) artifactory(ctx context.Context, r *readline.Readline) error {
	sub, err := c.az.cfg.Subscription(r.Args().At(1))
	if err != nil {
		return errors.Errorf("failed to retrieve subscription for: %q", r.Args().At(1))
	}

	acr, err := sub.Artifactory(r.Args().At(2))
	if err != nil {
		return errors.Errorf("failed to retrieve artifactoy for: %q", r.Args().At(2))
	}

	return c.cmd(ctx, "acr", "login",
		"--name", acr.Name,
		"--subscription", sub.Name,
		"--resource-group", acr.ResourceGroup,
	).Run()
}

func (c *Command) kubeconfig(ctx context.Context, r *readline.Readline) error {
	ifs := r.FlagSets().Internal()

	sub, err := c.az.cfg.Subscription(r.Args().At(1))
	if err != nil {
		return errors.Errorf("failed to retrieve subscription for: %q", r.Args().At(1))
	}

	k8s, err := sub.Cluster(r.Args().At(2))
	if err != nil {
		return errors.Errorf("failed to retrieve cluster for: %q", r.Args().At(2))
	}

	kubectlCluster := c.kubectl.Cluster(c.clusterNameFn(r.Args().At(2), k8s))
	if kubectlCluster == nil {
		return errors.Errorf("failed to retrieve kubectl cluster for: %q", k8s.Name)
	}

	profile, err := ifs.GetString("profile")
	if err != nil {
		return err
	}

	c.l.Success("retrieved kubectl cluster config")

	if err := c.cmd(ctx, "aks", "get-credentials",
		"--name", k8s.Name,
		"--subscription", sub.Name,
		"--resource-group", k8s.ResourceGroup,
		"--overwrite-existing",
	).
		Env(kubectlCluster.Env(profile)).
		Run(); err != nil {
		return err
	}

	c.l.Success("converted kubectl cluster config using kubelogin")

	if err := pkgexec.NewCommand(ctx, "kubelogin", "convert-kubeconfig",
		"-l", "azurecli",
	).
		Env(kubectlCluster.Env(profile)).
		Run(); err != nil {
		return err
	}

	if k8s.ProxyURL != "" {
		c.l.Info("setting proxy URL:", k8s.ProxyURL)

		kc, err := kubeconfig.LoadFromFile(kubectlCluster.Config(profile))
		if err != nil {
			return err
		}

		if kc.Clusters[kc.Contexts[kc.CurrentContext].Cluster].ProxyURL != k8s.ProxyURL {
			kc.Clusters[kc.Contexts[kc.CurrentContext].Cluster].ProxyURL = k8s.ProxyURL
			if err := kubeconfig.WriteToFile(kc, kubectlCluster.Config(profile)); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Command) vaultTarget(r *readline.Readline) (Subscription, Vault, error) {
	sub, err := c.az.cfg.Subscription(r.Args().At(1))
	if err != nil {
		return Subscription{}, Vault{}, errors.Errorf("failed to retrieve subscription for: %q", r.Args().At(1))
	}

	vault, err := sub.Vault(r.Args().At(2))
	if err != nil {
		return Subscription{}, Vault{}, errors.Errorf("failed to retrieve vault for: %q", r.Args().At(2))
	}

	return sub, vault, nil
}

func (c *Command) vaultKeyCreate(ctx context.Context, r *readline.Readline) error {
	sub, vault, err := c.vaultTarget(r)
	if err != nil {
		return err
	}

	return c.cmd(ctx, "keyvault", "key", "create",
		"--name", r.Args().At(5),
		"--vault-name", vault.Name,
		"--subscription", sub.Name,
	).
		Args(r.FlagSets().Default().Visited().Args()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run()
}

func (c *Command) vaultKeyList(ctx context.Context, r *readline.Readline) error {
	sub, vault, err := c.vaultTarget(r)
	if err != nil {
		return err
	}

	return c.cmd(ctx, "keyvault", "key", "list",
		"--vault-name", vault.Name,
		"--subscription", sub.Name,
	).
		Args(r.FlagSets().Default().Visited().Args()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run()
}

func (c *Command) vaultKeyDelete(ctx context.Context, r *readline.Readline) error {
	sub, vault, err := c.vaultTarget(r)
	if err != nil {
		return err
	}

	return c.cmd(ctx, "keyvault", "key", "delete",
		"--name", r.Args().At(5),
		"--vault-name", vault.Name,
		"--subscription", sub.Name,
	).
		Args(r.FlagSets().Default().Visited().Args()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run()
}

func (c *Command) vaultSecretSet(ctx context.Context, r *readline.Readline) error {
	sub, vault, err := c.vaultTarget(r)
	if err != nil {
		return err
	}

	return c.cmd(ctx, "keyvault", "secret", "set",
		"--name", r.Args().At(5),
		"--vault-name", vault.Name,
		"--subscription", sub.Name,
	).
		Args(r.FlagSets().Default().Visited().Args()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run()
}

func (c *Command) vaultSecretList(ctx context.Context, r *readline.Readline) error {
	sub, vault, err := c.vaultTarget(r)
	if err != nil {
		return err
	}

	return c.cmd(ctx, "keyvault", "secret", "list",
		"--vault-name", vault.Name,
		"--subscription", sub.Name,
	).
		Args(r.FlagSets().Default().Visited().Args()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run()
}

func (c *Command) vaultSecretDelete(ctx context.Context, r *readline.Readline) error {
	sub, vault, err := c.vaultTarget(r)
	if err != nil {
		return err
	}

	return c.cmd(ctx, "keyvault", "secret", "delete",
		"--name", r.Args().At(5),
		"--vault-name", vault.Name,
		"--subscription", sub.Name,
	).
		Args(r.FlagSets().Default().Visited().Args()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run()
}

func (c *Command) vaultCertificateSet(ctx context.Context, r *readline.Readline) error {
	sub, vault, err := c.vaultTarget(r)
	if err != nil {
		return err
	}

	return c.cmd(ctx, "keyvault", "certificate", "import",
		"--name", r.Args().At(5),
		"--vault-name", vault.Name,
		"--subscription", sub.Name,
	).
		Args(r.FlagSets().Default().Visited().Args()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run()
}

func (c *Command) vaultCertificateList(ctx context.Context, r *readline.Readline) error {
	sub, vault, err := c.vaultTarget(r)
	if err != nil {
		return err
	}

	return c.cmd(ctx, "keyvault", "certificate", "list",
		"--vault-name", vault.Name,
		"--subscription", sub.Name,
	).
		Args(r.FlagSets().Default().Visited().Args()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run()
}

func (c *Command) vaultCertificateDelete(ctx context.Context, r *readline.Readline) error {
	sub, vault, err := c.vaultTarget(r)
	if err != nil {
		return err
	}

	return c.cmd(ctx, "keyvault", "certificate", "delete",
		"--name", r.Args().At(5),
		"--vault-name", vault.Name,
		"--subscription", sub.Name,
	).
		Args(r.FlagSets().Default().Visited().Args()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run()
}

func (c *Command) completeVaultEntries(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
	sub, vault, err := c.vaultTarget(r)
	if err != nil {
		c.l.Debug(err.Error())
		return nil
	}

	noun := r.Args().At(3) // key | secret | certificate

	return c.az.cache.GetSuggests("vault-"+noun+"-"+sub.Name+"-"+vault.Name, func() any {
		var buf bytes.Buffer

		if err := c.cmd(ctx, "keyvault", noun, "list",
			"--vault-name", vault.Name,
			"--subscription", sub.Name,
			"--query", "[].name",
			"-o", "tsv",
		).Stdout(&buf).Run(); err != nil {
			c.l.Debug(err.Error())
			return []goprompt.Suggest{}
		}

		var ret []string

		for line := range strings.SplitSeq(strings.TrimSpace(buf.String()), "\n") {
			if line = strings.TrimSpace(line); line != "" {
				ret = append(ret, line)
			}
		}

		return suggests.List(ret)
	})
}

func (c *Command) raw(ctx context.Context, r *readline.Readline) error {
	return c.cmd(ctx, r.Args()[1:]...).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run()
}

func (c *Command) exec(ctx context.Context, r *readline.Readline) error {
	return c.cmd(ctx, r.Args()...).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run()
}

func (c *Command) cmd(ctx context.Context, args ...string) *pkgexec.Command {
	return pkgexec.NewCommand(ctx, "az", args...).Middleware(c.middlewares...)
}
