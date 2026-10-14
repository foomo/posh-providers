package lint

import "context"

// Linter is implemented by posh commands that can be discovered and run by
// the lint command.
type Linter interface {
	Name() string
	Lint(ctx context.Context, fix bool) error
}
