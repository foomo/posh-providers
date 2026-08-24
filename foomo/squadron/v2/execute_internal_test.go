package squadron

// This test lives in package squadron rather than squadron_test because the
// behaviour under test - the agent-mode refusal in (*Command).execute - is
// only reachable through a private method, and building the *Squadron it
// dereferences requires the unexported cfg field to carry a cluster with
// Confirm: true. An external test package can reach neither.

import (
	"context"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	squadroncfg "github.com/foomo/posh-providers/foomo/squadron"
	"github.com/foomo/posh-providers/kubernetes/kubectl"
	"github.com/foomo/posh/pkg/agent"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/readline"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newConfirmCommand builds a Command whose "prod" cluster requires
// confirmation, mirroring a project that singled that cluster out for human
// approval.
func newConfirmCommand(t *testing.T) *Command {
	t.Helper()

	l := log.NewFmt()

	// kubectl must be real: execute calls c.kubectl.Cluster(cluster).Env(profile)
	// before reaching the refusal branch. Env only reads cfg.ConfigPath, but the
	// constructor mkdir's it, so it has to be a real path. These tests are not
	// parallel because both viper and KUBECONFIG (set by kubectl.New) are global.
	viper.Set("kubectl", map[string]any{"configPath": t.TempDir()})
	t.Cleanup(viper.Reset)

	k, err := kubectl.New(l, &cache.MemoryCache{})
	require.NoError(t, err)

	sq := &Squadron{
		l:         l,
		configKey: "squadron",
		kubectl:   k,
	}
	sq.cfg.Path = t.TempDir()
	sq.cfg.Clusters = append(sq.cfg.Clusters, squadroncfg.Cluster{
		Name:    "prod",
		Confirm: true,
		Fleets:  []string{"default"},
	})

	return NewCommand(l, sq, k, &cache.MemoryCache{})
}

// newUpReadline returns a Readline positioned exactly as the tree would leave
// it for `squadron prod default myapp up myunit`, with the flag sets the "up"
// node registers so execute's flag lookups resolve.
func newUpReadline(t *testing.T, cmd string) *readline.Readline {
	t.Helper()

	r, err := readline.New(log.NewFmt())
	require.NoError(t, err)
	require.NoError(t, r.Parse("squadron prod default myapp "+cmd+" myunit"))

	// Sanity-check the arg layout execute indexes into: [0]=cluster,
	// [1]=fleet, [2]=squadron, [3]=cmd, [4:]=units.
	require.Len(t, r.Args(), 5)
	require.Equal(t, cmd, r.Args()[3])

	fs := readline.NewFlagSets()
	fs.Internal().Bool("no-override", false, "ignore override files")
	fs.Internal().Bool("build", false, "build image")
	fs.Internal().Bool("create-namespace", false, "create namespace if not exist")
	fs.Internal().String("tag", "", "image tag")
	fs.Internal().String("profile", "", "Profile to use.")
	fs.Internal().StringSlice("push-args", nil, "additional docker push args")
	fs.Internal().StringSlice("build-args", nil, "additional docker buildx build args")
	fs.Default().String("tags", "", "list of tags to include or exclude")
	fs.Default().String("revision", "", "revision number to rollback to")
	r.SetFlagSets(fs)
	require.NoError(t, r.ParseFlagSets())

	return r
}

// TestExecuteRefusesConfirmClusterUnderAgent proves the refusal branch is
// actually taken - not merely compiled - for every destructive verb.
func TestExecuteRefusesConfirmClusterUnderAgent(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	for _, cmd := range []string{"up", "down", "rollback"} {
		t.Run(cmd, func(t *testing.T) {
			// Force agent mode on deterministically. The ambient environment
			// sets CLAUDECODE, so this must not be left to detection.
			defer agent.SetDetected(true)()

			c := newConfirmCommand(t)
			r := newUpReadline(t, cmd)

			err := c.execute(context.Background(), r)

			require.Error(t, err, "%s against a Confirm cluster must be refused under agent mode", cmd)
			assert.Contains(t, err.Error(), "prod", "error should name the cluster")
			assert.Contains(t, err.Error(), "default", "error should name the fleet")
			assert.Contains(t, err.Error(), cmd, "error should name the refused command")
			assert.Contains(t, err.Error(), "interactive")
		})
	}
}

// TestExecuteAgentModeAllowsReadOnlyCluster pins the negative half: the
// refusal is scoped to destructive verbs, so a read-only verb against the same
// Confirm cluster must not be refused by that branch.
func TestExecuteAgentModeAllowsReadOnlyCluster(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	defer agent.SetDetected(true)()

	c := newConfirmCommand(t)
	r := newUpReadline(t, "status")

	// status shells out to the squadron binary, which is absent here, so the
	// call fails - but it must fail past the refusal branch, never with the
	// refusal message.
	if err := c.execute(context.Background(), r); err != nil {
		assert.NotContains(t, err.Error(), "requires interactive confirmation")
	}
}
