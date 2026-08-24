package zeus_test

import (
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/dreadl0ck/zeus"
	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
)

func TestCommandDocs(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	cmd := zeus.NewCommand(log.NewFmt(), &cache.MemoryCache{})

	testutils.AssertDescribed(t, "zeus", cmd)
	testutils.AssertSkilled(t, "zeus", cmd)
}
