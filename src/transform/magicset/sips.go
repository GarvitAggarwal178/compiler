package magicset

import (
	"dlc/src/ast"
	"dlc/src/plan"
)

// OrderBody shares the executable planner, seeded with the actual head demand.
func OrderBody(body []ast.Literal, initBound map[string]bool) []ast.Literal {
	result, err := plan.Body(body, initBound)
	if err != nil {
		panic(err)
	}
	return result.Literals
}

func arithVarsBound(a ast.Arith, bound map[string]bool) bool {
	return plan.ArithBound(a, bound)
}

func addBoundVars(lit ast.Literal, bound map[string]bool) {
	plan.Bind(lit, bound)
}
