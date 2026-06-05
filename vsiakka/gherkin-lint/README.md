# gherkin-lint

[gherkin-lint](https://github.com/vsiakka/gherkin-lint) provider for [posh](https://github.com/foomo/posh).

Exposes a `*Command` that:
- runs as a standalone posh command (`gherkin-lint [path...]`)
- satisfies `pkg.Linter` so it can be registered with `arbitrary/lint`

Accepts an optional repeatable `path` arg with completion over every directory containing a `wdio.conf.ts` file. Defaults to all such directories.

## Usage

```go
import (
	"github.com/foomo/posh-providers/arbitrary/lint"
	gherkinlint "github.com/foomo/posh-providers/vsiakka/gherkin-lint"
)

cmd := gherkinlint.NewCommand(l, cache)
lintCmd := lint.NewCommand(l, commands)
```
