package gocontentful

import (
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/stretchr/testify/require"
)

func TestBuildArgs(t *testing.T) {
	t.Parallel()
	testingx.Tags(t, tagx.Short)

	const dir = "path/to/pkg"

	full := Config{
		SpaceID:      "space",
		CMAKey:       "key",
		Environment:  "master",
		Region:       "us",
		ContentTypes: []string{"a", "b", "c"},
	}

	tests := []struct {
		name     string
		cfg      func(cfg Config) Config
		expected []string
	}{
		{
			name: "all set",
			cfg:  func(cfg Config) Config { return cfg },
			expected: []string{
				"-spaceid", "space", "-cmakey", "key", "-environment", "master",
				"-region", "us", "-contenttypes", "a,b,c", dir,
			},
		},
		{
			name: "region empty",
			cfg:  func(cfg Config) Config { cfg.Region = ""; return cfg },
			expected: []string{
				"-spaceid", "space", "-cmakey", "key", "-environment", "master",
				"-contenttypes", "a,b,c", dir,
			},
		},
		{
			name: "cmakey empty",
			cfg:  func(cfg Config) Config { cfg.CMAKey = ""; return cfg },
			expected: []string{
				"-spaceid", "space", "-environment", "master",
				"-region", "us", "-contenttypes", "a,b,c", dir,
			},
		},
		{
			name: "contenttypes empty",
			cfg:  func(cfg Config) Config { cfg.ContentTypes = nil; return cfg },
			expected: []string{
				"-spaceid", "space", "-cmakey", "key", "-environment", "master",
				"-region", "us", dir,
			},
		},
		{
			name: "region eu",
			cfg:  func(cfg Config) Config { cfg.Region = "eu"; return cfg },
			expected: []string{
				"-spaceid", "space", "-cmakey", "key", "-environment", "master",
				"-region", "eu", "-contenttypes", "a,b,c", dir,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual := buildArgs(tt.cfg(full), dir)
			require.Equal(t, tt.expected, actual)
			require.Equal(t, dir, actual[len(actual)-1])
		})
	}
}
