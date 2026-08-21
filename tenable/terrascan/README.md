# terrascan

[terrascan](https://github.com/tenable/terrascan) provider for [posh](https://github.com/foomo/posh).

Exposes a `*Command` that:
- runs as a standalone posh command with `helm`, `terraform`, and `docker` subcommands (each accepts optional repeatable path args with completion)
- satisfies `pkg.Linter` so it can be registered with `arbitrary/lint` (runs all three modes sequentially)

## Usage

```go
import (
	"github.com/foomo/posh-providers/arbitrary/lint"
	"github.com/foomo/posh-providers/tenable/terrascan"
)

// terrascan implements Lint(ctx, fix), so registering it makes `lint` pick it
// up - lint takes the whole command registry rather than a linters option.
inst.commands.Add(terrascan.NewCommand(l, cache))
inst.commands.Add(lint.NewCommand(l, inst.commands))
```

Invocation:

```
terrascan helm [path...]
terrascan terraform [path...]
terrascan docker [path...]
```

Paths default to every directory containing the mode's marker file (`Chart.yaml` / `main.tf` / `Dockerfile`).
