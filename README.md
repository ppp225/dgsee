# dgsee   [![GoDoc](https://godoc.org/github.com/ppp225/dgsee?status.svg)](https://godoc.org/github.com/ppp225/dgsee)   
[dgraph](https://github.com/dgraph-io) consistency check and visualize edge cardinalities

# Why

Planning a schema is one thing, importing and inserting data is another. Issues may sneak into the system without being noticed.
I needed a tool to check if some of my edges have correct cardinalities - to avoid and detect any potential errors in the future.

# Install

```bash
go get -u github.com/ppp225/dgsee

import (
  "github.com/ppp225/dgsee"
)
```

# Usage

See `/example/main.go` for full example.

dgsee offers 2 functions:
```
dgsee.DiscoverEdges(ctx, dg)
ok := dgsee.RunConsistencyChecks(ctx, dg, relations)
```

* `DiscoverEdges` queries schema and then the database, and lists edges along with example consistency check code for this database.
* `RunConsistencyChecks` takes a config, and checks if constraints are met.

## `DiscoverEdges`

Output looks like this:
```
[INFO]               |      (C)ardinality |             | [predicate name]                     | [nodes] |(C)| -edges- |(C)| [nodes] |
[INFO] -------------------------------------------------------------------------------------------------------------------------------
[INFO] Found relation     (1..41)-(1..108) for predicate [director.film]                           [1623] (n)<--7391--->(n)   [6356]
[INFO] Found relation    (1..3290)-(1..14) for predicate [genre]                                   [6356] (n)<--23679-->(n)    [283]
[INFO] Found relation              (-)-(1) for predicate [performance.actor]                     [119258] (-)--119258-->(1)      [-]
...
```
full output can be found here: `/example/output-examples/1million.txt`

At a glance we can see, for example for the `[genre]` edge, that there are 6356 movies, 283 genres, each move has from 1 to 14 genres, and there are 23679 edges.

## `RunConsistencyChecks`

Lets assume 14 genres per movie is too much, and we want only the most relevant. 
We can then enforce, that the above [genre] edge should have between 1 and 3 genres per movie. So we resolve the issues, and then create a definition like:
```
var relations = []dgsee.EdgeCardinality{
  {"n", "genre", "1..3"}
  // {"n", "genre", "n"} // or if we want to keep it without constrants
}
ok := dgsee.RunConsistencyCheck(ctx, dg, relations) // returns true if constraints are met, or false if they fail.
```
# Future plans

* Add checks for other predicates than uid
* Performance tune (currently uses has(*) function for all edges. Takes a second for 1million dataset and a minute for 21million)
* Integrate with types - currently doesn't detect orphan nodes (i.e. 1..n will not fail (when a movie has no genres) because we don't know the starting node)

# Note

Everything may or may not change ¯\\\_(ツ)\_/¯
