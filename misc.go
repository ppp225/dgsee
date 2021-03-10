package dgsee

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/dgraph-io/dgo/v200"
	log "github.com/ppp225/lvlog"
)

type schema []struct {
	Predicate string `json:"predicate"`
	Type      string `json:"type"`
	Reverse   bool   `json:"reverse,omitempty"`
}

func querySchema(ctx context.Context, txn *dgo.Txn) schema {
	q := `schema {
			type
			reverse
		}`
	resp, err := txn.Query(ctx, q)
	if err != nil {
		panic(err)
	}
	var decode struct {
		Schema schema `json:"schema"`
	}
	if err := json.Unmarshal(resp.GetJson(), &decode); err != nil {
		log.Panic(err)
	}
	return decode.Schema
}

func formatCardinality(from, to, bmin, bmax, fmin, fmax string) (bminmax, fminmax string) {
	switch {
	case bmin == "-" && bmax == "-" && (from == "n"): // when relation is "n", no query is made, as it always resolves correctly
		bminmax = from
	case bmin == bmax:
		bminmax = bmin
	default:
		bminmax = fmt.Sprintf("%s..%s", bmin, bmax)
	}
	switch {
	case fmin == "-" && fmax == "-" && (to == "n"): // when relation is "n", no query is made, as it always resolves correctly
		fminmax = to
	case fmin == fmax:
		fminmax = fmin
	default:
		fminmax = fmt.Sprintf("%s..%s", fmin, fmax)
	}
	return bminmax, fminmax
}

func formatCardinalityShort(bmin, bmax, fmin, fmax string) (bs, fs string) {
	bmaxi, _ := strconv.Atoi(bmax)
	fmaxi, _ := strconv.Atoi(fmax)
	switch {
	case bmin == bmax:
		bs = bmin
	case bmaxi > 1:
		bs = "n"
	default:
		bs = fmt.Sprintf("%s..%s", bmin, bmax)
	}
	switch {
	case fmin == fmax:
		fs = fmin
	case fmaxi > 1:
		fs = "n"
	default:
		fs = fmt.Sprintf("%s..%s", fmin, fmax)
	}
	return
}
