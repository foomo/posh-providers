package lint_test

import (
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/arbitrary/lint"
	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/foomo/posh/pkg/command"
	"github.com/foomo/posh/pkg/log"
)

func TestCommandDocs(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	// Linters are discovered from the registry by type assertion; an empty one
	// is enough to describe the tree.
	cmd := lint.NewCommand(log.NewFmt(), command.Commands{})

	testutils.AssertDescribed(t, "lint", cmd)
	testutils.AssertSkilled(t, "lint", cmd)
}
