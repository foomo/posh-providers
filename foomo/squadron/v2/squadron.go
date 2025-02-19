package squadron

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/acarl005/stripansi"
	"github.com/foomo/posh-providers/foomo/squadron"
	"github.com/foomo/posh-providers/kubernets/kubectl"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/shell"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type (
	Squadron struct {
		l         log.Logger
		cfg       squadron.Config
		configKey string
		kubectl   *kubectl.Kubectl
	}
	Option func(*Squadron) error
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func WithConfigKey(v string) Option {
	return func(o *Squadron) error {
		o.configKey = v
		return nil
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func New(l log.Logger, kubectl *kubectl.Kubectl, opts ...Option) (*Squadron, error) {
	inst := &Squadron{
		l:         l,
		kubectl:   kubectl,
		configKey: "squadron",
	}
	for _, opt := range opts {
		if opt != nil {
			if err := opt(inst); err != nil {
				return nil, err
			}
		}
	}
	if err := viper.UnmarshalKey(inst.configKey, &inst.cfg); err != nil {
		return nil, err
	}

	return inst, nil
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (s *Squadron) Cluster(name string) (squadron.Cluster, bool) {
	return s.cfg.Cluster(name)
}

func (s *Squadron) Exists(name string) bool {
	if _, err := os.Stat(filepath.Join(s.cfg.Path, name, "squadron.yaml")); err != nil {
		return false
	}
	return true
}

func (s *Squadron) UnitExists(ctx context.Context, squadron, cluster, fleet, name string, override bool) bool {
	units, _ := s.ListUnits(ctx, squadron, cluster, fleet, override)
	return slices.Contains(units, name)
}

func (s *Squadron) List() ([]string, error) {
	files, err := os.ReadDir(s.cfg.Path)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, value := range files {
		if value.IsDir() && !strings.HasPrefix(value.Name(), ".") {
			results = append(results, value.Name())
		}
	}
	return results, nil
}

func (s *Squadron) ListUnits(ctx context.Context, squadron, cluster, fleet string, override bool) ([]string, error) {
	var units []string
	files := strings.Join(s.GetFiles(squadron, cluster, fleet, override), ",")
	out, err := shell.New(ctx, s.l, "squadron", "list", squadron, "-f", files).
		Dir(s.cfg.Path).
		Output()
	if err != nil {
		return nil, errors.WithMessage(err, string(out))
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = stripansi.Strip(line)
		if len(line) > 11 && line[8:11] == "─" {
			units = append(units, line[11:])
		}
	}
	return units, nil
}

func (s *Squadron) UnitDirs(squadron string) []string {
	var units []string
	dirs, err := os.ReadDir(filepath.Join(s.cfg.Path, squadron))
	if err != nil {
		return nil
	}
	for _, dir := range dirs {
		if _, err := os.Stat(filepath.Join(s.cfg.Path, squadron, dir.Name(), "squadron.yaml")); err == nil {
			units = append(units, dir.Name())
		}
	}
	return units
}

func (s *Squadron) GetFiles(squadron, cluster, fleet string, override bool) []string {
	allFiles := []string{
		"squadron.yaml",
		"squadron.override.yaml",
		fmt.Sprintf("squadron.%s.yaml", fleet),
		fmt.Sprintf("squadron.%s.override.yaml", fleet),
		fmt.Sprintf("squadron.%s.yaml", cluster),
		fmt.Sprintf("squadron.%s.override.yaml", cluster),
		fmt.Sprintf("squadron.%s.%s.yaml", cluster, fleet),
		fmt.Sprintf("squadron.%s.%s.override.yaml", cluster, fleet),
	}
	var squadrons []string
	if squadron == "" {
		if value, err := s.List(); err != nil {
			s.l.Debug(err.Error())
			return nil
		} else {
			squadrons = value
		}
	} else {
		squadrons = []string{squadron}
	}
	for _, value := range squadrons {
		allFiles = append(allFiles,
			fmt.Sprintf("%s/squadron.yaml", value),
			fmt.Sprintf("%s/squadron.override.yaml", value),
			fmt.Sprintf("%s/squadron.%s.yaml", value, fleet),
			fmt.Sprintf("%s/squadron.%s.override.yaml", value, fleet),
			fmt.Sprintf("%s/squadron.%s.yaml", value, cluster),
			fmt.Sprintf("%s/squadron.%s.override.yaml", value, cluster),
			fmt.Sprintf("%s/squadron.%s.%s.yaml", value, cluster, fleet),
			fmt.Sprintf("%s/squadron.%s.%s.override.yaml", value, cluster, fleet),
		)
		for _, unit := range s.UnitDirs(value) {
			allFiles = append(allFiles,
				fmt.Sprintf("%s/%s/squadron.yaml", value, unit),
				fmt.Sprintf("%s/%s/squadron.override.yaml", value, unit),
				fmt.Sprintf("%s/%s/squadron.%s.yaml", value, unit, fleet),
				fmt.Sprintf("%s/%s/squadron.%s.override.yaml", value, unit, fleet),
				fmt.Sprintf("%s/%s/squadron.%s.yaml", value, unit, cluster),
				fmt.Sprintf("%s/%s/squadron.%s.override.yaml", value, unit, cluster),
				fmt.Sprintf("%s/%s/squadron.%s.%s.yaml", value, unit, cluster, fleet),
				fmt.Sprintf("%s/%s/squadron.%s.%s.override.yaml", value, unit, cluster, fleet),
			)
		}
	}
	var files []string
	for _, file := range allFiles {
		if !override && strings.Contains(file, ".override.") {
			continue
		}
		if _, err := os.Stat(filepath.Join(s.cfg.Path, file)); err == nil {
			files = append(files, file)
		}
	}
	return files
}
