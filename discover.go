package dgsee

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/dgo"
	log "github.com/ppp225/lvlog"
)

// DiscoverEdges queries schema and lists all relations. Prints out example code to consistency check edges for found cardinalities
func DiscoverEdges(ctx context.Context, dg *dgo.Dgraph) {
	txn := dg.NewTxn()
	defer txn.Discard(ctx)
	// get schema
	schema := querySchema(ctx, txn)
	maxPredicateLen := 0
	for _, r := range schema {
		if r.Type == "uid" {
			if len(r.Predicate) > maxPredicateLen {
				maxPredicateLen = len(r.Predicate)
			}
		}
	}
	// for each edge
	schemaConsistencyString := `// Example schema consistency check code for this schema:
func main() {
var relations = []dgsee.EdgeCardinality{
`
	log.Infof("              | (C)ardinality |             | predicate name                 | nodes|(C)| edges |(C)|nodes")
	log.Infof("-----------------------------------------------------------------------------------------------------------")
	for _, r := range schema {
		if r.Type != "uid" {
			continue
		}
		forwardQuery := fmt.Sprintf(discoverQueryFormat, r.Predicate)
		backwardQuery := ""
		if r.Reverse {
			backwardQuery = fmt.Sprintf(discoverQueryFormat, "~"+r.Predicate)
		}
		fmax, fmin, fnodes, edges := runDiscoverQuery(ctx, txn, forwardQuery)
		bmax, bmin, bnodes, _ := runDiscoverQuery(ctx, txn, backwardQuery)
		log.Infof(formatDiscoverMsg(r.Predicate, bmin, bmax, fmin, fmax, bnodes, fnodes, edges, r.Reverse, maxPredicateLen))
		schemaConsistencyString += formatRelationSCCString(r.Predicate, bmin, bmax, fmin, fmax)
	}
	schemaConsistencyString += `}
dg := dgNewClient("localhost:9080")
ctx := context.Background()
log.Printf("Schema Consistency Check ResultOK=[%t]", RunConsistencyChecks(ctx, dg, relations))
}`
	log.Info(schemaConsistencyString)
}

func formatRelationSCCString(edge, bmin, bmax, fmin, fmax string) string {
	bs, fs := formatCardinalityShort(bmin, bmax, fmin, fmax)
	return fmt.Sprintf("{\"%s\", \"%s\", \"%s\"},\n", bs, edge, fs)
}

func formatDiscoverMsg(edge, bmin, bmax, fmin, fmax, bnodes, fnodes, edges string, reverse bool, edgeMaxLen int) (formatted string) {
	bminmax, fminmax := formatCardinality("-", "-", bmin, bmax, fmin, fmax)
	bs, fs := formatCardinalityShort(bmin, bmax, fmin, fmax)
	fromto := fmt.Sprintf("(%s)-(%s)", bminmax, fminmax)
	paddingLen := fmt.Sprintf("%%%ds", edgeMaxLen-len(edge)+2)
	padding := fmt.Sprintf(paddingLen, "")
	// pad edges num
	edgesPadded := edges
	switch len(edges) {
	case 1:
		edgesPadded = "-" + edgesPadded
		fallthrough
	case 2:
		edgesPadded = edgesPadded + "-"
		fallthrough
	case 3:
		edgesPadded = "-" + edgesPadded
		fallthrough
	case 4:
		edgesPadded = edgesPadded + "-"
		fallthrough
	case 5:
		edgesPadded = "-" + edgesPadded
		fallthrough
	case 6:
		edgesPadded = edgesPadded + "-"
		fallthrough
	default:
	}
	// reverse?
	rev := "-"
	if reverse {
		rev = "<"
	}
	return fmt.Sprintf("Found relation %20s for predicate [%s]%s%8s (%s)%s-%s->(%s) %8s", fromto, edge, padding, "["+fnodes+"]", bs, rev, edgesPadded, fs, "["+bnodes+"]")
}

func runDiscoverQuery(ctx context.Context, txn *dgo.Txn, query string) (max, min, nodes, edges string) {
	// if no query, no errors
	if query == "" {
		return "-", "-", "-", "-"
	}
	// run query
	resp, err := txn.Query(ctx, query)
	if err != nil {
		panic(err)
	}
	// decode response NODE: float, because count on not existing nodes gives a float
	var decode struct {
		Q []struct {
			Max   float32 `json:"max"`
			Min   float32 `json:"min"`
			Nodes float32 `json:"nodes"`
			Edges float32 `json:"edges"`
		} `json:"q"`
	}
	if err := json.Unmarshal(resp.GetJson(), &decode); err != nil {
		log.Panic(err)
	}
	// returns decode values as strings
	return fmt.Sprintf("%.0f", decode.Q[0].Max), fmt.Sprintf("%.0f", decode.Q[1].Min), fmt.Sprintf("%.0f", decode.Q[2].Nodes), fmt.Sprintf("%.0f", decode.Q[3].Edges)
}
