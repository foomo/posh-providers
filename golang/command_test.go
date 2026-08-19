package golang_test

import (
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/golang"
	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
)

func TestCommandDocs(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	// Not parallel: the sibling providers' TestCommandDocs reach viper, which is a singleton.
	cmd := golang.NewCommand(log.NewFmt(), &cache.MemoryCache{})

	testutils.AssertDescribed(t, "go", cmd)
	testutils.AssertSkilled(t, "go", cmd)
}
