package main

import (
	"fmt"

	"github.com/cloudprivacylabs/opencypher"
	"github.com/cloudprivacylabs/opencypher/graph"
)

// Examples are from: https://neo4j.com/docs/cypher-manual/current/clauses/merge/
func main() {

	getGraph := func() graph.Graph {
		g := graph.NewOCGraph()
		// Evaluation context knows the graph we are working on
		ectx := opencypher.NewEvalContext(g)

		_, err := opencypher.ParseAndEvaluate(`CREATE (charlieSheen :Person {bornIn:'New York',chauffeurName: 'John Brown',name:'Charlie Sheen'}),
		     (oliverStone :Person {bornIn: 'New York',chauffeurName: 'Bill White',name: 'Oliver Stone'}),
				 (michaelDouglas :Person {bornIn: 'New Jersey', chauffeurName: 'John Brown', name: 'Michael Douglas'}),
				 (martinSheen :Person {bornIn: 'Ohio',chauffeurName: 'Bob Brown',name: 'Martin Sheen'}),
				 (robReiner :Person {bornIn: 'New York',chauffeurName: 'Ted Green',name: 'Rob Reiner'}),
				 (ws :Movie {title: 'Wall Street'}),
				 (tap :Movie {title: 'The American President'})`, ectx)
		if err != nil {
			panic(err)
		}

		_, err = opencypher.ParseAndEvaluate(`CREATE (charlieSheen)-[:ACTED_IN]->(ws),
       (michaelDouglas)-[:ACTED_IN]->(ws),
		   (oliverStone)-[:ACTED_IN]->(ws),
		   (martinSheen)-[:ACTED_IN]->(ws),
		   (michaelDouglas)-[:ACTED_IN]->(tap),
		   (martinSheen)-[:ACTED_IN]->(tap),
		   (robReiner)-[:ACTED_IN]->(tap),
		   (charlieSheen)-[:FATHER]->(martinSheen)`, ectx)
		if err != nil {
			panic(err)
		}
		return g
	}

	// Merge single node with a label
	g := getGraph()
	res, err := opencypher.ParseAndEvaluate(`MERGE (robert:Critic) RETURN robert`, opencypher.NewEvalContext(g))
	if err != nil {
		panic(err)
	}
	fmt.Println(`MERGE (robert:Critic) RETURN robert:`, res)

	// Merge single node with properties
	//
	// Merging a single node with properties where not all properties
	// match any existing node.  A new node with the name 'Charlie
	// Sheen' will be created since not all properties matched the
	// existing 'Charlie Sheen' node.
	g = getGraph()
	res, err = opencypher.ParseAndEvaluate(`MERGE (charlie {name: 'Charlie Sheen', age: 10})
		RETURN charlie`, opencypher.NewEvalContext(g))
	if err != nil {
		panic(err)
	}
	fmt.Println(`MERGE (charlie {name: 'Charlie Sheen', age: 10}) RETURN charlie:`, res)

	// 2.3. Merge single node specifying both label and property
	// Merging a single node with both label and property matching an existing node.
	// 'Michael Douglas' will be matched and the name and bornIn properties returned.
	g = getGraph()
	res, err = opencypher.ParseAndEvaluate(`MERGE (michael:Person {name: 'Michael Douglas'})
		 RETURN michael.name, michael.bornIn`, opencypher.NewEvalContext(g))
	if err != nil {
		panic(err)
	}
	fmt.Println(`MERGE (michael:Person {name: 'Michael Douglas'}) RETURN michael.name, michael.bornIn:`, res)

	// 2.4. Merge single node derived from an existing node property For
	// some property 'p' in each bound node in a set of nodes, a single
	// new node is created for each unique value for 'p'.
	g = getGraph()
	res, err = opencypher.ParseAndEvaluate(`MATCH (person:Person)
	 MERGE (city:City {name: person.bornIn})
	 RETURN person.name, person.bornIn, city`, opencypher.NewEvalContext(g))
	if err != nil {
		panic(err)
	}
	fmt.Println(`MATCH (person:Person) MERGE (city:City {name: person.bornIn}) RETURN person.name, person.bornIn, city:`, res)

	//  Merge with ON CREATE
	// Merge a node and set properties if the node needs to be created.
	// The query creates the 'keanu' node and sets a timestamp on creation time.
	g = getGraph()
	res, err = opencypher.ParseAndEvaluate(` MERGE (keanu:Person {name: 'Keanu Reeves'})
		 ON CREATE
		   SET keanu.created = timestamp()
		 RETURN keanu.name, keanu.created`, opencypher.NewEvalContext(g))
	if err != nil {
		panic(err)
	}
	fmt.Println(`MERGE (keanu:Person {name: 'Keanu Reeves'}) ON CREATE SET keanu.created = timestamp() RETURN keanu.name, keanu.created:`, res)

	// 3.2. Merge with ON MATCH
	// Merging nodes and setting properties on found nodes.
	g = getGraph()
	res, err = opencypher.ParseAndEvaluate(`	MERGE (person:Person)
		ON MATCH
		   SET person.found = true
		 RETURN person.name, person.found`, opencypher.NewEvalContext(g))
	if err != nil {
		panic(err)
	}
	fmt.Println(`MERGE (person:Person) ON MATCH SET person.found = true RETURN person.name, person.found:`, res)

	// 4.1. Merge on a relationship MERGE can be used to match or create
	// a relationship.  'Charlie Sheen' had already been marked as
	// acting in 'Wall Street', so the existing relationship is found
	// and returned. Note that in order to match or create a
	// relationship when using MERGE, at least one bound node must be
	// specified, which is done via the MATCH clause in the above
	// example.

	g = getGraph()
	res, err = opencypher.ParseAndEvaluate(`	 MATCH
		   (charlie:Person {name: 'Charlie Sheen'}),
		   (wallStreet:Movie {title: 'Wall Street'})
		 MERGE (charlie)-[r:ACTED_IN]->(wallStreet)
		 RETURN charlie.name, type(r), wallStreet.title
	`, opencypher.NewEvalContext(g))
	if err != nil {
		panic(err)
	}
	fmt.Println(`MATCH (charlie:Person {name: 'Charlie Sheen'}), (wallStreet:Movie {title: 'Wall Street'}) MERGE (charlie)-[r:ACTED_IN]->(wallStreet) RETURN charlie.name, type(r), wallStreet.title:`, res)

	// 4.2. Merge on multiple relationships In our example graph,
	// 'Oliver Stone' and 'Rob Reiner' have never worked together. When
	// we try to MERGE a "movie between them, a new 'movie' node is
	// created.
	g = getGraph()
	res, err = opencypher.ParseAndEvaluate(`	 	 MATCH
		   (oliver:Person {name: 'Oliver Stone'}),
		   (reiner:Person {name: 'Rob Reiner'})
		 MERGE (oliver)-[:DIRECTED]->(movie:Movie)<-[:ACTED_IN]-(reiner)
		 RETURN movie
	`, opencypher.NewEvalContext(g))
	if err != nil {
		panic(err)
	}
	fmt.Println(`MATCH (oliver:Person {name: 'Oliver Stone'}), (reiner:Person {name: 'Rob Reiner'}) MERGE (oliver)-[:DIRECTED]->(movie:Movie)<-[:ACTED_IN]-(reiner) RETURN movie:`, res)

	// 4.3. Merge on an undirected relationship
	// MERGE can also be used with an undirected relationship. When it needs to create a new one, it will pick a direction.
	g = getGraph()
	res, err = opencypher.ParseAndEvaluate(`		 MATCH
		   (charlie:Person {name: 'Charlie Sheen'}),
		   (oliver:Person {name: 'Oliver Stone'})
		 MERGE (charlie)-[r:KNOWS]-(oliver)
		 RETURN r
	`, opencypher.NewEvalContext(g))
	if err != nil {
		panic(err)
	}
	fmt.Println(`MATCH (charlie:Person {name: 'Charlie Sheen'}), (oliver:Person {name: 'Oliver Stone'}) MERGE (charlie)-[r:KNOWS]-(oliver) RETURN r:`, res)

	// 4.4. Merge on a relationship between two existing nodes
	// MERGE can be used in conjunction with preceding MATCH and MERGE clauses to create a relationship between two bound nodes 'm' and 'n', where 'm' is returned by MATCH and 'n' is created or matched by the earlier MERGE.
	// This builds on the example from Merge single node derived from an existing node property. The second MERGE creates a BORN_IN relationship between each person and a city corresponding to the value of the person’s bornIn property. 'Charlie Sheen', 'Rob Reiner' and 'Oliver Stone' all have a BORN_IN relationship to the 'same' City node ('New York').

	g = getGraph()
	res, err = opencypher.ParseAndEvaluate(`		 	 MATCH (person:Person)
		 MERGE (city:City {name: person.bornIn})
		 MERGE (person)-[r:BORN_IN]->(city)
		 RETURN person.name, person.bornIn, city
	`, opencypher.NewEvalContext(g))
	if err != nil {
		panic(err)
	}
	fmt.Println(`MATCH (person:Person) MERGE (city:City {name: person.bornIn}) MERGE (person)-[r:BORN_IN]->(city) RETURN person.name, person.bornIn, city:`, res)

	// 4.5. Merge on a relationship between an existing node and a
	// merged node derived from a node property. MERGE can be used to
	// simultaneously create both a new node 'n' and a relationship
	// between a bound node 'm' and 'n'.
	g = getGraph()
	res, err = opencypher.ParseAndEvaluate(`	 MATCH (person:Person)
		 MERGE (person)-[r:HAS_CHAUFFEUR]->(chauffeur:Chauffeur {name: person.chauffeurName})
		 RETURN person.name, person.chauffeurName, chauffeur
	`, opencypher.NewEvalContext(g))
	if err != nil {
		panic(err)
	}
	fmt.Println(`MATCH (person:Person) MERGE (person)-[r:HAS_CHAUFFEUR]->(chauffeur:Chauffeur {name: person.chauffeurName}) RETURN person.name, person.chauffeurName, chauffeur:`, res)
}
