package az

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/check"
	"github.com/foomo/posh/pkg/shell"
)

func AuthChecker(ctx context.Context, l log.Logger) []check.Info {
	name := "Azure"
	out, err := shell.New(ctx, l, "az", "account", "list", "--output", "json").Quiet().CombinedOutput()
	if err != nil {
		return []check.Info{check.NewFailureInfo(name, "Error: "+err.Error())}
	} else if strings.Contains(string(out), "az login") {
		return []check.Info{check.NewNoteInfo(name, "Unauthenticated")}
	}

	var res []map[string]any
	note := "Authenticated"
	if err := json.Unmarshal(out, &res); err == nil {
		if len(res) > 0 && res[0]["user"] != nil {
			if user, ok := res[0]["user"].(map[string]any); ok {
				note += fmt.Sprintf(" as %s: %s", user["type"], user["name"])
			}
		}
	}
	return []check.Info{check.NewSuccessInfo(name, note)}
}
