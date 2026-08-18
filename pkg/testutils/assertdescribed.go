// Package agentdoc validates that a provider's command tree is documented well
// enough to be useful in `posh agent catalog` and the generated SKILL.md.
//
// The catalog publishes whatever descriptions the tree carries, so a missing
// one is not a cosmetic gap: it is a command an agent cannot tell apart from
// its siblings. This package is the regression guard, called from a provider's
// own test.
package testutils

import (
	"fmt"
	"strings"
	"testing"

	"github.com/foomo/posh/pkg/command"
)

// AssertDescribed asserts that v implements command.Describer and that every
// node and argument it describes carries a description.
//
// It is the one call a provider's test needs:
//
//	func TestCommandDocs(t *testing.T) {
//		cmd, err := bruno.NewCommand(log.NewFmt())
//		require.NoError(t, err)
//		agentdoc.AssertDescribed(t, "bruno", cmd)
//	}
func AssertDescribed(t *testing.T, name string, v any) {
	t.Helper()

	describer, ok := v.(command.Describer)
	if !ok {
		t.Errorf("%s does not implement command.Describer, so `posh agent catalog` cannot describe it", name)

		return
	}

	if values := UndocumentedDescribed(describer.Describe(t.Context())); len(values) > 0 {
		t.Errorf("%s has undocumented commands:\n  %s", name, strings.Join(values, "\n  "))
	}
}

// UndocumentedDescribed walks a described command tree and reports every node and
// argument missing a description, in depth-first order. It returns nil when the
// tree is fully documented.
//
// Emptiness is tested after trimming whitespace: a description of "" or " " is
// exactly as useless to an agent as an absent one, and both occur in this repo.
func UndocumentedDescribed(info command.CommandInfo) []string {
	var ret []string

	if strings.TrimSpace(info.Description) == "" {
		ret = append(ret, fmt.Sprintf("node %q has no description", info.FullPath))
	}

	for _, arg := range info.Arguments {
		if strings.TrimSpace(arg.Description) == "" {
			ret = append(ret, fmt.Sprintf("node %q argument %q has no description", info.FullPath, arg.Name))
		}
	}

	for _, value := range info.Subcommands {
		ret = append(ret, UndocumentedDescribed(value)...)
	}

	return ret
}
