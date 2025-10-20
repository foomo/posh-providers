package gokazi

import (
	"context"

	"github.com/foomo/gokazi/pkg/gokazi"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/check"
)

func Checker(gk *gokazi.Gokazi) check.Checker {
	return func(ctx context.Context, l log.Logger) []check.Info {
		ls, err := gk.List(ctx)
		if err != nil {
			return []check.Info{{
				Name:   "Gokazi",
				Note:   err.Error(),
				Status: check.StatusFailure,
			}}
		}

		var ret []check.Info
		for key, task := range ls {
			ret = append(ret, check.Info{
				Name: "Task: " + key,
				Note: task.Config.Description,
				Status: func() check.Status {
					if task.Running {
						return check.StatusSuccess
					}

					return check.StatusNote
				}(),
			})
		}

		return ret
	}
}
