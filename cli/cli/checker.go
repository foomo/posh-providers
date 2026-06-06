package cli

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/check"
	"github.com/foomo/posh/pkg/shell"
)

var scopesLineRE = regexp.MustCompile(`(?m)^\s*-\s*Token scopes:\s*(.+)$`)

func ScopesChecker(c *CLI) check.Checker {
	return func(ctx context.Context, l log.Logger) []check.Info {
		const icon = "⛹"

		cfg := c.Config()
		if len(cfg.Scopes) == 0 {
			return nil
		}

		hosts := make([]string, 0, len(cfg.Scopes))
		for h := range cfg.Scopes {
			hosts = append(hosts, h)
		}

		sort.Strings(hosts)

		infos := make([]check.Info, 0, len(hosts))

		for _, host := range hosts {
			name := "gh:" + host
			required := cfg.RequiredScopes(host)

			raw, err := shell.New(ctx, l, "gh", "auth", "status", "--hostname", host).CombinedOutput()
			if err != nil {
				infos = append(infos, check.NewNoteInfo(icon, name, "not authenticated"))

				continue
			}

			granted := parseGrantedScopes(string(raw))

			var missing []string

			for _, want := range required {
				if !containsScope(granted, want) {
					missing = append(missing, want)
				}
			}

			if len(missing) > 0 {
				infos = append(infos, check.NewFailureInfo(icon, name,
					fmt.Sprintf("missing scopes: %s", strings.Join(missing, ", "))))
			} else {
				infos = append(infos, check.NewSuccessInfo(icon, name,
					fmt.Sprintf("scopes: %s", strings.Join(required, ", "))))
			}
		}

		return infos
	}
}

func parseGrantedScopes(stdout string) []string {
	m := scopesLineRE.FindStringSubmatch(stdout)
	if len(m) < 2 {
		return nil
	}

	parts := strings.Split(m[1], ",")
	out := make([]string, 0, len(parts))

	for _, p := range parts {
		v := strings.TrimSpace(p)
		v = strings.Trim(v, "'\"")

		if v != "" {
			out = append(out, v)
		}
	}

	return out
}

func containsScope(granted []string, want string) bool {
	return slices.Contains(granted, want)
}
