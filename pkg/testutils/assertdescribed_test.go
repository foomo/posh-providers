package testutils_test

import (
	"context"
	"testing"

	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/foomo/posh/pkg/command"
	"github.com/stretchr/testify/assert"
)

func TestUndocumentedDescribed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  command.CommandInfo
		expect []string
	}{
		{
			name: "fully documented tree is clean",
			input: command.CommandInfo{
				FullPath:    "bruno",
				Description: "Run Bruno requests",
				Subcommands: []command.CommandInfo{
					{
						FullPath:    "bruno run",
						Description: "Run the Bruno cli",
						Arguments: []command.ArgInfo{
							{Name: "env", Description: "Environment name"},
						},
					},
				},
			},
			expect: nil,
		},
		{
			name: "missing node description is reported",
			input: command.CommandInfo{
				FullPath:    "squadron",
				Description: "Manage your squadron",
				Subcommands: []command.CommandInfo{
					{FullPath: "squadron <cluster>"},
				},
			},
			expect: []string{`node "squadron <cluster>" has no description`},
		},
		{
			// nova and pluto set Description: "" explicitly; a presence check
			// would pass them.
			name: "explicitly blanked description is reported",
			input: command.CommandInfo{
				FullPath:    "nova",
				Description: "",
			},
			expect: []string{`node "nova" has no description`},
		},
		{
			name: "whitespace-only description is reported",
			input: command.CommandInfo{
				FullPath:    "pluto",
				Description: "   ",
			},
			expect: []string{`node "pluto" has no description`},
		},
		{
			name: "missing arg description is reported",
			input: command.CommandInfo{
				FullPath:    "golang",
				Description: "Run go",
				Arguments: []command.ArgInfo{
					{Name: "package"},
				},
			},
			expect: []string{`node "golang" argument "package" has no description`},
		},
		{
			name: "reports every gap, depth first",
			input: command.CommandInfo{
				FullPath:    "root",
				Description: "Root",
				Subcommands: []command.CommandInfo{
					{
						FullPath:  "root a",
						Arguments: []command.ArgInfo{{Name: "x"}},
					},
				},
			},
			expect: []string{
				`node "root a" has no description`,
				`node "root a" argument "x" has no description`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expect, testutils.UndocumentedDescribed(tt.input))
		})
	}
}

type describer struct{ info command.CommandInfo }

func (d describer) Describe(ctx context.Context) command.CommandInfo { return d.info }

func TestAssertDescribed(t *testing.T) {
	t.Parallel()

	// A documented Describer passes on the real *testing.T.
	testutils.AssertDescribed(t, "ok", describer{info: command.CommandInfo{
		FullPath:    "ok",
		Description: "Fine",
	}})

	// A non-Describer fails. Verified against a throwaway T so the failure does
	// not fail this test.
	fake := &testing.T{}
	testutils.AssertDescribed(fake, "nope", struct{}{})
	assert.True(t, fake.Failed(), "expected a non-Describer to fail the assertion")
}
