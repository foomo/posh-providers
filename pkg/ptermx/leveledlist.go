// Package ptermx provides conversions for pterm types that pterm itself does
// not ship.
package ptermx

import (
	"github.com/pterm/pterm"
)

// LeveledList flattens a pterm.TreeNode into the pterm.LeveledList that
// agent.Tree consumes, walking depth-first pre-order so the flat order matches
// the way the tree renders.
//
// It is the inverse of putils.TreeFromLeveledList, which pterm ships but only
// in the one direction. Commands that already build a TreeNode - the natural
// shape when the data is nested - need this to reach agent.Tree without
// rewriting their renderer.
//
// The root becomes level 0 and every child sits one level below its parent, so
// a round trip through putils.TreeFromLeveledList reproduces the input.
func LeveledList(node pterm.TreeNode) pterm.LeveledList {
	var ret pterm.LeveledList

	var walk func(n pterm.TreeNode, level int)
	walk = func(n pterm.TreeNode, level int) {
		ret = append(ret, pterm.LeveledListItem{Level: level, Text: n.Text})

		for _, child := range n.Children {
			walk(child, level+1)
		}
	}

	walk(node, 0)

	return ret
}
