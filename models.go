package dgsee

// EdgeCardinality represents a single edge with cardinality of the relation
// if edge isn't reverse, from should be "-"
// example of correct from/to values: "0" "1" "2" "n" " 0..1" "0..2" "1..n" "1..5"
// examples: {"1", "movieCredit", "n"}, {"n", "parent", "2"}, {"1..n", "something", "n"},
type EdgeCardinality struct {
	From string
	Edge string
	To   string
}

const (
	discoverQueryFormat = `query {
		var(func: has(<%[1]s>)) {
			CNT as count(<%[1]s>)
			NODE as count(uid)
		}
		q(){
			max: max(val(CNT))
			min: min(val(CNT))
			nodes: sum(val(NODE))
			edges: sum(val(CNT))
		}
	}`
	// scc = schema consistency check
	sccQueryFormat = `query {
		var(func: has(<%[1]s>)) {
			CNT as count(<%[1]s>)
			NODE as count(uid)
		}
		var(func: uid(CNT)) @filter(%[2]s) {
			ERRORS as count(uid)
		}
		q(){
			max: max(val(CNT))
			min: min(val(CNT))
			nodes: sum(val(NODE))
			errors: min(val(ERRORS))
		}
	}`
	// filters filter out correct values, so only errors remain
	filterExact   = `NOT eq(val(CNT), %[1]s)`
	filterFromTo  = `lt(val(CNT), %[1]s) OR gt(val(CNT), %[2]s)`
	filterFromToN = `lt(val(CNT), %[1]s)`
	// filterZeroToN = `always true`
)
