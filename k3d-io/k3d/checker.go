package k3d

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/check"
	"github.com/foomo/posh/pkg/shell"
)

func ClusterChecker(inst *K3d) check.Checker {
	return func(ctx context.Context, l log.Logger) []check.Info {
		title := "K3d"

		res, err := shell.New(ctx, l, "k3d", "cluster", "list", "-o", "json").CombinedOutput()
		if err != nil {
			return []check.Info{check.NewFailureInfo("⚡︎", title, "Down")}
		}

		var clusters []struct {
			Name           string `json:"name"`
			ServersRunning int    `json:"serversRunning"`
		}
		if err := json.Unmarshal(res, &clusters); err != nil {
			return []check.Info{check.NewFailureInfo("⚡︎", title, "Unknown")}
		}

		var ret []check.Info
		for _, name := range inst.cfg.ClusterNames() {
			var found bool
			c, _ := inst.cfg.Cluster(name)
			title += " (" + c.Alias + ")"
			note := "127.0.0.1:" + c.Port
			for _, cluster := range clusters {
				if cluster.Name == name {
					found = true
					if cluster.ServersRunning == 0 {
						ret = append(ret, check.NewNoteInfo("☸", title, note))
					} else {
						ret = append(ret, check.NewSuccessInfo("☸", title, note))
					}
				}
			}
			if !found {
				ret = append(ret, check.NewNoteInfo("☸", title, "Stopped"))
			}
		}

		return ret
	}
}

func RegistryChecker(inst *K3d) check.Checker {
	return func(ctx context.Context, l log.Logger) []check.Info {
		title := "K3d registry"

		res, err := shell.New(ctx, l, "k3d", "registry", "list", "-o", "json").CombinedOutput()
		if err != nil {
			return []check.Info{check.NewFailureInfo("⚡︎", title, "Down")}
		}

		var registries []struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(res, &registries); err != nil {
			return []check.Info{check.NewFailureInfo("⚡︎", title, "Unknown")}
		}

		for _, registry := range registries {
			if registry.Name == fmt.Sprintf("k3d-%s", inst.cfg.Registry.Name) {
				ips, err := net.DefaultResolver.LookupIPAddr(ctx, registry.Name)
				if err != nil {
					return []check.Info{check.NewFailureInfo("⚡︎", title, err.Error())}
				}

				var configured bool

				for _, ip := range ips {
					if ip.IP.String() == "127.0.0.1" {
						configured = true
						break
					}
				}

				if !configured {
					return []check.Info{check.NewFailureInfo("⚡︎", title, "Missing /etc/hosts entry for: "+registry.Name)}
				}

				return []check.Info{check.NewSuccessInfo("☸", title, "127.0.0.1:"+inst.cfg.Registry.Port)}
			}
		}

		return []check.Info{check.NewNoteInfo("☸", title, "127.0.0.1:"+inst.cfg.Registry.Port)}
	}
}
