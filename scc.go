package dgsee

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dgraph-io/dgo"
	log "github.com/ppp225/lvlog"
)

// RunConsistencyChecks checks, if schema relations have correct 1-1 or 1-n relations
func RunConsistencyChecks(ctx context.Context, dg *dgo.Dgraph, relations []EdgeCardinality) (ok bool) {
	txn := dg.NewTxn()
	defer txn.Discard(ctx)
	// check if all edges are accounted for
	schema := querySchema(ctx, txn)
	count := 0
	for _, r := range schema {
		if r.Type == "uid" {
			count++
		}
	}
	if count != len(relations) {
		log.Warnf("dgsee.RunConsistencyChecks: Not all edges are checked. Checking %d edges out of %d found in schema", len(relations), count)
	}
	return runSCC(ctx, txn, relations)
}

func runSCC(ctx context.Context, txn *dgo.Txn, relations []EdgeCardinality) (ok bool) {
	ok = true
	for _, r := range relations {
		forwardQuery := ""
		backwardQuery := ""
		// assign queries
		edge := r.Edge
		filter := getFilter(r.To)
		if filter != "" {
			forwardQuery = fmt.Sprintf(sccQueryFormat, edge, filter)
		}
		edge = "~" + r.Edge
		filter = getFilter(r.From)
		if filter != "" {
			backwardQuery = fmt.Sprintf(sccQueryFormat, edge, filter)
		}
		// run queries
		ferr, fmin, fmax, fnodes := runQueryAndReturnErrorCount(ctx, txn, forwardQuery)
		berr, bmin, bmax, bnodes := runQueryAndReturnErrorCount(ctx, txn, backwardQuery)
		if ferr != 0 || berr != 0 {
			ok = false
			log.Errorf(formatErrorMsg(r.Edge, r.From, r.To, bmin, bmax, fmin, fmax, berr, ferr))
		}
		if fnodes == 0 || bnodes == 0 {
			ok = false
			log.Errorf(formatErrorNoRelationsMsg(r.Edge, r.From, r.To, bmin, bmax, fmin, fmax, berr, ferr))
		}
		// log.Infof(formatInfoMsg(r.Edge, r.From, r.To, bmin, bmax, fmin, fmax))
	}
	return
}
func getFilter(relation string) (filter string) {
	rel := strings.ToLower(relation)
	// no reverse edge - ignore
	if rel == "-" || rel == "n" || rel == "0..n" {
		return ""
	}
	arr := strings.Split(rel, "..")
	if len(arr) == 1 {
		return fmt.Sprintf(filterExact, arr[0])
	}
	from := arr[0]
	to := arr[1]
	if to == "n" {
		return fmt.Sprintf(filterFromToN, from)
	}
	return fmt.Sprintf(filterFromTo, from, to)
}

func runQueryAndReturnErrorCount(ctx context.Context, txn *dgo.Txn, query string) (errorCount int, min, max string, nodes int) {
	// if no query, no errors
	if query == "" {
		return 0, "-", "-", -1
	}
	// run query
	resp, err := txn.Query(ctx, query)
	if err != nil {
		panic(err)
	}
	// decode response
	var decode struct {
		Q []struct {
			Max   float64 `json:"max"`
			Min   float64 `json:"min"`
			Nodes int     `json:"nodes"`
			Count int     `json:"errors"`
		} `json:"q"`
	}
	if err := json.Unmarshal(resp.GetJson(), &decode); err != nil {
		log.Panic(err)
	}
	// returns 0 if no errors, and min max values for error reporting
	return decode.Q[3].Count, fmt.Sprintf("%.0f", decode.Q[1].Min), fmt.Sprintf("%.0f", decode.Q[0].Max), decode.Q[2].Nodes
}

func formatInfoMsg(edge, from, to, bmin, bmax, fmin, fmax string) (formatted string) {
	bminmax, fminmax := formatCardinality(from, to, bmin, bmax, fmin, fmax)
	return fmt.Sprintf("Edge [%s] has relation (%s)-(%s), as defined. Exactly: (%s)-(%s).", edge, from, to, bminmax, fminmax)
}
func formatErrorMsg(edge, from, to, bmin, bmax, fmin, fmax string, berr, ferr int) (formatted string) {
	bminmax, fminmax := formatCardinality(from, to, bmin, bmax, fmin, fmax)
	return fmt.Sprintf("Expected edge [%s] to have relation (%s)-(%s), but got (%s)-(%s). Number of errors: (%d)-(%d).", edge, from, to, bminmax, fminmax, berr, ferr)
}
func formatErrorNoRelationsMsg(edge, from, to, bmin, bmax, fmin, fmax string, berr, ferr int) (formatted string) {
	bminmax, fminmax := formatCardinality(from, to, bmin, bmax, fmin, fmax)
	return fmt.Sprintf("Expected edge [%s] to have relation (%s)-(%s), but got (%s)-(%s). Dead edge. No nodes with this edge exist!", edge, from, to, bminmax, fminmax)
}
