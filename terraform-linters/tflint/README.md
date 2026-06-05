# tflint

[tflint](https://github.com/terraform-linters/tflint) provider for [posh](https://github.com/foomo/posh).

Exposes a `*Command` that:
- runs as a standalone posh command (`tflint [path...] [--fix]`)
- satisfies `pkg.Linter` so it can be registered with `arbitrary/lint`

Accepts an optional repeatable `path` arg with completion over every directory containing a `main.tf` file. Defaults to all such directories.

## Usage

```go
import (
	"github.com/foomo/posh-providers/arbitrary/lint"
	"github.com/foomo/posh-providers/terraform-linters/tflint"
)

cmd := tflint.NewCommand(l, cache)
lintCmd := lint.NewCommand(l, commands)
```
