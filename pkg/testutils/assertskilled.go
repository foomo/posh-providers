package testutils

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/foomo/posh/pkg/command"
)

// SkillHeadings are the section headings a provider's SKILL.md may use, in the
// order they should appear.
//
// The vocabulary is fixed rather than free-form because the fragments are
// spliced into one project-wide SKILL.md alongside every other command: a
// consistent set of headings is what lets an agent skim it. "Hazards" carries
// the side effects and agent-hostile behaviour, so it comes first.
//
// It is named for its payload rather than "Notes": a generic bucket invites
// restating the command tree, which renders directly above it, while a sentence
// that is not a hazard looks out of place under this heading.
//
// "Behaviour" is the pressure valve that keeps that true. A provider often
// surprises without endangering - a field meaning two different things depending
// on the subtree, a defect that only corrupts displayed metadata - and filing
// those under "Hazards" dilutes the section an agent consults to decide what
// needs approval. The test is consequence, not severity: if acting on it wrongly
// destroys state, blocks unattended, or reaches further than the command name
// suggests, it is a hazard; if it merely defeats an expectation, it is
// behaviour.
var SkillHeadings = []string{"Hazards", "Behaviour", "Configuration", "Examples", "References"}

// RequiredSkillHeading must appear in every fragment.
//
// RenderSkill prefixes the generated file with a blanket disclaimer that posh
// enforces no access control and that a listed command is not necessarily safe
// to run unattended. That text applies equally to a read-only list command and
// to one that drops a database, so it carries no signal on its own: the
// per-command fragment is the only place that can say which verbs actually
// destroy state and which need a human to approve them. A fragment omitting it
// silently asserts the command has no side effects worth naming.
const RequiredSkillHeading = "Hazards"

// AssertSkilled asserts that v implements command.Skiller and that the markdown
// it contributes is valid in the position it is rendered into.
//
// It pairs with AssertDescribed in a provider's test:
//
//	func TestCommandDocs(t *testing.T) {
//		cmd, err := bruno.NewCommand(log.NewFmt())
//		require.NoError(t, err)
//		testutils.AssertDescribed(t, "bruno", cmd)
//		testutils.AssertSkilled(t, "bruno", cmd)
//	}
func AssertSkilled(t *testing.T, name string, v any) {
	t.Helper()

	skiller, ok := v.(command.Skiller)
	if !ok {
		t.Errorf("%s does not implement command.Skiller, so it contributes nothing to SKILL.md", name)

		return
	}

	if values := InvalidSkill(skiller.Skill(t.Context())); len(values) > 0 {
		t.Errorf("%s has an invalid SKILL.md:\n  %s", name, strings.Join(values, "\n  "))
	}
}

// InvalidSkill reports every structural problem in a provider's contributed
// SKILL.md markdown, in the order encountered. It returns nil when the markdown
// is valid.
//
// The checks are the ones that make the difference between a fragment that
// renders correctly and one that corrupts the generated file. plugin.RenderSkill
// writes the markdown verbatim under a level 3 command heading, so a heading of
// level 3 or shallower escapes its own section and silently reparents every
// command rendered after it.
func InvalidSkill(skill string) []string {
	var ret []string

	if strings.TrimSpace(skill) == "" {
		return []string{"contributes no markdown, so the command may be omitted from SKILL.md entirely"}
	}

	var (
		fenced   bool
		seen     int
		required bool
	)

	for line := range strings.SplitSeq(skill, "\n") {
		trimmed := strings.TrimSpace(line)

		// Headings inside a fenced block are content - this file documents the
		// heading rules using the very headings it checks for.
		if strings.HasPrefix(trimmed, "```") {
			fenced = !fenced

			continue
		}

		if fenced || !strings.HasPrefix(trimmed, "#") {
			continue
		}

		hashes := len(trimmed) - len(strings.TrimLeft(trimmed, "#"))
		if hashes < 4 {
			ret = append(ret, fmt.Sprintf("heading %q is level %d; level 3 is the command heading itself, so use level 4 or deeper", trimmed, hashes))

			continue
		}

		// Only top level sections are constrained; a fragment is free to
		// subdivide one with a level 5 heading.
		if hashes > 4 {
			continue
		}

		title := strings.TrimSpace(strings.TrimLeft(trimmed, "#"))

		if title == RequiredSkillHeading {
			required = true
		}

		index := slices.Index(SkillHeadings, title)
		if index < 0 {
			ret = append(ret, fmt.Sprintf("heading %q is not one of %s", title, strings.Join(SkillHeadings, ", ")))

			continue
		}

		if index < seen {
			ret = append(ret, fmt.Sprintf("heading %q is out of order; expected the order %s", title, strings.Join(SkillHeadings, ", ")))
		}

		seen = index
	}

	if fenced {
		ret = append(ret, "has an unclosed code fence")
	}

	if !required {
		ret = append(ret, fmt.Sprintf("has no %q section; name the verbs that mutate state and which of them need manual approval, or say plainly that every verb is read-only", RequiredSkillHeading))
	}

	return ret
}
