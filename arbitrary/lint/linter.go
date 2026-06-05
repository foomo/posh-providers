package lint

import "context"

type Linter interface {
	Name() string
	Lint(ctx context.Context, fix bool) error
}
