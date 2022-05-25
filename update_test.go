package opencypher

import (
	"testing"

	"github.com/cloudprivacylabs/opencypher/graph"
)

func TestUpdate(t *testing.T) {
	var stefan, george, swedish, peter graph.Node
	getGraph := func() graph.Graph {
		g := graph.NewOCGraph()
		// Examples from neo4j documentation
		stefan = g.NewNode(nil, map[string]interface{}{"name": "Stefan"})
		george = g.NewNode(nil, map[string]interface{}{"name": "George"})
		swedish = g.NewNode([]string{"Swedish"}, map[string]interface{}{"name": "Andy", "age": 36, "hungry": true})
		peter = g.NewNode(nil, map[string]interface{}{"name": "Peter", "age": 34})
		g.NewEdge(stefan, swedish, "KNOWS", nil)
		g.NewEdge(swedish, peter, "KNOWS", nil)
		g.NewEdge(george, peter, "KNOWS", nil)
		return g
	}

	// Set a property
	g := getGraph()
	v, err := ParseAndEvaluate(`match (n {name: 'Andy'}) set n.surname='Taylor' return n.name,n.surname`, NewEvalContext(g))
	if err != nil {
		t.Error(err)
	}
	if len(v.Get().(ResultSet).Rows) != 1 {
		t.Errorf("Rows expected to be 1: %d", len(v.Get().(ResultSet).Rows))
	}
	if v, _ := swedish.GetProperty("surname"); v != "Taylor" {
		t.Errorf("Wrong update: %+v", swedish)
	}

	// Remove a property
	g = getGraph()
	v, err = ParseAndEvaluate(`MATCH (n {name: 'Andy'})
SET n.name = null
RETURN n.name, n.age`, NewEvalContext(g))
	if err != nil {
		t.Error(err)
	}
	if _, ok := swedish.GetProperty("name"); ok {
		t.Errorf("Property still exists")
	}

	// Copy properties
	g = getGraph()
	v, err = ParseAndEvaluate(`match (at {name: 'Andy'}), (pn {name: 'Peter'}) set at=pn return at.name,at.age,pn.name,pn.age`, NewEvalContext(g))
	if err != nil {
		t.Error(err)
	}
	n, _ := swedish.GetProperty("name")
	a, _ := swedish.GetProperty("age")
	if n != "Peter" || a != 34 {
		t.Errorf("Wrong update prop: %v", swedish)
	}

	// Replace properties
	g = getGraph()
	v, err = ParseAndEvaluate(`MATCH (p {name: 'Peter'})
SET p = {name: 'Peter Smith', position: 'Entrepreneur'}
RETURN p.name, p.age, p.position`, NewEvalContext(g))
	if err != nil {
		t.Error(err)
	}
	if n, _ = peter.GetProperty("name"); n != "Peter Smith" {
		t.Errorf("Wrong name")
	}
	if n, _ = peter.GetProperty("position"); n != "Entrepreneur" {
		t.Errorf("Wrong position")
	}

	// Remove all properties
	g = getGraph()
	v, err = ParseAndEvaluate(`MATCH (p {name: 'Peter'})
SET p = {}
RETURN p.name, p.age`, NewEvalContext(g))
	if err != nil {
		t.Error(err)
	}
	if _, ok := peter.GetProperty("name"); ok {
		t.Errorf("Name still exsist")
	}
	if _, ok := peter.GetProperty("age"); ok {
		t.Errorf("Age still exists")
	}

	// Mutate specific props
	g = getGraph()
	v, err = ParseAndEvaluate(`MATCH (p {name: 'Peter'})
SET p += {age: 38, hungry: true, position: 'Entrepreneur'}
RETURN p.name, p.age, p.hungry, p.position`, NewEvalContext(g))
	if err != nil {
		t.Error(err)
	}
	if x, _ := peter.GetProperty("age"); x != 38 {
		t.Errorf("Wrong age")
	}
	if x, _ := peter.GetProperty("hungry"); x != true {
		t.Errorf("Wrong hungry")
	}
	if x, _ := peter.GetProperty("position"); x != "Entrepreneur" {
		t.Errorf("wrong position")
	}

	// Set a label
	g = getGraph()
	v, err = ParseAndEvaluate(`MATCH (n {name: 'Stefan'})
SET n:German
RETURN n.name, labels(n) AS labels`, NewEvalContext(g))
	if err != nil {
		t.Error(err)
	}
	if !stefan.HasLabel("German") {
		t.Errorf("Cannot set label")
	}
}

func TestDelete(t *testing.T) {
	var andy, unk, timothy, peter graph.Node
	getGraph := func() graph.Graph {
		g := graph.NewOCGraph()
		// Examples from neo4j documentation
		andy = g.NewNode([]string{"Person"}, map[string]interface{}{"name": "Andy", "age": 36})
		unk = g.NewNode([]string{"Person"}, map[string]interface{}{"name": "UNKNOWN"})
		timothy = g.NewNode([]string{"Person"}, map[string]interface{}{"name": "Timothy", "age": 25})
		peter = g.NewNode([]string{"Person"}, map[string]interface{}{"name": "Peter", "age": 34})
		g.NewEdge(andy, timothy, "KNOWS", nil)
		g.NewEdge(andy, peter, "KNOWS", nil)
		return g
	}

	_ = unk
	// Delete single node
	g := getGraph()
	_, err := ParseAndEvaluate(`MATCH (n:Person {name: 'UNKNOWN'})
DELETE n`, NewEvalContext(g))
	if err != nil {
		t.Error(err)
	}
	if g.NumNodes() != 3 {
		t.Errorf("Excess nodes")
	}

	// Delete all nodes
	g = getGraph()
	_, err = ParseAndEvaluate(`MATCH (n)
DETACH DELETE n`, NewEvalContext(g))
	if err != nil {
		t.Error(err)
	}
	if g.NumNodes() != 0 {
		t.Errorf("Excess nodes")
	}

	// Delete node with links
	g = getGraph()
	_, err = ParseAndEvaluate(`	MATCH (n {name: 'Andy'})
DETACH DELETE n`, NewEvalContext(g))
	if err != nil {
		t.Error(err)
	}
	if g.NumNodes() != 3 {
		t.Errorf("Excess nodes")
	}

	// Delete relationships
	g = getGraph()
	_, err = ParseAndEvaluate(`MATCH (n {name: 'Andy'})-[r:KNOWS]->()
DELETE r`, NewEvalContext(g))
	if err != nil {
		t.Error(err)
	}
	if g.NumNodes() != 4 {
		t.Errorf("Wrong number of nodes")
	}
	if andy.GetEdges(graph.OutgoingEdge).Next() {
		t.Errorf("Still connected")
	}
}

func TestRemove(t *testing.T) {
	var andy, timothy, peter graph.Node
	getGraph := func() graph.Graph {
		g := graph.NewOCGraph()
		// Examples from neo4j documentation
		andy = g.NewNode([]string{"Swedish"}, map[string]interface{}{"name": "Andy", "age": 36})
		timothy = g.NewNode([]string{"Swedish"}, map[string]interface{}{"name": "Timothy", "age": 25})
		peter = g.NewNode([]string{"Swedish", "German"}, map[string]interface{}{"name": "Peter", "age": 34})
		g.NewEdge(andy, timothy, "KNOWS", nil)
		g.NewEdge(andy, peter, "KNOWS", nil)
		return g
	}

	// Remove a property
	g := getGraph()
	_, err := ParseAndEvaluate(`	MATCH (a {name: 'Andy'})
REMOVE a.age
RETURN a.name, a.age`, NewEvalContext(g))
	if err != nil {
		t.Error(err)
	}
	if _, ok := andy.GetProperty("age"); ok {
		t.Errorf("Cannot remove age")
	}

	// Remove labels
	g = getGraph()
	_, err = ParseAndEvaluate(`MATCH (n {name: 'Peter'})
REMOVE n:German:Swedish
RETURN n.name, labels(n)`, NewEvalContext(g))
	if err != nil {
		t.Error(err)
	}
	if len(peter.GetLabels()) != 0 {
		t.Errorf("Has labels")
	}
}

func TestCreate(t *testing.T) {

	// Create one node
	g := graph.NewOCGraph()
	_, err := ParseAndEvaluate(`create (n)`, NewEvalContext(g))
	if err != nil {
		t.Error(err)
	}
	if g.NumNodes() != 1 {
		t.Errorf("Did not create a node")
	}

	// Create multiple nodes
	g = graph.NewOCGraph()
	_, err = ParseAndEvaluate(`CREATE (n), (m)`, NewEvalContext(g))
	if err != nil {
		t.Error(err)
	}
	if g.NumNodes() != 2 {
		t.Errorf("Did not create two nodes")
	}

	// Create node with label
	g = graph.NewOCGraph()
	_, err = ParseAndEvaluate(`CREATE (n:Person)`, NewEvalContext(g))
	if err != nil {
		t.Error(err)
	}
	if g.NumNodes() != 1 {
		t.Errorf("Did not create person node")
	}
	nodes := g.GetNodes()
	nodes.Next()
	if nodes.Node().GetLabels().Slice()[0] != "Person" {
		t.Errorf("Wrong labels")
	}
	// Create node with labels
	g = graph.NewOCGraph()
	_, err = ParseAndEvaluate(`CREATE (n:Person:Swedish)`, NewEvalContext(g))
	if err != nil {
		t.Error(err)
	}
	if g.NumNodes() != 1 {
		t.Errorf("Did not create person node")
	}
	nodes = g.GetNodes()
	nodes.Next()
	node := nodes.Node()
	if !node.GetLabels().Has("Person") || !node.GetLabels().Has("Swedish") {
		t.Errorf("Wrong labels")
	}

	// Create with properties
	g = graph.NewOCGraph()
	ret, err := ParseAndEvaluate(`CREATE (a:Person {name: 'Andy'})
RETURN a.name`, NewEvalContext(g))
	if err != nil {
		t.Error(err)
	}

	nodes = g.GetNodes()
	nodes.Next()
	node = nodes.Node()
	if !node.GetLabels().Has("Person") {
		t.Errorf("Wrong labels")
	}
	if s, _ := node.GetProperty("name"); s != "Andy" {
		t.Errorf("Wrong name")
	}
	if ret.Get().(ResultSet).Rows[0]["1"].Get() != "Andy" {
		t.Errorf("Wrong result: %v", ret)
	}

	// Create relationship
	g = graph.NewOCGraph()
	nodea := g.NewNode([]string{"Person"}, map[string]interface{}{"name": "A"})
	nodeb := g.NewNode([]string{"Person"}, map[string]interface{}{"name": "B"})
	ret, err = ParseAndEvaluate(`MATCH
  (a:Person),
  (b:Person)
WHERE a.name = 'A' AND b.name = 'B'
CREATE (a)-[r:RELTYPE]->(b)
RETURN r`, NewEvalContext(g))
	if err != nil {
		t.Error(err)
	}

	// There must be an edge between a and b
	edges := nodea.GetEdges(graph.OutgoingEdge)
	if !edges.Next() {
		t.Errorf("No edge")
	}
	edge := edges.Edge()
	if edge.GetTo() != nodeb {
		t.Errorf("Wrong target")
	}
	if edge.GetLabel() != "RELTYPE" {
		t.Errorf("Wronglabel")
	}

	// Create a relationship and set properties
	g = graph.NewOCGraph()
	nodea = g.NewNode([]string{"Person"}, map[string]interface{}{"name": "A"})
	nodeb = g.NewNode([]string{"Person"}, map[string]interface{}{"name": "B"})
	ret, err = ParseAndEvaluate(`MATCH
  (a:Person),
  (b:Person)
WHERE a.name = 'A' AND b.name = 'B'
CREATE (a)-[r:RELTYPE {name: a.name + '<->' + b.name}]->(b)
RETURN  r.name`, NewEvalContext(g))
	if err != nil {
		t.Error(err)
	}

	// There must be an edge between a and b
	edges = nodea.GetEdges(graph.OutgoingEdge)
	if !edges.Next() {
		t.Errorf("No edge")
	}
	edge = edges.Edge()
	if edge.GetTo() != nodeb {
		t.Errorf("Wrong target")
	}
	if edge.GetLabel() != "RELTYPE" {
		t.Errorf("Wronglabel")
	}
	if ret.Get().(ResultSet).Rows[0]["1"].Get() != "A<->B" {
		t.Errorf("Wrong name: %v", ret)
	}

	// Create full path
	g = graph.NewOCGraph()
	ret, err = ParseAndEvaluate(`CREATE p = (andy {name:'Andy'})-[:WORKS_AT]->(neo)<-[:WORKS_AT]-(michael {name: 'Michael'})
RETURN p`, NewEvalContext(g))
	if err != nil {
		t.Error(err)
	}
	if g.NumNodes() != 3 {
		t.Errorf("Not 3 nodes")
	}

}

func TestMerge(t *testing.T) {

	var charlieSheen, oliverStone, michaelDouglas, martinSheen, robReiner graph.Node
	var ws, tap graph.Node

	getGraph := func() graph.Graph {
		g := graph.NewOCGraph()
		// Examples from neo4j documentation
		charlieSheen = g.NewNode([]string{"Person"}, map[string]interface{}{
			"bornIn":        "New York",
			"chauffeurName": "John Brown",
			"name":          "Charlie Sheen",
		})
		oliverStone = g.NewNode([]string{"Person"}, map[string]interface{}{
			"bornIn":        "New York",
			"chauffeurName": "Bill White",
			"name":          "Oliver Stone",
		})
		michaelDouglas = g.NewNode([]string{"Person"}, map[string]interface{}{
			"bornIn":        "New Jersey",
			"chauffeurName": "John Brown",
			"name":          "Michael Douglas",
		})
		martinSheen = g.NewNode([]string{"Person"}, map[string]interface{}{
			"bornIn":        "Ohio",
			"chauffeurName": "Bob Brown",
			"name":          "Martin Sheen",
		})
		robReiner = g.NewNode([]string{"Person"}, map[string]interface{}{
			"bornIn":        "New York",
			"chauffeurName": "Ted Green",
			"name":          "Rob Reiner",
		})
		ws = g.NewNode([]string{"Movie"}, map[string]interface{}{
			"title": "Wall Street",
		})
		tap = g.NewNode([]string{"Movie"}, map[string]interface{}{
			"title": "The American President",
		})
		return g
	}
	_ = charlieSheen
	_ = oliverStone
	_ = michaelDouglas
	_ = martinSheen
	_ = robReiner
	_ = ws
	_ = tap

	// Merge single node with a label
	g := getGraph()
	n := g.NumNodes()
	res, err := ParseAndEvaluate(`MERGE (robert:Critic) RETURN robert`, NewEvalContext(g))
	if err != nil {
		t.Error(err)
	}
	if g.NumNodes() != n+1 {
		t.Errorf("No new nodes")
	}

	// Merge single node with properties
	//
	// Merging a single node with properties where not all properties
	// match any existing node.  A new node with the name 'Charlie
	// Sheen' will be created since not all properties matched the
	// existing 'Charlie Sheen' node.
	g = getGraph()
	n = g.NumNodes()
	res, err = ParseAndEvaluate(`MERGE (charlie {name: 'Charlie Sheen', age: 10})
	RETURN charlie`, NewEvalContext(g))
	if err != nil {
		t.Error(err)
	}
	if g.NumNodes() != n+1 {
		t.Errorf("No new nodes")
	}

	// 2.3. Merge single node specifying both label and property
	// Merging a single node with both label and property matching an existing node.
	// 'Michael Douglas' will be matched and the name and bornIn properties returned.
	g = getGraph()
	n = g.NumNodes()
	res, err = ParseAndEvaluate(`MERGE (michael:Person {name: 'Michael Douglas'})
	 RETURN michael.name, michael.bornIn`, NewEvalContext(g))
	if err != nil {
		t.Error(err)
	}
	if g.NumNodes() != n {
		t.Errorf("No new nodes expected")
	}
	if res.Get().(ResultSet).Rows[0]["1"].Get() != "Michael Douglas" ||
		res.Get().(ResultSet).Rows[0]["2"].Get() != "New Jersey" {
		t.Errorf("Wrong result")
	}

	// 2.4. Merge single node derived from an existing node property For
	// some property 'p' in each bound node in a set of nodes, a single
	// new node is created for each unique value for 'p'.
	g = getGraph()
	n = g.NumNodes()
	res, err = ParseAndEvaluate(`MATCH (person:Person)
 MERGE (city:City {name: person.bornIn})
 RETURN person.name, person.bornIn, city`, NewEvalContext(g))
	if err != nil {
		t.Error(err)
	}
	if g.NumNodes() != n+3 {
		t.Errorf("3 new nodes expected, got %d %+v", g.NumNodes(), res.Get())
	}

	//  Merge with ON CREATE
	// Merge a node and set properties if the node needs to be created.
	// The query creates the 'keanu' node and sets a timestamp on creation time.
	g = getGraph()
	n = g.NumNodes()
	res, err = ParseAndEvaluate(` MERGE (keanu:Person {name: 'Keanu Reeves'})
	 ON CREATE
	   SET keanu.created = timestamp()
	 RETURN keanu.name, keanu.created`, NewEvalContext(g))
	if err != nil {
		t.Error(err)
	}
	if g.NumNodes() != n+1 {
		t.Errorf("1 new nodes expected, got %d %+v", g.NumNodes(), res.Get())
	}
	if res.Get().(ResultSet).Rows[0]["1"].Get() != "Keanu Reeves" ||
		res.Get().(ResultSet).Rows[0]["2"].Get() == nil {
		t.Errorf("Wrong data: %+v", res.Get().(ResultSet).Rows[0])
	}
	// 3.2. Merge with ON MATCH
	// Merging nodes and setting properties on found nodes.
	g = getGraph()
	n = g.NumNodes()
	res, err = ParseAndEvaluate(`	MERGE (person:Person)
	ON MATCH
	   SET person.found = true
	 RETURN person.name, person.found`, NewEvalContext(g))
	if err != nil {
		t.Error(err)
	}
	if v, _ := charlieSheen.GetProperty("found"); v != true {
		t.Errorf("not found")
	}
	if v, _ := martinSheen.GetProperty("found"); v != true {
		t.Errorf("not found")
	}
	if v, _ := robReiner.GetProperty("found"); v != true {
		t.Errorf("not found")
	}
	if v, _ := michaelDouglas.GetProperty("found"); v != true {
		t.Errorf("not found")
	}
	if v, _ := oliverStone.GetProperty("found"); v != true {
		t.Errorf("not found")
	}

	// 4. Merge relationships
	// 4.1. Merge on a relationship
	// MERGE can be used to match or create a relationship.

	// Query
	// Cypher
	// Copy to Clipboard
	// Run in Neo4j Browser
	// MATCH
	//   (charlie:Person {name: 'Charlie Sheen'}),
	//   (wallStreet:Movie {title: 'Wall Street'})
	// MERGE (charlie)-[r:ACTED_IN]->(wallStreet)
	// RETURN charlie.name, type(r), wallStreet.title
	// 'Charlie Sheen' had already been marked as acting in 'Wall Street', so the existing relationship is found and returned. Note that in order to match or create a relationship when using MERGE, at least one bound node must be specified, which is done via the MATCH clause in the above example.

	// Table 9. Result
	// charlie.name	type(r)	wallStreet.title
	// "Charlie Sheen"

	// "ACTED_IN"

	// "Wall Street"

	// Rows: 1

	// 4.2. Merge on multiple relationships
	// Query
	// Cypher
	// Copy to Clipboard
	// Run in Neo4j Browser
	// MATCH
	//   (oliver:Person {name: 'Oliver Stone'}),
	//   (reiner:Person {name: 'Rob Reiner'})
	// MERGE (oliver)-[:DIRECTED]->(movie:Movie)<-[:ACTED_IN]-(reiner)
	// RETURN movie
	// In our example graph, 'Oliver Stone' and 'Rob Reiner' have never worked together. When we try to MERGE a "movie between them, Neo4j will not use any of the existing movies already connected to either person. Instead, a new 'movie' node is created.

	// Table 10. Result
	// movie
	// Node[7]{}

	// Rows: 1
	// Nodes created: 1
	// Relationships created: 2
	// Labels added: 1

	// 4.3. Merge on an undirected relationship
	// MERGE can also be used with an undirected relationship. When it needs to create a new one, it will pick a direction.

	// Query
	// Cypher
	// Copy to Clipboard
	// Run in Neo4j Browser
	// MATCH
	//   (charlie:Person {name: 'Charlie Sheen'}),
	//   (oliver:Person {name: 'Oliver Stone'})
	// MERGE (charlie)-[r:KNOWS]-(oliver)
	// RETURN r
	// As 'Charlie Sheen' and 'Oliver Stone' do not know each other this MERGE query will create a KNOWS relationship between them. The direction of the created relationship is arbitrary.

	// Table 11. Result
	// r
	// :KNOWS[8]{}

	// Rows: 1
	// Relationships created: 1

	// 4.4. Merge on a relationship between two existing nodes
	// MERGE can be used in conjunction with preceding MATCH and MERGE clauses to create a relationship between two bound nodes 'm' and 'n', where 'm' is returned by MATCH and 'n' is created or matched by the earlier MERGE.

	// Query
	// Cypher
	// Copy to Clipboard
	// Run in Neo4j Browser
	// MATCH (person:Person)
	// MERGE (city:City {name: person.bornIn})
	// MERGE (person)-[r:BORN_IN]->(city)
	// RETURN person.name, person.bornIn, city
	// This builds on the example from Merge single node derived from an existing node property. The second MERGE creates a BORN_IN relationship between each person and a city corresponding to the value of the person’s bornIn property. 'Charlie Sheen', 'Rob Reiner' and 'Oliver Stone' all have a BORN_IN relationship to the 'same' City node ('New York').

	// Table 12. Result
	// person.name	person.bornIn	city
	// "Charlie Sheen"

	// "New York"

	// Node[7]{name:"New York"}

	// "Martin Sheen"

	// "Ohio"

	// Node[8]{name:"Ohio"}

	// "Michael Douglas"

	// "New Jersey"

	// Node[9]{name:"New Jersey"}

	// "Oliver Stone"

	// "New York"

	// Node[7]{name:"New York"}

	// "Rob Reiner"

	// "New York"

	// Node[7]{name:"New York"}

	// Rows: 5
	// Nodes created: 3
	// Relationships created: 5
	// Properties set: 3
	// Labels added: 3

	// 4.5. Merge on a relationship between an existing node and a merged node derived from a node property
	// MERGE can be used to simultaneously create both a new node 'n' and a relationship between a bound node 'm' and 'n'.

	// Query
	// Cypher
	// Copy to Clipboard
	// Run in Neo4j Browser
	// MATCH (person:Person)
	// MERGE (person)-[r:HAS_CHAUFFEUR]->(chauffeur:Chauffeur {name: person.chauffeurName})
	// RETURN person.name, person.chauffeurName, chauffeur
	// As MERGE found no matches — in our example graph, there are no nodes labeled with Chauffeur and no HAS_CHAUFFEUR relationships — MERGE creates five nodes labeled with Chauffeur, each of which contains a name property whose value corresponds to each matched Person node’s chauffeurName property value. MERGE also creates a HAS_CHAUFFEUR relationship between each Person node and the newly-created corresponding Chauffeur node. As 'Charlie Sheen' and 'Michael Douglas' both have a chauffeur with the same name — 'John Brown' — a new node is created in each case, resulting in 'two' Chauffeur nodes having a name of 'John Brown', correctly denoting the fact that even though the name property may be identical, these are two separate people. This is in contrast to the example shown above in Merge on a relationship between two existing nodes, where we used the first MERGE to bind the City nodes to prevent them from being recreated (and thus duplicated) in the second MERGE.

	// Table 13. Result
	// person.name	person.chauffeurName	chauffeur
	// "Charlie Sheen"

	// "John Brown"

	// Node[7]{name:"John Brown"}

	// "Martin Sheen"

	// "Bob Brown"

	// Node[8]{name:"Bob Brown"}

	// "Michael Douglas"

	// "John Brown"

	// Node[9]{name:"John Brown"}

	// "Oliver Stone"

	// "Bill White"

	// Node[10]{name:"Bill White"}

	// "Rob Reiner"

	// "Ted Green"

	// Node[11]{name:"Ted Green"}

	// Rows: 5
	// Nodes created: 5
	// Relationships created: 5
	// Properties set: 5
	// Labels added: 5

}
