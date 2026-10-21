package sema

import (
	"dlc/src/ast"
	"dlc/src/plan"
)

// AtomInputsBound is the compatibility entry point for the shared planner.
func AtomInputsBound(a *ast.Atom, bound map[string]bool) bool {
	return plan.AtomInputsBound(a, bound)
}
