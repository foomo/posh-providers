module github.com/foomo/posh-providers/pulumi

go 1.25

replace (
	github.com/c-bata/go-prompt v0.2.6 => github.com/franklinkim/go-prompt v0.2.7-0.20210427061716-a8f4995d7aa5
	github.com/foomo/posh-providers => ../
	github.com/foomo/posh-providers/azure => ../azure
	github.com/foomo/posh-providers/google => ../google
	github.com/foomo/posh-providers/kubernetes => ../kubernetes
	github.com/foomo/posh-providers/onepassword => ../onepassword
	github.com/pkg/term => github.com/pkg/term v1.1.0
)

require (
	github.com/foomo/go v0.0.3
	github.com/foomo/posh v0.15.0
	github.com/foomo/posh-providers/azure v0.0.0-00010101000000-000000000000
	github.com/foomo/posh-providers/google v0.0.0-00010101000000-000000000000
	github.com/foomo/posh-providers/onepassword v0.0.0-00010101000000-000000000000
	github.com/invopop/jsonschema v0.13.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/viper v1.21.0
	github.com/stretchr/testify v1.11.1
)

require (
	atomicgo.dev/cursor v0.2.0 // indirect
	atomicgo.dev/keyboard v0.2.9 // indirect
	atomicgo.dev/schedule v0.1.0 // indirect
	github.com/1Password/connect-sdk-go v1.5.3 // indirect
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/c-bata/go-prompt v0.2.6 // indirect
	github.com/charlievieth/fastwalk v1.0.14 // indirect
	github.com/clipperhouse/uax29/v2 v2.2.0 // indirect
	github.com/containerd/console v1.0.5 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/foomo/posh-providers/kubernetes v0.0.0-00010101000000-000000000000 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/gookit/color v1.6.0 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/lithammer/fuzzysearch v1.1.8 // indirect
	github.com/mailru/easyjson v0.9.1 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.19 // indirect
	github.com/mattn/go-tty v0.0.7 // indirect
	github.com/neilotoole/slogt v1.1.0 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/pkg/term v1.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/pterm/pterm v0.12.82 // indirect
	github.com/sagikazarmark/locafero v0.12.0 // indirect
	github.com/samber/lo v1.52.0 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/uber/jaeger-client-go v2.30.0+incompatible // indirect
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	github.com/wk8/go-ordered-map/v2 v2.1.8 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/term v0.36.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
