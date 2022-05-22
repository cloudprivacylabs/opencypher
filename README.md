[![GoDoc](https://godoc.org/github.com/cloudprivacylabs/opencypher?status.svg)](https://godoc.org/github.com/cloudprivacylabs/opencypher)
[![Go Report Card](https://goreportcard.com/badge/github.com/cloudprivacylabs/opencypher)](https://goreportcard.com/report/github.com/cloudprivacylabs/opencypher)
[![Build Status](https://github.com/cloudprivacylabs/opencypher/actions/workflows/CI.yml/badge.svg?branch=main)](https://github.com/cloudprivacylabs/opencypher/actions/workflows/CI.yml)
# Embedded openCypher interpreter and labeled property graphs

This Go module contains a openCypher interpreter partial
implementation and a labeled property graph implementation. The
labeled property graph package can be used independently from the
openCypher library.

This Go module is part of the [Layered Schema
Architecture](https://layeredschemas.org).

## openCypher

At this point, this library provides partial support for openCypher
expressions. More support will be added as needed. unsupported
features include:

  * MERGE
  * CALL
  * Transactions
  * CSV related functions
  
openCypher expressions are evaluated using an evaluation context. 

### Create Nodes

See examples/create directory.

```
import (
	"fmt"

	"github.com/cloudprivacylabs/opencypher"
	"github.com/cloudprivacylabs/opencypher/graph"
)

func main() {
	grph := graph.NewOCGraph()
	ectx := opencypher.NewEvalContext(grph)
	_, err := opencypher.ParseAndEvaluate(`CREATE (n:Person), (m)`, ectx)
	if err != nil {
		panic(err)
	}
	v, err := opencypher.ParseAndEvaluate(`MATCH (x:Person) return x as person`, ectx)
	if err != nil {
		panic(err)
	}
	fmt.Println(v.Get().(opencypher.ResultSet).Rows[0]["person"])
}
```

### Evaluation Context

Variables defined in an expression will be in the evaluation context,
and can be used to affect the results of subsequent expression.

See examples/context.

```
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
}
```

### Querying and result sets

Using the example in https://neo4j.com/docs/cypher-manual/current/clauses/match/

```
	// Get all nodes
	ectx = opencypher.NewEvalContext(grph)
	res, err := opencypher.ParseAndEvaluate(`match (n) return n`, ectx)
	fmt.Println("match (n) return n:", res.Get().(opencypher.ResultSet).Rows)

	// Get all movies
	ectx = opencypher.NewEvalContext(grph)
	res, err = opencypher.ParseAndEvaluate(`match (n:Movie) return n.title`, ectx)
	fmt.Println("match (n:Movie) return n.title:", res.Get().(opencypher.ResultSet).Rows)

	// Get related node
	ectx = opencypher.NewEvalContext(grph)
	res, err = opencypher.ParseAndEvaluate(`match (director {name: 'Oliver Stone'}) --(movie:Movie) return movie.title`, ectx)
	fmt.Println("match (director {name: 'Oliver Stone'}) --(movie:Movie) return movie.title:", res.Get().(opencypher.ResultSet).Rows)
	ectx = opencypher.NewEvalContext(grph)
	res, err = opencypher.ParseAndEvaluate(`match (director {name: 'Oliver Stone'}) --> (movie:Movie) return movie.title`, ectx)
	fmt.Println("match (director {name: 'Oliver Stone'}) --> (movie:Movie) return movie.title:", res.Get().(opencypher.ResultSet).Rows)

	// Get relationship type
	ectx = opencypher.NewEvalContext(grph)
	res, err = opencypher.ParseAndEvaluate(`match (:Person {name: 'Oliver Stone'}) -[r]->(movie) return r`, ectx)
	fmt.Println("match (:Person {name: 'Oliver Stone'}) -[r]->(movie) return r:", res.Get().(opencypher.ResultSet).Rows)
```

### Values

Opencypher expressions return an object of type `Value`. `Value.Get`
returns the value contained in the value object. For most queries,
this value is of type `opencypher.ResultSet`. A `ResultSet` contains
`Rows` that are `map[string]Value` objects. If the query explicitly
names its columns, the map will contains those names as the
keys. Otherwise, the columns will be "1", "2", etc. 

The number of rows in the result set:

```
rs:=value.Get().(opencypher.ResultSet)
numResults:=len(rs.Rows)
```

Iterate the results:

```
for _,row := range resultSet.Rows {
   for colName, colValue := range row {
      // colName is the column name
      // colValue is an opencypher.Value object
      fmt.Println(colName,value.Get())
   }
}
```

## Labeled Property Graph

This labeled property graph package implements the openCypher model of
labeled property graphs. A labeled property graph (LPG) contains nodes
and directed edges between those nodes. Every node contains:

  * Labels: Set of string tokens that usually identify the type of the
    node,
  * Properties: Key-value pairs.
  
Every edge contains:
  * A label: String token that identifies a relationship, and
  * Properties: Key-value pairs.
  
A graph indexes its nodes and edges, so finding a node, or a pattern
usually does not involve iterating through all possibilities. 

Create a graph using `NewOCGraph` function:

```
g := graph.NewOCGraph()
// Create two nodes
n1 := g.NewNode([]string{"label1"},map[string]interface{}{"prop": "value1" })
n2 := g.NewNode([]string{"label2"},map[string]interface{}{"prop": "value2" })
// Connect the two nodes with an edge
edge:=g.NewEdge(n1,n2,"relatedTo",nil)
```

The LPG library uses an iterator model to address nodes and edges
because the underlying algorithm to collect nodes and edges mathcing a
certain criteria may depend on the existence of indexes. Both incoming
and outgoing edges of nodes are accessible:

```
for edges:=n1.GetEdges(graph.OutgoingEdge); edges.Next(); {
  edge:=edges.Edge()
  // edge.GetTo() and edge.GetFrom() are the adjacent nodes
}
```


The graph indexes nodes by label, so access to nodes using labels is
fast. You can add additional indexes on properties:

```
g := graph.NewOCGraph()
// Index all nodes with property 'prop'
g.AddNodePropertyIndex("prop")

// This access should be fast
nodes := g.GetNodesWithProperty("prop")

// This will go through all nodes
slowNodes:= g.GetNodesWithProperty("propWithoutIndex")
```

Graph library supports searching patterns. The following example
searches for the pattern that match 

```
(:label1) -[]->({prop:value})`
```

and returns the head nodes for every matching path:

```
pattern := graph.Pattern{ 
 // Node containing label 'label1'
 {
   Labels: graph.NewStringSet("label1"),
 },
 // Edge of length 1
 {
   Min: 1, 
   Max: 1,
 },
 // Node with property prop=value
 {
   Properties: map[string]interface{} {"prop":"value"},
 }}
nodes, err:=pattern.FindNodes(g,nil)
```

