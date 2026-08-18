package testutils_test

import (
	"context"
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
			name:  "empty markdown is reported",
			input: "   \n",
			expect: []string{
				"contributes no markdown, so the command may be omitted from SKILL.md entirely",
			},
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
			input:  "#### Hazards\n\nRead-only.\n\n#### Examples\n\n```bash\n# not a heading\nposh execute bruno list\n```\n",
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
			input: "#### Hazards\n\nProse.\n\n```bash\nposh execute bruno list\n",
			expect: []string{
				"has an unclosed code fence",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expect, testutils.InvalidSkill(tt.input))
		})
	}
}

type skiller struct{ skill string }

func (s skiller) Skill(ctx context.Context) string { return s.skill }

func TestAssertSkilled(t *testing.T) {
	t.Parallel()

	// A valid Skiller passes on the real *testing.T.
	testutils.AssertSkilled(t, "ok", skiller{skill: "#### Hazards\n\nFine.\n"})

	// A non-Skiller fails. Verified against a throwaway T so the failure does
	// not fail this test.
	fake := &testing.T{}
	testutils.AssertSkilled(fake, "nope", struct{}{})
	assert.True(t, fake.Failed(), "expected a non-Skiller to fail the assertion")

	// So does a Skiller whose markdown would corrupt the generated file.
	fake = &testing.T{}
	testutils.AssertSkilled(fake, "shallow", skiller{skill: "### Hazards\n\nProse.\n"})
	assert.True(t, fake.Failed(), "expected a level 3 heading to fail the assertion")

	// So does one that never says what the command does to live state.
	fake = &testing.T{}
	testutils.AssertSkilled(fake, "unhazarded", skiller{skill: "#### References\n\n- https://example.com\n"})
	assert.True(t, fake.Failed(), "expected a missing Hazards section to fail the assertion")
}
