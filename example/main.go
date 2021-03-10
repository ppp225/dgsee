package main

import (
	"context"
	"log"

	"github.com/dgraph-io/dgo/v200"
	"github.com/dgraph-io/dgo/v200/protos/api"
	"github.com/ppp225/dgsee"
	"google.golang.org/grpc"
)

func main() {
	// example for 1million.rdf.gz dataset
	var relations = []dgsee.EdgeCardinality{
		{"-", "actor.film", "n"},
		{"n", "director.film", "n"},
		{"n", "genre", "n"},
		{"-", "performance.actor", "1"},
		{"-", "performance.character", "1"},
		{"-", "performance.film", "1"},
		{"-", "performance.special_performance_type", "0"},
		{"-", "starring", "n"},
		{"-", "type", "0"},
	}

	dg := dgNewClient("localhost:9080")
	ctx := context.Background()
	dgsee.DiscoverEdges(ctx, dg)                                                                         // runs discover, lists relations and prints example RunConsistencyChecks code, as above.
	log.Printf("Schema Consistency Check ResultOK=[%t]", dgsee.RunConsistencyChecks(ctx, dg, relations)) // returns false, if errors in defined relations are found.
}

func dgNewClient(ip string) *dgo.Dgraph {
	conn, err := grpc.Dial(ip, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	return dgo.NewDgraphClient(
		api.NewDgraphClient(conn),
	)
}
