package ptermx_test

import (
	"testing"

	"github.com/foomo/posh-providers/pkg/ptermx"
	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"
)

func TestLeveledList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  pterm.TreeNode
		expect pterm.LeveledList
	}{
		{
			name:   "single node",
			input:  pterm.TreeNode{Text: "root"},
			expect: pterm.LeveledList{{Level: 0, Text: "root"}},
		},
		{
			name: "one level of children",
			input: pterm.TreeNode{
				Text:     "root",
				Children: []pterm.TreeNode{{Text: "a"}, {Text: "b"}},
			},
			expect: pterm.LeveledList{
				{Level: 0, Text: "root"},
				{Level: 1, Text: "a"},
				{Level: 1, Text: "b"},
			},
		},
		{
			name: "nested depth three, pre-order",
			input: pterm.TreeNode{
				Text: "root",
				Children: []pterm.TreeNode{
					{Text: "a", Children: []pterm.TreeNode{{Text: "a1"}}},
					{Text: "b"},
				},
			},
			expect: pterm.LeveledList{
				{Level: 0, Text: "root"},
				{Level: 1, Text: "a"},
				{Level: 2, Text: "a1"},
				{Level: 1, Text: "b"},
			},
		},
		{
			name:   "empty text root is preserved",
			input:  pterm.TreeNode{},
			expect: pterm.LeveledList{{Level: 0, Text: ""}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expect, ptermx.LeveledList(tt.input))
		})
	}
}
