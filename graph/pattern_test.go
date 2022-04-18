package graph

import (
	"testing"
)

func TestPattern(t *testing.T) {
	graph := NewOCGraph()
	graph.index.NodePropertyIndex("key", graph)
	nodes := make([]Node, 0)
	for i := 0; i < 10; i++ {
		nodes = append(nodes, graph.NewNode([]string{"a"}, nil))
	}
	for i := 0; i < 8; i++ {
		graph.NewEdge(nodes[i], nodes[i+1], "label", nil)
	}
	nodes[5].SetProperty("key", "value")
	symbols := make(map[string]*PatternSymbol)
	pat := Pattern{
		{},
		{Min: 1, Max: 1},
		{Name: "nodes", Labels: nil, Properties: map[string]interface{}{"key": "value"}}}
	if _, i := pat.getFastestElement(graph, map[string]*PatternSymbol{}); i != 2 {
		t.Errorf("Expecting 2, got %d", i)
	}
	plan, err := pat.GetPlan(graph, symbols)
	if err != nil {
		t.Error(err)
		return
	}
	acc := &DefaultMatchAccumulator{}
	plan.Run(graph, symbols, acc)
	if _, ok := acc.Symbols[0]["nodes"].(Node); !ok {
		t.Errorf("Expecting one node, got: %v", acc)
	}

	pat = Pattern{
		{Labels: NewStringSet("bogus")},
		{Min: 1, Max: 1},
		{Name: "nodes", Properties: map[string]interface{}{"key": "value"}},
	}

	symbols = make(map[string]*PatternSymbol)
	plan, err = pat.GetPlan(graph, symbols)
	if err != nil {
		t.Error(err)
		return
	}
	acc = &DefaultMatchAccumulator{}
	plan.Run(graph, symbols, acc)
	if len(acc.Paths) != 0 {
		t.Errorf("Expecting 0 node, got: %+v", acc)
	}

	pat = Pattern{
		{},
		{},
		{Properties: map[string]interface{}{"key": "value2"}},
	}
	if _, i := pat.getFastestElement(graph, map[string]*PatternSymbol{}); i != 2 {
		t.Errorf("Expecting 2, got %d", i)
	}
	pat = Pattern{
		{Properties: map[string]interface{}{"key": "value2"}},
		{},
		{},
	}
	if _, i := pat.getFastestElement(graph, map[string]*PatternSymbol{}); i != 0 {
		t.Errorf("Expecting 0, got %d", i)
	}

}

func TestLoopPattern(t *testing.T) {
	graph := NewOCGraph()
	graph.index.NodePropertyIndex("key", graph)
	nodes := make([]Node, 0)
	for i := 0; i < 10; i++ {
		nodes = append(nodes, graph.NewNode([]string{"a"}, nil))
	}
	for i := 0; i < 8; i++ {
		graph.NewEdge(nodes[i], nodes[i+1], "label", nil)
	}
	symbols := make(map[string]*PatternSymbol)
	symbols["n"] = &PatternSymbol{}
	symbols["n"].Add(nodes[0])
	pat := Pattern{
		{Name: "n"},
		{Min: 1, Max: 1},
		{Name: "n"},
	}
	out := DefaultMatchAccumulator{}
	err := pat.Run(graph, symbols, &out)
	if err != nil {
		t.Error(err)
		return
	}
	if len(out.Symbols) > 0 {
		t.Errorf("Expecting 0 node, got: %+v", out)
	}

	// Create a loop
	graph.NewEdge(nodes[0], nodes[0], "label", nil)
	out = DefaultMatchAccumulator{}
	err = pat.Run(graph, symbols, &out)
	if err != nil {
		t.Error(err)
		return
	}
	if len(out.Symbols) != 1 {
		t.Errorf("Expecting 1 node, got: %v", out)
	}
}

func TestVariableLengthPath(t *testing.T) {
	graph := NewOCGraph()
	graph.index.NodePropertyIndex("key", graph)
	nodes := make([]Node, 0)
	for i := 0; i < 10; i++ {
		nodes = append(nodes, graph.NewNode([]string{"a"}, nil))
	}
	for i := 0; i < 8; i++ {
		graph.NewEdge(nodes[i], nodes[i+1], "label", nil)
	}

	nodes[1].SetProperty("property", "value")
	nodes[4].SetProperty("property", "value")

	symbols := make(map[string]*PatternSymbol)
	pat := Pattern{
		{Name: "n", Properties: map[string]interface{}{"property": "value"}},
		{Min: 1, Max: 1},
		{Properties: map[string]interface{}{"property": "value"}},
	}
	out := DefaultMatchAccumulator{}
	err := pat.Run(graph, symbols, &out)
	if err != nil {
		t.Error(err)
		return
	}
	if len(out.Paths) != 0 {
		t.Errorf("Expecting 0 nodes")
	}
	pat = Pattern{
		{Name: "n", Properties: map[string]interface{}{"property": "value"}},
		{Min: 1, Max: 4},
		{Properties: map[string]interface{}{"property": "value"}},
	}
	out = DefaultMatchAccumulator{}
	err = pat.Run(graph, symbols, &out)
	if err != nil {
		t.Error(err)
		return
	}
	if len(out.Paths[0].([]Edge)) != 3 {
		t.Errorf("Expecting 3 nodes: %+v", out)
	}

}
