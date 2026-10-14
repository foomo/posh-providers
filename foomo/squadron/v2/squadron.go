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
	"github.com/foomo/posh-providers/kubernetes/kubectl"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/shell"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type (
	// Squadron resolves squadron units and their config files on disk and
	// exposes the squadron cluster/fleet configuration to the Command.
	Squadron struct {
		l         log.Logger
		cfg       squadron.Config
		configKey string
		kubectl   *kubectl.Kubectl
	}
	// Option configures a Squadron via functional options passed to New.
	Option func(*Squadron) error
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

// WithConfigKey sets the viper key the provider reads its configuration
// from, overriding the default "squadron".
func WithConfigKey(v string) Option {
	return func(o *Squadron) error {
		o.configKey = v
		return nil
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

// New builds a Squadron, reading its configuration from viper. It returns an
// error if the configuration cannot be unmarshalled.
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

// Config returns the provider's configuration.
func (s *Squadron) Config() squadron.Config {
	return s.cfg
}

// Cluster returns the squadron.Cluster configured under name and reports
// whether it was found.
func (s *Squadron) Cluster(name string) (squadron.Cluster, bool) {
	return s.cfg.Cluster(name)
}

// Exists reports whether a squadron.yaml exists for the squadron named name.
func (s *Squadron) Exists(name string) bool {
	if _, err := os.Stat(filepath.Join(s.cfg.Path, name, "squadron.yaml")); err != nil {
		return false
	}

	return true
}

// UnitExists reports whether a unit named name is listed for the given
// squadron, cluster and fleet.
func (s *Squadron) UnitExists(ctx context.Context, squadron, cluster, fleet, name string, override bool) bool {
	units, _ := s.ListUnits(ctx, squadron, cluster, fleet, override)
	return slices.Contains(units, name)
}

// List returns the names of the squadrons found below the configured path.
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

// ListUnits runs `squadron list` for the given squadron, cluster and fleet
// and returns the unit names from its output.
func (s *Squadron) ListUnits(ctx context.Context, squadron, cluster, fleet string, override bool) ([]string, error) {
	var units []string

	files := strings.Join(s.GetFiles(squadron, cluster, fleet, override), ",")

	out, err := shell.New(ctx, s.l, "squadron", "list", squadron, "-f", files).
		Dir(s.cfg.Path).
		Output()
	if err != nil {
		return nil, errors.WithMessage(err, string(out))
	}

	for line := range strings.SplitSeq(string(out), "\n") {
		line = stripansi.Strip(line)
		if len(line) > 11 && line[8:11] == "─" {
			units = append(units, line[11:])
		}
	}

	return units, nil
}

// UnitDirs returns the names of the sub directories of squadron that contain
// their own squadron.yaml.
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

// GetFiles returns the squadron config files that exist on disk for the
// given squadron (or every squadron, if squadron is empty), cluster and
// fleet, in override precedence order. Override files are omitted unless
// override is true.
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
