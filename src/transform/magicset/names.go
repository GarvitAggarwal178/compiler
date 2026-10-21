package magicset

import (
	"fmt"

	"dlc/src/ast"
	"dlc/src/transform"
)

type checkpointKey struct {
	predicate      adornedKey
	rule, position int
}

// GeneratedNames assigns names by semantic role, not by a shared string
// prefix: magic(p,bf) and adorned(magic_p,bf) otherwise have the same name.
// Existing benchmark names survive when available. Allocation is deterministic
// and reserves every source relation before choosing any generated name.
type GeneratedNames struct {
	Origins          map[string]transform.RelationOrigin
	relations, magic map[adornedKey]string
	checkpoints      map[checkpointKey]string
}

func AllocateNames(prog *ast.Program, result *AdornResult) *GeneratedNames {
	n := &GeneratedNames{Origins: map[string]transform.RelationOrigin{}, relations: map[adornedKey]string{}, magic: map[adornedKey]string{}, checkpoints: map[checkpointKey]string{}}
	sourceIDs := map[*ast.Clause]string{}
	for i, c := range prog.Clauses {
		sourceIDs[c] = transform.RuleID(i)
	}
	used := map[string]bool{}
	for _, d := range prog.Decls {
		used[d.Name] = true
	}
	fresh := func(preferred string) string {
		name := preferred
		for i := 1; used[name]; i++ {
			name = fmt.Sprintf("%s_dlc%d", preferred, i)
		}
		used[name] = true
		return name
	}
	for _, key := range result.Order {
		n.relations[key] = fresh(key.RelName())
		n.magic[key] = fresh(key.MagicRelName())
		n.Origins[n.relations[key]] = transform.RelationOrigin{ID: n.relations[key], Role: "adorned", SourcePredicate: key.pred}
		n.Origins[n.magic[key]] = transform.RelationOrigin{ID: n.magic[key], Role: "magic", SourcePredicate: key.pred}
		for i, rule := range result.Rules[key] {
			for k := 0; k <= len(rule.OrderedBody); k++ {
				name := fresh(fmt.Sprintf("sup_%s_r%d_%d", key.RelName(), i, k))
				n.checkpoints[checkpointKey{key, i, k}] = name
				n.Origins[name] = transform.RelationOrigin{ID: name, Role: "checkpoint", SourcePredicate: key.pred, SourceRuleID: sourceIDs[rule.Source]}
			}
		}
	}
	return n
}

func (n *GeneratedNames) Relation(key adornedKey) string { return n.relations[key] }
func (n *GeneratedNames) Magic(key adornedKey) string    { return n.magic[key] }
func (n *GeneratedNames) Checkpoint(key adornedKey, rule, position int) string {
	return n.checkpoints[checkpointKey{key, rule, position}]
}

// requiredOriginals preserves full extents needed outside the specialized
// query path. Keeping declarations alone is insufficient: every untouched
// consumer (including negation) needs its original dependencies' rules too.
func requiredOriginals(prog *ast.Program, result *AdornResult, declined map[string]bool, queries map[*ast.Clause]bool) map[string]bool {
	required := map[string]bool{}
	for _, d := range prog.Decls {
		if (d.Kind == ast.DeclOutput || d.Kind == ast.DeclInput) && predicateAdorned(result, d.Name) {
			required[d.Name] = true
		}
	}
	dependencies := func(c *ast.Clause) []string {
		var names []string
		for _, lit := range c.Body {
			switch v := lit.(type) {
			case *ast.Atom:
				names = append(names, v.Name)
			case *ast.NegatedAtom:
				names = append(names, v.Atom.Name)
			}
		}
		return names
	}
	for _, c := range prog.Clauses {
		if (result.Untouched[c.Head.Name] || declined[c.Head.Name]) && !queries[c] {
			for _, name := range dependencies(c) {
				required[name] = true
			}
		}
	}
	for changed := true; changed; {
		changed = false
		for _, c := range prog.Clauses {
			if required[c.Head.Name] {
				for _, name := range dependencies(c) {
					if !required[name] {
						required[name] = true
						changed = true
					}
				}
			}
		}
	}
	return required
}
