package yarn_test

import (
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/foomo/posh-providers/yarnpkg/yarn"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
)

func TestCommandDocs(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	// Not parallel: the sibling providers' TestCommandDocs reach viper, which is a singleton.
	cmd := yarn.NewCommand(log.NewFmt(), &cache.MemoryCache{})

	testutils.AssertDescribed(t, "yarn", cmd)
	testutils.AssertSkilled(t, "yarn", cmd)
}
