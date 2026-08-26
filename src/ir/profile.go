package ir

import "strconv"

// The three types below mirror Soufflé's own `-p` JSON profile log shape
// exactly (root.program.relation.<name>.num-tuples,
// root.program.relation.<name>.iteration.<i>.num-tuples) -- not a new
// format, so harness/parse_profile.py and harness/tuple_report.py, which
// already parse Soufflé's real profile logs, work unmodified against
// dlc's own output too. §3.7: "The entire measurement apparatus depends
// on this being comparable."

type IterationProfile struct {
	NumTuples int `json:"num-tuples"`
}

type RelationProfile struct {
	NumTuples int                          `json:"num-tuples"`
	Iteration map[string]IterationProfile `json:"iteration,omitempty"`
}

type ProgramProfile struct {
	Relation map[string]RelationProfile `json:"relation"`
}

type RootProfile struct {
	Program ProgramProfile `json:"program"`
}

type ProfileDoc struct {
	Root RootProfile `json:"root"`
}

// EmitProfile converts a set of Relations' Stats into the Soufflé-shaped
// document above. Only rule-derived counts (RelationStats, never raw
// tuple storage) go in -- an EDB relation nobody ever called
// RecordSeedInsert/RecordIterationInsert on contributes num-tuples: 0,
// matching Soufflé's own convention that input relations don't carry a
// derived-tuple count (harness/tuple_report.py's is_input_relation
// check already treats a Soufflé profile this way).
func EmitProfile(relations map[string]*Relation) ProfileDoc {
	rels := map[string]RelationProfile{}
	for name, r := range relations {
		rp := RelationProfile{NumTuples: r.Stats.SeedInserts}
		if len(r.Stats.IterationInserts) > 0 {
			iterMap := map[string]IterationProfile{}
			for i, c := range r.Stats.IterationInserts {
				iterMap[strconv.Itoa(i)] = IterationProfile{NumTuples: c}
			}
			rp.Iteration = iterMap
		}
		rels[name] = rp
	}
	return ProfileDoc{Root: RootProfile{Program: ProgramProfile{Relation: rels}}}
}
