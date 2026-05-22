package az

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/check"
	"github.com/foomo/posh/pkg/shell"
)

func AuthChecker(ctx context.Context, l log.Logger) []check.Info {
	name := "Azure"

	out, err := shell.New(ctx, l, "az", "account", "list", "--output", "json").CombinedOutput()
	if err != nil {
		return []check.Info{check.NewNoteInfo("⛹", name, "Unauthorized")}
	}

	var note string

	var res []map[string]any
	if err := json.Unmarshal(out, &res); err != nil || len(res) == 0 {
		return []check.Info{check.NewNoteInfo("⛹", name, "Unauthorized")}
	}

	if res[0]["user"] != nil {
		if user, ok := res[0]["user"].(map[string]any); ok {
			note = fmt.Sprintf("%s (%s)", user["name"], user["type"])
		}
	}

	return []check.Info{check.NewSuccessInfo("⛹", name, note)}
}
