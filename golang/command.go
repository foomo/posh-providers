package golang

import (
	"context"
	"path"
	"slices"
	"strings"
	"time"

	prompt2 "github.com/c-bata/go-prompt"
	"github.com/foomo/go/options"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/exec"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/files"
	"github.com/foomo/posh/pkg/util/suggests"
	"golang.org/x/sync/errgroup"
	"k8s.io/utils/env"
)

type Command struct {
	l                log.Logger
	cache            cache.Namespace
	commandTree      tree.Root
	execGolangciLint exec.CommandProvider
}

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func CommandWithExecGolangciLint(v exec.CommandProvider) options.Option[*Command] {
	return func(c *Command) {
		c.execGolangciLint = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, cache cache.Cache, opts ...options.Option[*Command]) *Command {
	inst := &Command{
		l:     l.Named("go"),
		cache: cache.Get("go"),
		execGolangciLint: func(ctx context.Context, args ...string) *exec.Command {
			return exec.NewCommand(ctx, "golangci-lint", args...)
		},
	}

	options.Apply(inst, opts...)

	pathModArg := &tree.Arg{
		Name:     "path",
		Optional: true,
		Suggest: func(ctx context.Context, p tree.Root, r *readline.Readline) []prompt2.Suggest {
			return inst.completePaths(ctx, "go.mod", true)
		},
	}

	pathGenerateArg := &tree.Arg{
		Name:     "path",
		Optional: true,
		Suggest: func(ctx context.Context, p tree.Root, r *readline.Readline) []prompt2.Suggest {
			return inst.completePaths(ctx, "generate.go", false)
		},
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        "go",
		Description: "Go related tasks",
		Nodes: tree.Nodes{
			{
				Name:        "mod",
				Description: "Run go mod commands",
				Nodes: tree.Nodes{
					{
						Name:        "tidy",
						Description: "Run go mod tidy",
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							fs.Internal().Int("parallel", 0, "Number of parallel processes")
							return nil
						},
						Args:    []*tree.Arg{pathModArg},
						Execute: inst.modTidy,
					},
					{
						Name:        "download",
						Description: "Run go mod download",
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							fs.Internal().Int("parallel", 0, "Number of parallel processes")
							return nil
						},
						Args:    []*tree.Arg{pathModArg},
						Execute: inst.modDownload,
					},
					{
						Name:        "outdated",
						Description: "Show go mod outdated",
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							fs.Internal().Int("parallel", 0, "Number of parallel processes")
							return nil
						},
						Args:    []*tree.Arg{pathModArg},
						Execute: inst.modOutdated,
					},
					{
						Name:        "upgrade",
						Description: "Show go mod upgrade",
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							fs.Internal().Int("parallel", 0, "Number of parallel processes")
							return nil
						},
						Args:    []*tree.Arg{pathModArg},
						Execute: inst.modUpgrade,
					},
				},
			},
			{
				Name:        "work",
				Description: "Manage go.work file",
				Nodes: tree.Nodes{
					{
						Name:        "init",
						Description: "Generate go.work file",
						Execute:     inst.workInit,
					},
					{
						Name:        "sync",
						Description: "Sync go.work file",
						Execute:     inst.workSync,
					},
					{
						Name:        "use",
						Description: "Add go.work entry",
						Args: []*tree.Arg{
							{
								Name:    "path",
								Suggest: nil,
							},
						},
						Execute: inst.workUse,
					},
				},
			},
			{
				Name:        "generate",
				Description: "Run go mod commands",
				Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
					fs.Internal().Int("parallel", 0, "Number of parallel processes")
					return nil
				},
				Args:    []*tree.Arg{pathGenerateArg},
				Execute: inst.generate,
			},
			{
				Name:        "clean",
				Description: "Run golangci lint cache clean",
				Nodes: tree.Nodes{
					{
						Name:        "lint",
						Description: "Lint cache",
						Execute:     inst.cleanLintCache,
					},
					{
						Name:        "build",
						Description: "Build cache",
						Execute:     inst.cleanBuildCache,
					},
					{
						Name:        "mod",
						Description: "Mod cache",
						Execute:     inst.cleanModCache,
					},
					{
						Name:        "fuzz",
						Description: "Fuzz cache",
						Execute:     inst.cleanFuzzCache,
					},
					{
						Name:        "test",
						Description: "Test cache",
						Execute:     inst.cleanTestCache,
					},
				},
			},
			{
				Name:        "lint",
				Description: "Run golangci lint",
				Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
					fs.Default().Duration("timeout", 0, "Timeout for total work")
					fs.Default().Bool("fast-only", false, "Run only fast linters from enabled linters set")
					fs.Default().Bool("new", false, "Show only new issues")
					fs.Default().Bool("fix", false, "Fix found issue")
					fs.Default().String("out-format", "", "Formats of output")
					fs.Default().Int("concurrency", 0, "Number of CPUs to use")
					fs.Default().StringSlice("build-tags", []string{"safe"}, "Build tag")

					fs.Internal().Int("parallel", 0, "Number of parallel processes")

					return nil
				},
				Args:    []*tree.Arg{pathModArg},
				Execute: inst.lint,
			},
			{
				Name:        "test",
				Description: "Run go test",
				Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
					fs.Internal().String("tags", "", "Comma separeted string of tags")
					return nil
				},
				Args: []*tree.Arg{
					pathModArg,
					{
						Name:     "package",
						Optional: true,
						Suggest: func(ctx context.Context, p tree.Root, r *readline.Readline) []prompt2.Suggest {
							return inst.completeTestPackages(ctx, r.Args().At(1))
						},
					},
					{
						Name:     "test",
						Optional: true,
						Suggest: func(ctx context.Context, p tree.Root, r *readline.Readline) []prompt2.Suggest {
							return inst.completeTests(ctx, r.Args().At(1), r.Args().At(2))
						},
					},
				},
				Execute: inst.test,
			},
			{
				Name:        "fuzz",
				Description: "Run go test fuzzing",
				Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
					fs.Internal().String("tags", "", "Comma separeted string of tags")
					return nil
				},
				Args: []*tree.Arg{
					pathModArg,
					{
						Name:     "package",
						Optional: true,
						Suggest: func(ctx context.Context, p tree.Root, r *readline.Readline) []prompt2.Suggest {
							return inst.completeTestPackages(ctx, r.Args().At(1))
						},
					},
					{
						Name:     "fuzz",
						Optional: true,
						Suggest: func(ctx context.Context, p tree.Root, r *readline.Readline) []prompt2.Suggest {
							return inst.completeFuzz(ctx, r.Args().At(1), r.Args().At(2))
						},
					},
				},
				Execute: inst.fuzz,
			},
			{
				Name:        "bench",
				Description: "Run go test benchmarks",
				Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
					fs.Internal().String("tags", "", "Comma separeted string of tags")
					return nil
				},
				Args: []*tree.Arg{
					pathModArg,
					{
						Name:     "package",
						Optional: true,
						Suggest: func(ctx context.Context, p tree.Root, r *readline.Readline) []prompt2.Suggest {
							return inst.completeTestPackages(ctx, r.Args().At(1))
						},
					},
					{
						Name:     "bench",
						Optional: true,
						Suggest: func(ctx context.Context, p tree.Root, r *readline.Readline) []prompt2.Suggest {
							return inst.completeBenchmarks(ctx, r.Args().At(1), r.Args().At(2))
						},
					},
				},
				Execute: inst.bench,
			},
			{
				Name:        "build",
				Description: "Run go build",
				Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
					fs.Internal().Int("parallel", 0, "Number of parallel processes")
					fs.Internal().Bool("force", false, "Force rebuilding of up-to-date packages (go build -a)")
					fs.Internal().Bool("verbose", false, "Print the names of packages as they are compiled (go build -v)")
					fs.Internal().Bool("cgo", false, "Enable cgo (CGO_ENABLED=1)")

					return nil
				},
				Args: []*tree.Arg{
					pathModArg,
					{
						Name:     "package",
						Repeat:   true,
						Optional: true,
						Suggest: func(ctx context.Context, p tree.Root, r *readline.Readline) []prompt2.Suggest {
							return inst.completePackages(ctx, r.Args().At(1))
						},
					},
				},
				Execute: inst.build,
			},
		},
	})

	return inst
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (c *Command) Name() string {
	return c.commandTree.Node().Name
}

func (c *Command) Description() string {
	return c.commandTree.Node().Description
}

func (c *Command) Complete(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
	return c.commandTree.Complete(ctx, r)
}

func (c *Command) Execute(ctx context.Context, r *readline.Readline) error {
	return c.commandTree.Execute(ctx, r)
}

func (c *Command) Help(ctx context.Context, r *readline.Readline) string {
	return c.commandTree.Help(ctx, r)
}

func (c *Command) Lint(ctx context.Context, fix bool) error {
	c.l.Info("Running golangci-lint run...")

	args := []string{"run"}
	if fix {
		args = append(args, "--fix")
	}

	for _, value := range c.paths(ctx, "go.mod", true) {
		c.l.Info("└ " + value)

		if err := c.execGolangciLint(ctx, args...).Dir(value).Run(); err != nil {
			return err
		}
	}

	return nil
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) build(ctx context.Context, r *readline.Readline) error {
	var paths []string
	if r.Args().HasIndex(1) {
		paths = []string{r.Args().At(1)}
	} else {
		paths = c.paths(ctx, "go.mod", true)
	}

	slices.Sort(paths)

	args := c.getBuildTags()
	if r.AdditionalArgs().Len() > 1 {
		args = r.AdditionalArgs().From(1)
	}

	if value, _ := r.FlagSets().Internal().GetBool("force"); value {
		args = append(args, "-a")
	}

	if value, _ := r.FlagSets().Internal().GetBool("verbose"); value {
		args = append(args, "-v")
	}

	var envs []string
	if value, _ := r.FlagSets().Internal().GetBool("cgo"); value {
		envs = append(envs, "CGO_ENABLED=1")
	} else {
		envs = append(envs, "CGO_ENABLED=0")
	}

	c.l.Info("Building executables ...")

	ctx, wg := c.wg(ctx, r)

	// build only the selected packages
	if r.Args().HasIndex(2) {
		start := time.Now()

		for _, value := range r.Args().From(2) {
			c.l.Infof("🔧 | %s", value)
		}

		if err := shell.New(ctx, c.l, "go", "build", "-buildvcs=false", "-o", "/dev/null").
			Args(args...).
			Args(r.Args().From(2)...).
			Env(envs...).
			Dir(r.Args().At(1)).
			Run(); err != nil {
			c.l.Errorf("🔧 | build failed ⏱︎ %s", time.Since(start).Truncate(time.Second).String())
			return err
		}

		c.l.Successf("🔧 | done ⏱︎ %s", time.Since(start).Truncate(time.Second).String())

		return wg.Wait()
	}

	for i, value := range paths {
		pkgs := c.packages(ctx, value)
		if len(pkgs) == 0 {
			continue
		}

		wg.Go(func() error {
			c.l.Infof("🔧 | [%d|%d] %s", i+1, len(paths), value)

			start := time.Now()
			if err := shell.New(ctx, c.l, "go", "build", "-buildvcs=false", "-o", "/dev/null").
				Args(args...).
				Args(pkgs...).
				Env(envs...).
				Dir(value).
				Run(); err != nil {
				c.l.Errorf("🔧 | [%d|%d] %s ⏱︎ %s", i+1, len(paths), value, time.Since(start).Truncate(time.Second).String())
				return err
			}

			c.l.Successf("🔧 | [%d|%d] %s ⏱︎ %s", i+1, len(paths), value, time.Since(start).Truncate(time.Second).String())

			return nil
		})
	}

	return wg.Wait()
}

func (c *Command) test(ctx context.Context, r *readline.Readline) error {
	return c.goTest(ctx, r, "test", func(name string) []string {
		if name != "" {
			return []string{"-run", "^" + name + "$"}
		}

		return nil
	}, nil)
}

func (c *Command) fuzz(ctx context.Context, r *readline.Readline) error {
	// without a target the corpus is run as regular tests; -fuzz can only target one
	corpus := []string{"-run", "^Fuzz"}

	return c.goTest(ctx, r, "test fuzz", func(name string) []string {
		if name != "" {
			return []string{"-fuzz", "^" + name + "$"}
		}

		return corpus
	}, corpus)
}

func (c *Command) bench(ctx context.Context, r *readline.Readline) error {
	return c.goTest(ctx, r, "test bench", func(name string) []string {
		pattern := "."
		if name != "" {
			pattern = "^" + name + "$"
		}

		return []string{"-bench", pattern, "-run", "^$"}
	}, []string{"-bench", ".", "-run", "^$"})
}

// goTest runs `go test` for the test/fuzz/bench commands. They share the same MODULE/PACKAGE/TARGET
// arguments, logging and parallelism; only the extra flags differ: selected(targetName) yields the
// flags for a chosen package (optionally a single target) and all yields the flags for the
// per-module `./...` fallback.
func (c *Command) goTest(ctx context.Context, r *readline.Readline, label string, selected func(name string) []string, all []string) error {
	var (
		envs  []string
		paths []string
	)

	if r.Args().HasIndex(1) {
		paths = []string{r.Args().At(1)}
	} else {
		paths = c.paths(ctx, "go.mod", true)
	}

	slices.Sort(paths)

	fsi := r.FlagSets().Internal()
	if value, _ := fsi.GetString("tags"); value != "" {
		envs = append(envs, "GO_TEST_TAGS="+value)
	}

	args := c.getBuildTags()
	if r.AdditionalArgs().Len() > 1 {
		args = r.AdditionalArgs().From(1)
	}

	c.l.Infof("Running go %s ...", label)

	// run for the selected package, optionally a single target
	if pkg := r.Args().At(2); pkg != "" {
		runArgs := append(slices.Clone(args), selected(r.Args().At(3))...)

		c.l.Infof("🧪 | %s", pkg)

		start := time.Now()
		if err := shell.New(ctx, c.l, "go", "test", "-v").
			Args(runArgs...).
			Args(pkg).
			Env(envs...).
			Dir(r.Args().At(1)).
			Run(); err != nil {
			c.l.Errorf("🧪 | %s ⏱︎ %s", pkg, time.Since(start).Truncate(time.Second).String())
			return err
		}

		c.l.Successf("🧪 | %s ⏱︎ %s", pkg, time.Since(start).Truncate(time.Second).String())

		return nil
	}

	runArgs := append(slices.Clone(args), all...)

	ctx, wg := c.wg(ctx, r)

	for i, value := range paths {
		wg.Go(func() error {
			c.l.Infof("🧪 | [%d|%d] %s", i+1, len(paths), value)

			start := time.Now()
			if err := shell.New(ctx, c.l, "go", "test", "-v").
				Args(runArgs...).
				Args("./...").
				Env(envs...).
				Dir(value).
				Run(); err != nil {
				c.l.Errorf("🧪 | [%d|%d] %s ⏱︎ %s", i+1, len(paths), value, time.Since(start).Truncate(time.Second).String())
				return err
			}

			c.l.Successf("🧪 | [%d|%d] %s ⏱︎ %s", i+1, len(paths), value, time.Since(start).Truncate(time.Second).String())

			return nil
		})
	}

	return wg.Wait()
}

func (c *Command) modTidy(ctx context.Context, r *readline.Readline) error {
	var paths []string
	if r.Args().HasIndex(2) {
		paths = []string{r.Args().At(2)}
	} else {
		paths = c.paths(ctx, "go.mod", true)
	}

	slices.Sort(paths)

	var args []string
	if r.AdditionalArgs().Len() > 1 {
		args = r.AdditionalArgs().From(1)
	}

	ctx, wg := c.wg(ctx, r)
	c.l.Info("Running go mod tidy...")

	for _, value := range paths {
		wg.Go(func() error {
			c.l.Info("└ " + value)

			return shell.New(ctx, c.l,
				"go", "mod", "tidy",
			).
				Args(args...).
				Dir(value).
				Run()
		})
	}

	return wg.Wait()
}

func (c *Command) modDownload(ctx context.Context, r *readline.Readline) error {
	var paths []string
	if r.Args().HasIndex(2) {
		paths = []string{r.Args().At(2)}
	} else {
		paths = c.paths(ctx, "go.mod", true)
	}

	slices.Sort(paths)

	var args []string
	if r.AdditionalArgs().Len() > 1 {
		args = r.AdditionalArgs().From(1)
	}

	ctx, wg := c.wg(ctx, r)
	c.l.Info("Running go mod download...")

	for _, value := range paths {
		wg.Go(func() error {
			c.l.Info("└ " + value)

			return shell.New(ctx, c.l,
				"go", "mod", "download",
			).
				Args(args...).
				Dir(value).
				Run()
		})
	}

	return wg.Wait()
}

func (c *Command) modOutdated(ctx context.Context, r *readline.Readline) error {
	var paths []string
	if r.Args().HasIndex(2) {
		paths = []string{r.Args().At(2)}
	} else {
		paths = c.paths(ctx, "go.mod", true)
	}

	slices.Sort(paths)

	ctx, wg := c.wg(ctx, r)
	c.l.Info("Running go mod outdated...")

	for _, value := range paths {
		wg.Go(func() error {
			c.l.Info("└ " + value)

			return shell.New(ctx, c.l, "go",
				"list", "-u", "-m", "-json", "all",
				"|",
				"go-mod-outdated", "-update", "-direct",
			).
				Dir(value).
				Run()
		})
	}

	return wg.Wait()
}

func (c *Command) modUpgrade(ctx context.Context, r *readline.Readline) error {
	var paths []string
	if r.Args().HasIndex(2) {
		paths = []string{r.Args().At(2)}
	} else {
		paths = c.paths(ctx, "go.mod", true)
	}

	slices.Sort(paths)

	ctx, wg := c.wg(ctx, r)
	c.l.Info("Running go mod outdated...")

	for _, value := range paths {
		wg.Go(func() error {
			c.l.Info("└ " + value)

			return shell.New(ctx, c.l, "go",
				"list", "-u", "-m", "-f", "'{{if and (not .Indirect) .Update}}{{.Path}}{{end}}'", "all",
				"|",
				"xargs", "-n1", "-I{}", "go", "get", "{}@latest",
			).
				Dir(value).
				Run()
		})
	}

	return wg.Wait()
}

func (c *Command) workInit(ctx context.Context, r *readline.Readline) error {
	if err := shell.New(ctx, c.l, "go", "work", "init").Run(); err != nil {
		return err
	}

	if err := shell.New(ctx, c.l, "go", "work", "use", "-r", ".").Run(); err != nil {
		return err
	}

	if err := shell.New(ctx, c.l, "go", "work", "sync").Run(); err != nil {
		return err
	}

	return nil
}

func (c *Command) workUse(ctx context.Context, r *readline.Readline) error {
	return shell.New(ctx, c.l, "go").
		Args(r.Args()...).
		Args(r.AdditionalArgs()...).
		Run()
}

func (c *Command) workSync(ctx context.Context, r *readline.Readline) error {
	if err := shell.New(ctx, c.l, "go", "work", "use", "-r", ".").Run(); err != nil {
		return err
	}

	if err := shell.New(ctx, c.l, "go", "work", "sync").Run(); err != nil {
		return err
	}

	return nil
}

func (c *Command) lint(ctx context.Context, r *readline.Readline) error {
	fs := r.FlagSets().Default()
	fsi := r.FlagSets().Internal()

	var paths []string
	if r.Args().HasIndex(1) {
		paths = []string{r.Args().At(1)}
	} else {
		paths = c.paths(ctx, "go.mod", true)
	}

	slices.Sort(paths)

	var args []string

	ctx, wg := c.wg(ctx, r)
	c.l.Info("Running golangci-lint run...")

	if value, _ := fsi.GetInt("parallel"); value != 0 {
		args = append(args, "--allow-parallel-runners")
	}

	for _, value := range paths {
		wg.Go(func() error {
			c.l.Info("└ " + value)

			return c.execGolangciLint(ctx, "run").
				Args(args...).
				Args(fs.Visited().Args()...).
				Args(r.Flags()...).
				Args(r.AdditionalArgs()...).
				Dir(value).
				Run()
		})
	}

	return wg.Wait()
}

func (c *Command) cleanBuildCache(ctx context.Context, r *readline.Readline) error {
	return shell.New(ctx, c.l, "go", "clean", "-cache").Run()
}

func (c *Command) cleanTestCache(ctx context.Context, r *readline.Readline) error {
	return shell.New(ctx, c.l, "go", "clean", "-testcache").Run()
}

func (c *Command) cleanModCache(ctx context.Context, r *readline.Readline) error {
	return shell.New(ctx, c.l, "go", "clean", "-modcache").Run()
}

func (c *Command) cleanFuzzCache(ctx context.Context, r *readline.Readline) error {
	return shell.New(ctx, c.l, "go", "clean", "-fuzzcache").Run()
}

func (c *Command) cleanLintCache(ctx context.Context, r *readline.Readline) error {
	return c.execGolangciLint(ctx, "cache", "clean").Run()
}

func (c *Command) generate(ctx context.Context, r *readline.Readline) error {
	var paths []string
	if r.Args().HasIndex(1) {
		paths = append(paths, r.Args().At(1))
	} else {
		paths = c.paths(ctx, "generate.go", false)
	}

	slices.Sort(paths)

	ctx, wg := c.wg(ctx, r)
	c.l.Info("Running go generate...")

	for _, value := range paths {
		wg.Go(func() error {
			c.l.Info("└ " + value)

			return shell.New(ctx, c.l,
				"go", "generate", value,
			).
				Args(r.AdditionalArgs()...).
				Run()
		})
	}

	return wg.Wait()
}

func (c *Command) completePaths(ctx context.Context, filename string, dir bool) []goprompt.Suggest {
	return suggests.List(c.paths(ctx, filename, dir))
}

func (c *Command) completePackages(ctx context.Context, dir string) []goprompt.Suggest {
	return suggests.List(c.packages(ctx, dir))
}

//nolint:forcetypeassert
func (c *Command) packages(ctx context.Context, dir string) []string {
	if dir == "" {
		dir = "."
	}

	return c.cache.Get("packages-"+strings.ReplaceAll(dir, "/", "-"), func() any {
		// only buildable targets, i.e. main packages
		out, err := shell.New(ctx, c.l, "go", "list", "-find", "-f", `'{{if eq .Name "main"}}{{.ImportPath}}{{end}}'`).
			Args(c.getBuildTags()...).
			Args("./...").
			Quiet().
			Dir(dir).
			Output()
		if err != nil {
			c.l.Info("failed to list packages", err.Error())
			return []string{}
		}

		ret := make([]string, 0)

		for line := range strings.SplitSeq(strings.TrimSpace(string(out)), "\n") {
			if line != "" {
				ret = append(ret, line)
			}
		}

		slices.Sort(ret)

		return slices.Compact(ret)
	}).([]string)
}

func (c *Command) completeTestPackages(ctx context.Context, dir string) []goprompt.Suggest {
	return suggests.List(c.testPackages(ctx, dir))
}

//nolint:forcetypeassert
func (c *Command) testPackages(ctx context.Context, dir string) []string {
	if dir == "" {
		dir = "."
	}

	return c.cache.Get("test-packages-"+strings.ReplaceAll(dir, "/", "-"), func() any {
		// only packages that contain tests, rendered relative to their module
		out, err := shell.New(ctx, c.l, "go", "list", "-f", `'{{if or .TestGoFiles .XTestGoFiles}}{{.Module.Path}} {{.ImportPath}}{{end}}'`).
			Args(c.getBuildTags()...).
			Args("./...").
			Quiet().
			Dir(dir).
			Output()
		if err != nil {
			c.l.Debug("failed to list test packages", err.Error())
			return []string{}
		}

		ret := make([]string, 0)

		for line := range strings.SplitSeq(strings.TrimSpace(string(out)), "\n") {
			module, importPath, ok := strings.Cut(line, " ")
			if !ok {
				continue
			}

			switch {
			case importPath == module:
				ret = append(ret, ".")
			case strings.HasPrefix(importPath, module+"/"):
				ret = append(ret, "./"+strings.TrimPrefix(importPath, module+"/"))
			default:
				ret = append(ret, importPath)
			}
		}

		slices.Sort(ret)

		return slices.Compact(ret)
	}).([]string)
}

func (c *Command) completeTests(ctx context.Context, dir, pkg string) []goprompt.Suggest {
	if pkg == "" {
		return nil
	}

	return suggests.List(c.list(ctx, dir, pkg, "Test"))
}

func (c *Command) completeFuzz(ctx context.Context, dir, pkg string) []goprompt.Suggest {
	if pkg == "" {
		return nil
	}

	return suggests.List(c.list(ctx, dir, pkg, "Fuzz"))
}

func (c *Command) completeBenchmarks(ctx context.Context, dir, pkg string) []goprompt.Suggest {
	if pkg == "" {
		return nil
	}

	return suggests.List(c.list(ctx, dir, pkg, "Benchmark"))
}

// list returns the names of a single package's tests/fuzz/benchmarks whose name starts with the
// given prefix (e.g. "Test", "Fuzz", "Benchmark"). Only that package is compiled, so it is fast.
//
//nolint:forcetypeassert
func (c *Command) list(ctx context.Context, dir, pkg, prefix string) []string {
	if dir == "" {
		dir = "."
	}

	return c.cache.Get("list-"+prefix+"-"+strings.ReplaceAll(dir+"-"+pkg, "/", "-"), func() any {
		out, err := shell.New(ctx, c.l, "go", "test", "-list", "^"+prefix).
			Args(c.getBuildTags()...).
			Args(pkg).
			Quiet().
			Dir(dir).
			Output()
		if err != nil {
			c.l.Debug("failed to list tests", err.Error())
			return []string{}
		}

		ret := make([]string, 0)

		for line := range strings.SplitSeq(strings.TrimSpace(string(out)), "\n") {
			// -list also prints a trailing "ok <pkg> <time>" summary; keep only matching names
			if strings.HasPrefix(line, prefix) {
				ret = append(ret, line)
			}
		}

		slices.Sort(ret)

		return slices.Compact(ret)
	}).([]string)
}

//nolint:forcetypeassert
func (c *Command) paths(ctx context.Context, filename string, dir bool) []string {
	return c.cache.Get("paths-"+filename, func() any {
		if value, err := files.Find(ctx, ".", filename, files.FindWithIgnore(`^(node_modules|\.\w*)$`)); err != nil {
			c.l.Debug("failed to walk files", err.Error())
			return []string{}
		} else if dir {
			for i, s := range value {
				value[i] = path.Dir(s)
			}

			return value
		} else {
			return value
		}
	}).([]string)
}

func (c *Command) wg(ctx context.Context, r *readline.Readline) (context.Context, *errgroup.Group) {
	wg, ctx := errgroup.WithContext(ctx)
	if value, _ := r.FlagSets().Internal().GetInt("parallel"); value != 0 {
		wg.SetLimit(value)
	} else {
		wg.SetLimit(1)
	}

	return ctx, wg
}

func (c *Command) getBuildTags() []string {
	var buildTags []string
	if value := env.GetString("GO_BUILD_TAGS", "safe"); value != "" {
		buildTags = append(buildTags, "-tags", value)
	}

	return buildTags
}
