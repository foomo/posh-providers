package gherkinlint_test

import (
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/pkg/testutils"
	gherkinlint "github.com/foomo/posh-providers/vsiakka/gherkin-lint"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
)

func TestCommandDocs(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	cmd := gherkinlint.NewCommand(log.NewFmt(), &cache.MemoryCache{})

	testutils.AssertDescribed(t, "gherkin-lint", cmd)
	testutils.AssertSkilled(t, "gherkin-lint", cmd)
}
