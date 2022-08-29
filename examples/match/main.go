package main

import (
	"fmt"

	"github.com/cloudprivacylabs/lpg"
	"github.com/cloudprivacylabs/opencypher"
)

// Examples are from: https://neo4j.com/docs/cypher-manual/current/clauses/match/
func main() {
	// Create an empty graph
	grph := lpg.NewGraph()
	// Evaluation context knows the graph we are working on
	ectx := opencypher.NewEvalContext(grph)
	// Create the example graph. First create nodes
	_, err := opencypher.ParseAndEvaluate(`CREATE (OliverStone :Person {name:'Oliver Stone'}),
(MichaelDouglas :Person {name:'Michael Douglas'}),
(CharlieSheen :Person {name:'Charlie Sheen'}),
(MartinSheen :Person {name:'Martin Sheen'}),
(RobReiner :Person {name:'Rob Reiner'}),
(WallStreet :Movie {title:'Wall Street'}),
(AP :Movie {title:'The American President'})`, ectx)
	if err != nil {
		panic(err)
	}
	// Then connect them. The variables are already bound
	_, err = opencypher.ParseAndEvaluate(`CREATE (OliverStone)-[:DIRECTED]->(WallStreet),
(MichaelDouglas)-[:ACTED_IN {role:'Gordon Gekko'}]->(WallStreet),
(MichaelDouglas)-[:ACTED_IN {role:'President Andrew Shepherd'}]->(AP),
(CharlieSheen)-[:ACTED_IN {role:'Bud Fox'}]->(WallStreet),
(MartinSheen)-[:ACTED_IN {role:'Carl Fox'}]->(WallStreet),
(MartinSheen)-[:ACTED_IN {role:'A.J. MacInerney'}]->(AP),
(RobReiner)-[:DIRECTED]->(AP)`, ectx)

	// Create a new context
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

	// Match on relationship type
	ectx = opencypher.NewEvalContext(grph)
	res, err = opencypher.ParseAndEvaluate(`MATCH (wallstreet:Movie {title: 'Wall Street'})<-[:ACTED_IN]-(actor)
	RETURN actor.name`, ectx)
	fmt.Println(`MATCH (wallstreet:Movie {title: 'Wall Street'})<-[:ACTED_IN]-(actor) RETURN actor.name:`, res.Get().(opencypher.ResultSet).Rows)

	// Multiple relationship types
	ectx = opencypher.NewEvalContext(grph)
	res, err = opencypher.ParseAndEvaluate(`MATCH (wallstreet:Movie {title: 'Wall Street'})<-[:ACTED_IN|:DIRECTED]-(person)
	RETURN person.name`, ectx)
	fmt.Println(`MATCH (wallstreet:Movie {title: 'Wall Street'})<-[:ACTED_IN|:DIRECTED]-(person) RETURN person.name:`, res.Get().(opencypher.ResultSet).Rows)

	// Relationship type and variable
	ectx = opencypher.NewEvalContext(grph)
	res, err = opencypher.ParseAndEvaluate(`MATCH (wallstreet {title: 'Wall Street'})<-[r:ACTED_IN]-(actor) RETURN r.role`, ectx)
	fmt.Println(`MATCH (wallstreet {title: 'Wall Street'})<-[r:ACTED_IN]-(actor) RETURN r.role:`, res.Get().(opencypher.ResultSet).Rows)

	// Multiple relationships
	ectx = opencypher.NewEvalContext(grph)
	res, err = opencypher.ParseAndEvaluate(`MATCH (charlie {name: 'Charlie Sheen'})-[:ACTED_IN]->(movie)<-[:DIRECTED]-(director)
	RETURN movie.title, director.name`, ectx)
	fmt.Println(`MATCH (charlie {name: 'Charlie Sheen'})-[:ACTED_IN]->(movie)<-[:DIRECTED]-(director)
	RETURN movie.title, director.name:`, res.Get().(opencypher.ResultSet).Rows)

	// Variable length relationships
	ectx = opencypher.NewEvalContext(grph)
	res, err = opencypher.ParseAndEvaluate(`MATCH (charlie {name: 'Charlie Sheen'})-[:ACTED_IN*1..3]-(movie:Movie) RETURN movie.title`, ectx)
	fmt.Println(`MATCH (charlie {name: 'Charlie Sheen'})-[:ACTED_IN*1..3]-(movie:Movie) RETURN movie.title:`, res.Get().(opencypher.ResultSet).Rows)

	// Variable length relationships with multiple relationship types
	ectx = opencypher.NewEvalContext(grph)
	res, err = opencypher.ParseAndEvaluate(`MATCH (charlie {name: 'Charlie Sheen'})-[:ACTED_IN|DIRECTED*2]-(person:Person) RETURN person.name`, ectx)
	fmt.Println(`MATCH (charlie {name: 'Charlie Sheen'})-[:ACTED_IN|DIRECTED*2]-(person:Person) RETURN person.name:`, res.Get().(opencypher.ResultSet).Rows)
}
