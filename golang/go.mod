module github.com/foomo/posh-providers/golang

go 1.26.3

replace (
	github.com/c-bata/go-prompt v0.2.6 => github.com/franklinkim/go-prompt v0.2.7-0.20210427061716-a8f4995d7aa5
	github.com/foomo/posh-providers => ../
	github.com/pkg/term => github.com/pkg/term v1.1.0
)

require (
	github.com/c-bata/go-prompt v0.2.6
	github.com/foomo/posh v0.19.0
	golang.org/x/sync v0.20.0
	k8s.io/utils v0.0.0-20260507154919-ff6756f316d2
)

require (
	atomicgo.dev/cursor v0.2.0 // indirect
	atomicgo.dev/keyboard v0.2.10 // indirect
	atomicgo.dev/schedule v0.1.0 // indirect
	github.com/charlievieth/fastwalk v1.0.14 // indirect
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/containerd/console v1.0.5 // indirect
	github.com/gookit/color v1.6.1 // indirect
	github.com/lithammer/fuzzysearch v1.1.8 // indirect
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.22 // indirect
	github.com/mattn/go-runewidth v0.0.24 // indirect
	github.com/mattn/go-tty v0.0.8 // indirect
	github.com/neilotoole/slogt v1.1.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pkg/term v1.1.0 // indirect
	github.com/pterm/pterm v0.12.83 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	golang.org/x/sys v0.45.0 // indirect
	golang.org/x/term v0.43.0 // indirect
	golang.org/x/text v0.37.0 // indirect
)
