// Package testutils validates that a provider is documented well enough to be
// useful in `posh agent catalog` and the generated SKILL.md.
//
// Three generated layers carry a provider's documentation, and each has an
// assertion here, called from the provider's own test:
//
//   - AssertDescribed - the command tree. The catalog publishes whatever
//     descriptions the tree carries, so a missing one is not a cosmetic gap: it
//     is a command an agent cannot tell apart from its siblings.
//   - AssertSkilled - the SKILL.md fragment. It becomes the command's own
//     generated skill, loaded in full before an agent knows the command is
//     relevant, so its heading depth and section vocabulary have to hold and it
//     has to stay under a word ceiling. It also has to be written against the
//     name the command is registered under rather than the provider's default,
//     since a project can register it as something else entirely.
//   - AssertDocumentedConfig - the generated config.schema.json. The prose
//     defers to the schema for field shape, so an undocumented property is a
//     gap nothing else covers.
package testutils
