package squadron_test

import (
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/foomo/squadron/v2"
	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
)

func TestCommandDocs(t *testing.T) {
	t.Parallel()
	testingx.Tags(t, tagx.Short)

	// nil Squadron and Kubectl are safe here: every deref of them happens
	// inside a Values/Suggest/Flags closure, and tree.Root.Describe never
	// invokes those - it emits dynamic nodes as placeholders. The cache is
	// used directly in the constructor body, so it must be real. Verified
	// working.
	cmd := squadron.NewCommand(log.NewFmt(), nil, nil, &cache.MemoryCache{})

	testutils.AssertDescribed(t, "squadron", cmd)
	testutils.AssertSkilled(t, "squadron", cmd)
}
