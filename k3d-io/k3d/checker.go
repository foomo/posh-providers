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
		title := "K3d cluster"

		res, err := shell.New(ctx, l, "k3d", "cluster", "list", "-o", "json").Output()
		if err != nil {
			return []check.Info{check.NewFailureInfo("", title, fmt.Sprintf("Failed to list clusters (%s)", err.Error()))}
		}

		var clusters []struct {
			Name           string `json:"name"`
			ServersRunning int    `json:"serversRunning"`
		}
		if err := json.Unmarshal(res, &clusters); err != nil {
			return []check.Info{check.NewFailureInfo("", title, fmt.Sprintf("Failed to unmarshal clusters (%s)", err.Error()))}
		}

		for _, cluster := range clusters {
			if cluster.Name == inst.cfg.Registry.Name {
				if cluster.ServersRunning == 0 {
					return []check.Info{check.NewNoteInfo("", title, "Paused")}
				}
				return []check.Info{check.NewSuccessInfo("", title, "127.0.0.1:"+inst.cfg.Registry.Port)}
			}
		}

		return []check.Info{check.NewNoteInfo("", title, "127.0.0.1:"+inst.cfg.Registry.Port)}
	}
}

func RegistryChecker(inst *K3d) check.Checker {
	return func(ctx context.Context, l log.Logger) []check.Info {
		title := "K3d registry"

		res, err := shell.New(ctx, l, "k3d", "registry", "list", "-o", "json").Output()
		if err != nil {
			return []check.Info{check.NewFailureInfo("", title, err.Error())}
		}

		var registries []struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(res, &registries); err != nil {
			return []check.Info{check.NewFailureInfo("", title, err.Error())}
		}

		for _, registry := range registries {
			if registry.Name == fmt.Sprintf("k3d-%s", inst.cfg.Registry.Name) {
				ips, err := net.DefaultResolver.LookupIPAddr(ctx, registry.Name)
				if err != nil {
					return []check.Info{check.NewFailureInfo("", title, err.Error())}
				}

				var configured bool

				for _, ip := range ips {
					if ip.IP.String() == "127.0.0.1" {
						configured = true
						break
					}
				}

				if !configured {
					return []check.Info{check.NewFailureInfo("", title, "Missing /etc/hosts entry for: "+registry.Name)}
				}

				return []check.Info{check.NewSuccessInfo("", title, "127.0.0.1:"+inst.cfg.Registry.Port)}
			}
		}

		return []check.Info{check.NewNoteInfo("", title, "127.0.0.1:"+inst.cfg.Registry.Port)}
	}
}
