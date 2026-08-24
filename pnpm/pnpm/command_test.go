package pnpm_test

import (
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/foomo/posh-providers/pnpm/pnpm"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
)

func TestCommandDocs(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	cmd := pnpm.NewCommand(log.NewFmt(), &cache.MemoryCache{})

	testutils.AssertDescribed(t, "pnpm", cmd)
	testutils.AssertSkilled(t, "pnpm", cmd)
}
