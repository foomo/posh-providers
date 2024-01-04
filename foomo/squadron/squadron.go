package squadron

import (
	"context"
)

type Squadron interface {
	Cluster(name string) (Cluster, bool)
	Exists(name string) bool
	UnitExists(ctx context.Context, squadron, cluster, fleet, name string, override bool) bool
	List() ([]string, error)
	ListUnits(ctx context.Context, squadron, cluster, fleet string, override bool) ([]string, error)
	UnitDirs(squadron string) []string
	GetFiles(squadron, cluster, fleet string, override bool) []string
}
