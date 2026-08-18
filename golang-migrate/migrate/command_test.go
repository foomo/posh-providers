package migrate_test

import (
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/golang-migrate/migrate"
	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/foomo/posh/pkg/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestCommandDocs(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	viper.Set("migrate", map[string]any{
		"databases": map[string]string{"local": "pgx5://localhost:5432/db"},
		"sources":   map[string]string{"default": "file://migrations"},
	})
	t.Cleanup(viper.Reset)

	cmd, err := migrate.NewCommand(log.NewFmt())
	require.NoError(t, err)

	testutils.AssertDescribed(t, "migrate", cmd)
	testutils.AssertSkilled(t, "migrate", cmd)
}
