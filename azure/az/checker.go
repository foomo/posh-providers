package az

import (
	"context"
	"fmt"
	"strings"

	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/check"
	"github.com/foomo/posh/pkg/shell"
)

func AuthChecker(ctx context.Context, l log.Logger) []check.Info {
	name := "Azure"
	out, err := shell.New(ctx, l, "az", "account", "list", "--output", "none").Quiet().CombinedOutput()
	if err != nil {
		return []check.Info{check.NewFailureInfo(name, "Error: "+err.Error())}
	} else if strings.Contains(string(out), "az login") {
		return []check.Info{check.NewNoteInfo(name, "Unauthenticated")}
	}
	return []check.Info{check.NewSuccessInfo(name, fmt.Sprintf("Authenticated"))}
}
