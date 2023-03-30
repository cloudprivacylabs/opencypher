package opencypher

import (
	"testing"

	"github.com/cloudprivacylabs/lpg/v2"
)

func TestPatternExpr(t *testing.T) {
	/*
	   (root) -> (:c1)
	   (root) -> (:c2) -> (:c3)
	*/

	g := lpg.NewGraph()
	n1 := g.NewNode([]string{"root"}, nil)
	n2 := g.NewNode([]string{"c1"}, nil)
	n3 := g.NewNode([]string{"c2"}, nil)
	n4 := g.NewNode([]string{"c3"}, nil)

	g.NewEdge(n1, n2, "n1n2", nil)
	g.NewEdge(n1, n3, "n1n3", nil)
	g.NewEdge(n3, n4, "n3n4", nil)

	pe, err := ParsePatternExpr(`(this)<-[]-()-[]->(target:c2)`)
	if err != nil {
		t.Error(err)
		return
	}
	nodes, err := pe.FindRelative(n2)
	if err != nil {
		t.Error(err)
	}
	if len(nodes) != 1 {
		t.Errorf("Expecting 1, got %d", len(nodes))
		return
	}
	if nodes[0] != n3 {
		t.Errorf("Expecting n3, got %s", nodes[0])
	}

	pe, err = ParsePatternExpr(`(this)<-[]-()-[]->(target)`)
	if err != nil {
		t.Error(err)
		return
	}
	nodes, err = pe.FindRelative(n2)
	if err != nil {
		t.Error(err)
	}
	if len(nodes) != 2 {
		t.Errorf("Expecting 2, got %d", len(nodes))
		return
	}
	if (nodes[0] != n3 && nodes[0] != n2) || (nodes[1] != n3 && nodes[1] != n2) {
		t.Errorf("Expecting n2 n3, got %v", nodes)
	}

	pe, err = ParsePatternExpr(`(this)-[]-()-[]-(target:c2)`)
	if err != nil {
		t.Error(err)
		return
	}
	nodes, err = pe.FindRelative(n2)
	if err != nil {
		t.Error(err)
	}
	if len(nodes) != 1 {
		t.Errorf("Expecting 1, got %d", len(nodes))
		return
	}
	if nodes[0] != n3 {
		t.Errorf("Expecting n3, got %s", nodes[0])
	}

}
