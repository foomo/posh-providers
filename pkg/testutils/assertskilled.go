package testutils

import (
	"fmt"
	"regexp"
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

// MaxSkillWords is the per-fragment word ceiling.
//
// Every fragment is one file an agent loads in full, and the tokens are spent
// before it knows whether the command is relevant. The ceiling is what keeps a
// fragment to the part that is not recoverable from upstream docs: posh renders
// no arguments or flags anywhere, pointing at `posh agent catalog` and
// `posh help` instead, so a fragment enumerating them is duplication paid for on
// every load.
//
// It is a ceiling rather than a target. A genuinely dangerous provider is
// allowed to use it; compressing real warnings to hit a number is the failure
// this is meant to prevent, and the answer for a fragment that cannot fit is
// usually that it is restating structure rather than that it needs more room.
const MaxSkillWords = 400

// SkillProbeName is the registered name AssertSkilled renders a fragment under.
//
// It stands in for the real case the substitution exists for: a provider
// registered under something other than its default, like the squadron provider
// registered as `admiral`. Rendering under a name no provider defaults to is
// what lets the assertion distinguish a fragment that substituted from one that
// hardcoded, and it is deliberately not a plausible command name so it cannot
// collide with prose.
const SkillProbeName = "posh-probe-name"

// AssertSkilled asserts that the markdown v contributes under name is valid in
// the position it is rendered into.
//
// name is the provider's default command name. The fragment is deliberately
// rendered under a different one - SkillProbeName - because that is the case
// that catches a hardcoded default: a fragment substituting correctly comes back
// naming the probe, while one that hardcoded its own name comes back naming
// that, and the check can tell them apart. Rendering under the default would
// make the two indistinguishable.
//
// A provider that contributes nothing passes. posh generates a skill file only
// for a command that either has prose or has subcommands, so a fragment is
// optional - and a thin pass-through with no project-specific hazard is better
// with none, since the root index already carries its name and description.
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

	// Not implementing Skiller is a valid outcome, not a gap: the command still
	// appears in the root skill's index and still describes itself fully to
	// `posh agent catalog`. There is nothing to check.
	skiller, ok := v.(command.Skiller)
	if !ok {
		return
	}

	if values := InvalidSkill(skiller.Skill(t.Context(), SkillProbeName), name); len(values) > 0 {
		t.Errorf("%s has an invalid SKILL.md:\n  %s", name, strings.Join(values, "\n  "))
	}
}

// InvalidSkill reports every structural problem in a provider's contributed
// SKILL.md markdown, in the order encountered. It returns nil when the markdown
// is valid.
//
// def is the provider's *default* command name, which the fragment must not
// hardcode - see the bare-name check below. It is deliberately not the
// registered name: after substitution the registered name is exactly what a
// correct fragment contains, so checking for that would flag every fragment that
// got it right.
//
// The checks are the ones that make the difference between a fragment that
// renders correctly and one that corrupts the generated file, plus the two that
// keep it worth loading at all. plugin.RenderCommandSkill writes the markdown
// verbatim under a level 3 command heading, so a heading of level 3 or shallower
// escapes its own section and silently reparents every command rendered after
// it.
//
// An empty fragment is valid. posh renders a skill file only for a command that
// has prose or subcommands, so contributing nothing means the command is covered
// by the root index alone - the right outcome for a thin pass-through with no
// project-specific hazard.
func InvalidSkill(skill, def string) []string {
	var ret []string

	if strings.TrimSpace(skill) == "" {
		return nil
	}

	if words := len(strings.Fields(skill)); words > MaxSkillWords {
		ret = append(ret, fmt.Sprintf(
			"is %d words, over the %d word ceiling; cut anything derivable from the command tree or --help, and keep the hazards",
			words, MaxSkillWords))
	}

	ret = append(ret, bareNames(skill, def)...)

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

// bareNames reports occurrences of the provider's default command name that
// should have been the registered name instead.
//
// The fragments are embedded static markdown, so a Skiller substitutes the
// registered name into a placeholder. posh deliberately does not substitute for
// you: a provider that forgets still compiles, and ships prose telling an agent
// to run a command this project does not have - foomo/squadron registered as
// `admiral` documented `posh execute squadron ...` and a Behaviour section
// saying to run `squadron <cluster> <fleet> <squadron> list`. Both are wrong and
// neither breaks the build. This check is what makes the substitution
// enforceable rather than aspirational.
//
// Four things are exempt, because the name genuinely belongs to something other
// than the posh command:
//
//   - A code span holding the name *alone*, or as part of a longer token. The
//     config key, the upstream binary and filenames named after the tool are
//     real literals that must not move with the registration - `squadron.yaml`
//     is `squadron.yaml` whatever the command is called, and the `squadron` key
//     stays the `squadron` key. Wrapping them in backticks is already this
//     repo's convention for config keys and paths.
//   - URLs, including inside link targets, where the name is part of an upstream
//     address.
//   - Markdown link text, which names the upstream project or its docs. Forcing
//     it to the registered name would produce "[admiral documentation]" pointing
//     at squadron's docs; forcing it to a placeholder degrades it to a generic
//     "[upstream documentation]". Both are worse than the literal.
//   - A path segment - anything preceded by "/". A directory or file named after
//     the tool (`./svc/squadron`, `foomo/squadron/README.md`) is on disk under
//     that name whatever the command is called.
//
// A code span holding the name *followed by arguments* is deliberately not
// exempt: `squadron prod default all status` is an invocation, and the whole
// point of the substitution is that such a line names the registered command.
// Exempting every code span would let the most common form of the bug through.
//
// Fenced blocks are deliberately *not* exempt: the `posh execute <name>`
// examples live in fences, and they are the most direct way a fragment tells an
// agent to run the wrong thing.
func bareNames(skill, def string) []string {
	if def == "" {
		return nil
	}

	var (
		ret []string
		re  = wordRe(def)
	)

	for i, line := range strings.Split(skill, "\n") {
		if stripped := stripLiterals(line); re.MatchString(stripped) {
			ret = append(ret, fmt.Sprintf(
				"names %q on line %d outside a code span (%q); the fragment must use the registered name, so substitute it in Skill rather than hardcoding the default",
				def, i+1, strings.TrimSpace(line)))
		}
	}

	return ret
}

// stripLiterals blanks out the parts of a line where the tool's own name is a
// literal rather than the posh command: URLs, markdown link text, path segments,
// and code spans that do not look like an invocation.
//
// A span is treated as a literal only when it holds a single token - `squadron`,
// `squadron.yaml`, `--squadron`. A span with whitespace in it is a command line,
// so it is left in place for the check to see.
func stripLiterals(line string) string {
	line = urlRe.ReplaceAllString(line, " ")
	line = linkTextRe.ReplaceAllString(line, " ")
	line = pathSegmentRe.ReplaceAllString(line, " ")

	return codeSpanRe.ReplaceAllStringFunc(line, func(span string) string {
		if strings.ContainsAny(strings.Trim(span, "`"), " \t") {
			return span
		}

		return " "
	})
}

// wordRe matches name only as a whole word, so a fragment naming `gcloud` is not
// flagged by a provider called `cloud`, and vice versa. Hyphens count as word
// characters, so `gherkin-lint` does not match a fragment's `lint`.
func wordRe(name string) *regexp.Regexp {
	return regexp.MustCompile(`(^|[^\w-])` + regexp.QuoteMeta(name) + `($|[^\w-])`)
}

var (
	codeSpanRe = regexp.MustCompile("`[^`]*`")
	urlRe      = regexp.MustCompile(`https?://\S+`)
	// Markdown link text, so "[nova documentation](...)" keeps naming the
	// upstream project rather than being forced to the registered name.
	linkTextRe = regexp.MustCompile(`\[[^\]]*\]`)
	// A path segment: the name preceded by a "/" is a real directory or file on
	// disk, which does not move with the registration.
	pathSegmentRe = regexp.MustCompile(`/[\w.@/-]+`)
)
