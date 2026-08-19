package cdktf

type Config struct {
	// Directory the cdktf app lives in; every verb runs with this as its working
	// directory. Also the root walked for `*.stack.yaml` files to discover stack
	// names.
	Path string `json:"path" yaml:"path"`
	// Directory cdktf synthesizes into. Used only by `unlock`, which runs
	// terraform in `<outPath>/stacks/<stack>`, so leaving it unset breaks that
	// verb while every other one still works.
	OutPath string `json:"outPath" yaml:"outPath"`
}
