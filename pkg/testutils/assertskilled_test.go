package testutils_test

import (
	"context"
	"strings"
	"testing"

	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/stretchr/testify/assert"
)

func TestInvalidSkill(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		expect []string
	}{
		{
			name:   "a well formed fragment is clean",
			input:  "#### Hazards\n\nMutates a live cluster.\n\n#### References\n\n- https://example.com\n",
			expect: nil,
		},
		{
			name:   "a level 5 subsection is allowed",
			input:  "#### Hazards\n\nProse.\n\n##### Detail\n\nMore prose.\n",
			expect: nil,
		},
		{
			// The failure this whole check exists for: level 3 is the generated
			// command heading, so a level 3 here reparents every later command.
			name:  "a level 3 heading is reported",
			input: "### Hazards\n\nProse.\n",
			expect: []string{
				`heading "### Hazards" is level 3; level 3 is the command heading itself, so use level 4 or deeper`,
				`has no "Hazards" section; name the verbs that mutate state and which of them need manual approval, or say plainly that every verb is read-only`,
			},
		},
		{
			name:  "an unknown heading is reported",
			input: "#### Gotchas\n\nProse.\n",
			expect: []string{
				`heading "Gotchas" is not one of Hazards, Behaviour, Configuration, Examples, References`,
				`has no "Hazards" section; name the verbs that mutate state and which of them need manual approval, or say plainly that every verb is read-only`,
			},
		},
		{
			name:  "an out of order heading is reported",
			input: "#### References\n\n- https://example.com\n\n#### Hazards\n\nProse.\n",
			expect: []string{
				`heading "Hazards" is out of order; expected the order Hazards, Behaviour, Configuration, Examples, References`,
			},
		},
		{
			name:   "a Behaviour section is allowed between Hazards and Configuration",
			input:  "#### Hazards\n\nRead-only.\n\n#### Behaviour\n\nhostPort means two different things.\n\n#### Configuration\n\nKey `x`.\n",
			expect: nil,
		},
		{
			// Behaviour carries the warts, so it must not stand in for the
			// verdict Hazards exists to force.
			name:  "a Behaviour section does not satisfy the Hazards requirement",
			input: "#### Behaviour\n\nSurprising but harmless.\n",
			expect: []string{
				`has no "Hazards" section; name the verbs that mutate state and which of them need manual approval, or say plainly that every verb is read-only`,
			},
		},
		{
			name:  "Behaviour before Hazards is out of order",
			input: "#### Behaviour\n\nProse.\n\n#### Hazards\n\nProse.\n",
			expect: []string{
				`heading "Hazards" is out of order; expected the order Hazards, Behaviour, Configuration, Examples, References`,
			},
		},
		{
			// A fragment may omit any section but this one: the blanket
			// disclaimer in the generated preamble says nothing about which of
			// *this* command's verbs are destructive.
			name:  "a missing Hazards section is reported",
			input: "#### References\n\n- https://example.com\n",
			expect: []string{
				`has no "Hazards" section; name the verbs that mutate state and which of them need manual approval, or say plainly that every verb is read-only`,
			},
		},
		{
			// A shell comment inside a fenced example must not be read as a
			// heading, nor a markdown example of the heading rules.
			name:   "headings inside a code fence are content",
			input:  "#### Hazards\n\nRead-only.\n\n#### Examples\n\n```bash\n# not a heading\nposh execute {{cmd}} list\n```\n",
			expect: nil,
		},
		{
			// A Hazards heading inside a fence is content, not the required
			// section.
			name:  "a fenced Hazards heading does not satisfy the requirement",
			input: "#### Examples\n\n```markdown\n#### Hazards\n```\n",
			expect: []string{
				`has no "Hazards" section; name the verbs that mutate state and which of them need manual approval, or say plainly that every verb is read-only`,
			},
		},
		{
			name:  "an unclosed fence is reported",
			input: "#### Hazards\n\nProse.\n\n```bash\nposh execute {{cmd}} list\n",
			expect: []string{
				"has an unclosed code fence",
			},
		},
		{
			// A fragment is optional: posh generates a skill file only for a
			// command with prose or subcommands, so contributing nothing leaves
			// the command covered by the root index.
			name:   "empty markdown is valid",
			input:  "   \n",
			expect: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expect, testutils.InvalidSkill(tt.input, "bruno"))
		})
	}
}

func TestInvalidSkill_wordCeiling(t *testing.T) {
	t.Parallel()

	// "#### Hazards" is two of the counted words, so the body makes up the rest.
	at := "#### Hazards\n\n" + strings.Repeat("word ", testutils.MaxSkillWords-2)
	assert.Empty(t, testutils.InvalidSkill(at, "bruno"), "a fragment at the ceiling is allowed")

	over := testutils.InvalidSkill(at+"straw", "bruno")
	assert.Len(t, over, 1)
	assert.Contains(t, over[0], "word ceiling")
}

func TestInvalidSkill_bareName(t *testing.T) {
	t.Parallel()

	// Each case carries a Hazards section so the only thing under test is the
	// bare-name check.
	const hazards = "#### Hazards\n\nRead-only.\n\n"

	tests := []struct {
		name   string
		input  string
		expect bool
	}{
		// The documented squadron bug, in both the forms it shipped in.
		{
			name:   "a fenced posh execute example naming the default is reported",
			input:  "#### Examples\n\n```bash\nposh execute squadron prod default all status\n```\n",
			expect: true,
		},
		{
			name:   "prose telling the agent to run the default is reported",
			input:  "Run squadron <cluster> <fleet> list to see the units.\n",
			expect: true,
		},
		{
			name:   "the substituted placeholder is clean",
			input:  "Run {{cmd}} <cluster> <fleet> list to see the units.\n",
			expect: false,
		},
		// The name genuinely belongs to something that does not move with the
		// registration, so these must not be flagged.
		{
			name:   "the config key in a code span is allowed",
			input:  "#### Configuration\n\nThe `squadron` key of this project's posh config.\n",
			expect: false,
		},
		{
			// A code span is only a literal when it is a single token. With
			// arguments after it, it is an invocation and must be substituted -
			// this is the form that slipped through an earlier version.
			name:   "an invocation inside a code span is reported",
			input:  "Run `squadron list` first; what it prints is what run accepts.\n",
			expect: true,
		},
		{
			name:   "a flag named after the tool is allowed",
			input:  "#### Behaviour\n\nThe `--squadron` flag is unrelated.\n",
			expect: false,
		},
		{
			name:   "a filename named after the tool is allowed",
			input:  "#### Behaviour\n\nMerges `squadron.yaml` then `squadron.<fleet>.yaml`.\n",
			expect: false,
		},
		{
			name:   "an upstream URL containing the name is allowed",
			input:  "#### References\n\n- https://github.com/foomo/squadron for the format\n",
			expect: false,
		},
		{
			// Forcing this to the registered name would produce "[admiral
			// documentation]" pointing at squadron's docs; forcing it to the
			// placeholder degrades it to a generic "[upstream documentation]".
			name:   "markdown link text naming the upstream project is allowed",
			input:  "#### References\n\n- [squadron documentation](https://example.com)\n",
			expect: false,
		},
		{
			name:   "a path segment named after the tool is allowed",
			input:  "Run it from `./svc/squadron` in the checkout.\n",
			expect: false,
		},
		{
			name:   "an unbackticked path segment is allowed",
			input:  "#### References\n\n- foomo/squadron/README.md for plugin wiring\n",
			expect: false,
		},
		// Whole-word matching, so a substring collision is not a false positive.
		{
			name:   "a longer word containing the name is not reported",
			input:  "Runs squadrons-of-things.\n",
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := testutils.InvalidSkill(hazards+tt.input, "squadron")

			if !tt.expect {
				assert.Empty(t, got)

				return
			}

			assert.Len(t, got, 1)
			assert.Contains(t, got[0], "outside a code span")
		})
	}
}

// skiller substitutes the registered name the way a provider does, so the
// assertion is exercised through the same mechanism it is meant to enforce.
type skiller struct{ skill string }

func (s skiller) Skill(ctx context.Context, name string) string {
	return strings.ReplaceAll(s.skill, "{{cmd}}", name)
}

func TestAssertSkilled(t *testing.T) {
	t.Parallel()

	// A valid Skiller passes on the real *testing.T.
	testutils.AssertSkilled(t, "ok", skiller{skill: "#### Hazards\n\nFine.\n"})

	// So does a provider contributing no fragment at all - the command is then
	// covered by the root skill's index alone.
	testutils.AssertSkilled(t, "unskilled", struct{}{})
	testutils.AssertSkilled(t, "empty", skiller{skill: ""})

	// A Skiller whose markdown would corrupt the generated file fails. Verified
	// against a throwaway T so the failure does not fail this test.
	fake := &testing.T{}
	testutils.AssertSkilled(fake, "shallow", skiller{skill: "### Hazards\n\nProse.\n"})
	assert.True(t, fake.Failed(), "expected a level 3 heading to fail the assertion")

	// So does one that never says what the command does to live state.
	fake = &testing.T{}
	testutils.AssertSkilled(fake, "unhazarded", skiller{skill: "#### References\n\n- https://example.com\n"})
	assert.True(t, fake.Failed(), "expected a missing Hazards section to fail the assertion")

	// And so does one hardcoding its own default rather than substituting, which
	// is the failure that ships prose naming a command the project may not have.
	fake = &testing.T{}
	testutils.AssertSkilled(fake, "bruno", skiller{skill: "#### Hazards\n\nRun bruno list first.\n"})
	assert.True(t, fake.Failed(), "expected a hardcoded default name to fail the assertion")

	// The same fragment substituting the registered name passes. AssertSkilled
	// renders it under SkillProbeName, so this also proves the substitution
	// survives being registered under a name other than the default.
	testutils.AssertSkilled(t, "bruno", skiller{skill: "#### Hazards\n\nRun {{cmd}} list first.\n"})
}
