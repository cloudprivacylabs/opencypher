[![GoDoc](https://godoc.org/github.com/cloudprivacylabs/opencypher?status.svg)](https://godoc.org/github.com/cloudprivacylabs/opencypher)
[![Go Report Card](https://goreportcard.com/badge/github.com/cloudprivacylabs/opencypher)](https://goreportcard.com/report/github.com/cloudprivacylabs/opencypher)
[![Build Status](https://github.com/cloudprivacylabs/opencypher/actions/workflows/CI.yml/badge.svg?branch=main)](https://github.com/cloudprivacylabs/opencypher/actions/workflows/CI.yml)
# Embedded openCypher interpreter

openCypher is a query language for labeled property graphs. This Go
module contains an openCypher interpreter (partial) that works on the
Go LPG implementation given in
https://github.com/cloudprivacylabs/lpg.

More information on openCypher can be found here:

https://opencypher.org/

## openCypher

At this point, this library provides partial support for openCypher
expressions. More support will be added as needed. 
  
openCypher expressions are evaluated using an evaluation context. 

### Create Nodes

See examples/create directory.

```
import (
	"fmt"

	"github.com/cloudprivacylabs/opencypher"
	"github.com/cloudprivacylabs/lpg"
)

func main() {
	grph := graph.NewGraph()
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
	grph := graph.NewGraph()
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
	age, _ := v.Get().(opencypher.ResultSet).Rows[0]["1"].Get().(*graph.Node).GetProperty("age")
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

The return type of `Value.Get` is one of the following:
  * int
  * float64
  * bool
  * string
  * Duration
  * Date
  * Time
  * LocalDateTime
  * LocalTime
  * []Value
  * map[string]Value
  * lpg.StringSet
  * *lpg.Node
  * []*lpg.Edge
  * ResultSet


This Go module is part of the [Layered Schema
Architecture](https://layeredschemas.org).

