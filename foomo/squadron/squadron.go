package squadron

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/foomo/posh-providers/kubernets/kubectl"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/shell"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/exp/slices"
)

type (
	Squadron struct {
		l         log.Logger
		cfg       Config
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

func (s *Squadron) Cluster(name string) (Cluster, bool) {
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
	out, err := shell.New(ctx, s.l, "squadron", "list", "-f", files).
		Dir(path.Join(s.cfg.Path, squadron)).
		Output()
	if err != nil {
		return nil, errors.WithMessage(err, string(out))
	}
	for _, line := range strings.Split(string(out), "\n") {
		if line == "" {
			continue
		}
		units = append(units, line)
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
	var files []string
	allFiles := []string{
		"../squadron.yaml",
		"../squadron.override.yaml",
		fmt.Sprintf("../squadron.%s.yaml", fleet),
		fmt.Sprintf("../squadron.%s.override.yaml", fleet),
		fmt.Sprintf("../squadron.%s.yaml", cluster),
		fmt.Sprintf("../squadron.%s.override.yaml", cluster),
		fmt.Sprintf("../squadron.%s.%s.yaml", cluster, fleet),
		fmt.Sprintf("../squadron.%s.%s.override.yaml", cluster, fleet),
		"squadron.yaml",
		"squadron.override.yaml",
		fmt.Sprintf("squadron.%s.yaml", fleet),
		fmt.Sprintf("squadron.%s.override.yaml", fleet),
		fmt.Sprintf("squadron.%s.yaml", cluster),
		fmt.Sprintf("squadron.%s.override.yaml", cluster),
		fmt.Sprintf("squadron.%s.%s.yaml", cluster, fleet),
		fmt.Sprintf("squadron.%s.%s.override.yaml", cluster, fleet),
	}
	for _, unit := range s.UnitDirs(squadron) {
		allFiles = append(allFiles,
			fmt.Sprintf("%s/squadron.yaml", unit),
			fmt.Sprintf("%s/squadron.override.yaml", unit),
			fmt.Sprintf("%s/squadron.%s.yaml", unit, fleet),
			fmt.Sprintf("%s/squadron.%s.override.yaml", unit, fleet),
			fmt.Sprintf("%s/squadron.%s.yaml", unit, cluster),
			fmt.Sprintf("%s/squadron.%s.override.yaml", unit, cluster),
			fmt.Sprintf("%s/squadron.%s.%s.yaml", unit, cluster, fleet),
			fmt.Sprintf("%s/squadron.%s.%s.override.yaml", unit, cluster, fleet),
		)
	}
	for _, file := range allFiles {
		if !override && strings.Contains(file, ".override.") {
			continue
		}
		if _, err := os.Stat(filepath.Join(s.cfg.Path, squadron, file)); err == nil {
			files = append(files, file)
		}
	}
	return files
}
