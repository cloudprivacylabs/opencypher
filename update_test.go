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
