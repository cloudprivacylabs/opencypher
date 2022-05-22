package main

import (
	"fmt"

	"github.com/cloudprivacylabs/opencypher"
	"github.com/cloudprivacylabs/opencypher/graph"
)

func main() {
	// Create an empty graph
	grph := graph.NewOCGraph()
	// Evaluation context knows the graph we are working on
	ectx := opencypher.NewEvalContext(grph)
	// CREATE two nodes, one of which is a Person
	_, err := opencypher.ParseAndEvaluate(`CREATE (n:Person), (m {name:"Andy", title: "Developer"})`, ectx)
	if err != nil {
		panic(err)
	}
	// Get the Person node
	v, err := opencypher.ParseAndEvaluate(`MATCH (x:Person) return x as person`, ectx)
	if err != nil {
		panic(err)
	}
	// MATCH returns a result set. The first row of the resultset has
	// the named result "person" which is a node
	fmt.Println(v.Get().(opencypher.ResultSet).Rows[0]["person"])

	// Find Andy
	v, err = opencypher.ParseAndEvaluate(`MATCH (y {name:"Andy"}) return y`, ectx)
	if err != nil {
		panic(err)
	}
	// If results are not explicitly named, they are returned as column
	// "1", column "2", etc.
	fmt.Println(v.Get().(opencypher.ResultSet).Rows[0]["1"])

	// Find all nodes
	v, err = opencypher.ParseAndEvaluate(`MATCH (k) return k`, ectx)
	if err != nil {
		panic(err)
	}
	for _, row := range v.Get().(opencypher.ResultSet).Rows {
		fmt.Println(row)
	}
}
