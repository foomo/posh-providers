package exec

import (
	"context"
	"os/exec"
	"strings"

	"github.com/foomo/go/slices"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
)

// Complete delegates the completion to gcx's own cobra completion.
func Complete(ctx context.Context, cmd string, offset int, r *readline.Readline) []goprompt.Suggest {
	words := commandArgs(cmd, offset, r)

	return suggests(ctx, cmd, words)
}

// commandArgs returns the gcx command line, i.e. everything after the environment argument.
func commandArgs(cmd string, offset int, r *readline.Readline) []string {
	if r.Args().LenLte(1) {
		return nil
	}

	return slices.Map(r.Args().From(offset), strings.TrimSpace)
}

func suggests(ctx context.Context, cmd string, words []string) []goprompt.Suggest {
	out, err := exec.CommandContext(ctx, cmd, append([]string{"__complete"}, words...)...).Output()
	if err != nil {
		return nil
	}

	return parseSuggests(out)
}

func parseSuggests(out []byte) []goprompt.Suggest {
	var ret []goprompt.Suggest

	for line := range strings.SplitSeq(string(out), "\n") {
		if line == "" || strings.HasPrefix(line, ":") {
			break
		}

		text, description, _ := strings.Cut(line, "\t")
		ret = append(ret, goprompt.Suggest{Text: text, Description: description})
	}

	return ret
}
