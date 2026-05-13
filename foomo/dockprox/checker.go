package dockprox

import (
	"context"
	"log/slog"
	"os"

	gokaziconfig "github.com/foomo/gokazi/pkg/config"
	"github.com/foomo/gokazi/pkg/gokazi"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/check"
)

func Checker() check.Checker {
	gk := gokazi.New(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	gk.Add("dockprox", gokaziconfig.Task{
		Name: "dockprox",
	})

	return func(ctx context.Context, l log.Logger) []check.Info {
		name := "Dockprox"

		t, err := gk.Find(ctx, "dockprox")
		if err != nil {
			return []check.Info{check.NewFailureInfo("⚠", name, err.Error())}
		}

		if t.Running {
			return []check.Info{check.NewSuccessInfo("⚓︎", name, "Running")}
		}

		return []check.Info{check.NewNoteInfo("⚓︎", name, "Stopped")}
	}
}
