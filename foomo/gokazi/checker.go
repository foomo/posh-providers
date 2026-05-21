package gokazi

import (
	"context"

	"github.com/foomo/gokazi/pkg/gokazi"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/check"
)

func TasksChecker(gk *gokazi.Gokazi) check.Checker {
	return func(ctx context.Context, l log.Logger) []check.Info {
		ls, err := gk.List(ctx)
		if err != nil {
			return []check.Info{check.NewFailureInfo("⚡︎", "gokazi", err.Error())}
		}

		var ret []check.Info

		for key, task := range ls {
			if task.Running {
				ret = append(ret, check.NewSuccessInfo("◉", key, task.Config.Description))
			} else {
				ret = append(ret, check.NewNoteInfo("○", key, task.Config.Description))
			}
		}

		return ret
	}
}
