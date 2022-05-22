package main

import (
	"fmt"

	"github.com/cloudprivacylabs/opencypher"
	"github.com/cloudprivacylabs/opencypher/graph"
)

// Evaluation context keeps variables defined in expressions. It can
// be used to link queries.

func main() {
	// Create an empty graph
	grph := graph.NewOCGraph()
	// Evaluation context knows the graph we are working on
	ectx := opencypher.NewEvalContext(grph)
	// CREATE a path
	_, err := opencypher.ParseAndEvaluate(`CREATE (andy {name:"Andy"})-[:KNOWS]-> (stephen {name:"Stephen"})`, ectx)
	if err != nil {
		panic(err)
	}
	// ectx knows andy and stephen. So this will only update stephen, and not andy
	v, err := opencypher.ParseAndEvaluate(`MATCH (stephen) SET stephen.age=34 return stephen`, ectx)
	if err != nil {
		panic(err)
	}
	age, _ := v.Get().(opencypher.ResultSet).Rows[0]["1"].Get().(graph.Node).GetProperty("age")
	fmt.Println(age) // This will print 34

	// The context knows andy
	v, err = opencypher.ParseAndEvaluate(`return andy.name,andy.age`, ectx)
	if err != nil {
		panic(err)
	}
	name := v.Get().(opencypher.ResultSet).Rows[0]["1"].Get()
	age = v.Get().(opencypher.ResultSet).Rows[0]["2"].Get()
	fmt.Println(name, age) // This will print Andy <nil>, andy does not have an age propety

	// This will set age for all nodes
	_, err = opencypher.ParseAndEvaluate(`MATCH (x) SET x.age=35`, ectx)
	if err != nil {
		panic(err)
	}
	v, err = opencypher.ParseAndEvaluate("return andy.name,andy.age", ectx)
	name = v.Get().(opencypher.ResultSet).Rows[0]["1"].Get()
	age = v.Get().(opencypher.ResultSet).Rows[0]["2"].Get()
	fmt.Println(name, age) // Andy: 35
	v, err = opencypher.ParseAndEvaluate("return stephen.name,stephen.age", ectx)
	name = v.Get().(opencypher.ResultSet).Rows[0]["1"].Get()
	age = v.Get().(opencypher.ResultSet).Rows[0]["2"].Get()
	fmt.Println(name, age) // Stephen: 35
}
