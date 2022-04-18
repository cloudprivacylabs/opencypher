package graph

import (
	"fmt"
	"testing"
)

func TestEdgeMap(t *testing.T) {
	m := EdgeMap{}
	labels := []string{"a", "b", "c", "d", "e", "f"}
	data := make(map[string]struct{})
	for _, l := range labels {
		for i := 0; i < 10; i++ {
			edge := &OCEdge{label: l}
			edge.Properties = make(Properties)
			edge.Properties["index"] = i
			m.Add(edge)
			data[fmt.Sprintf("%s:%d", l, i)] = struct{}{}
		}
	}
	// itr: 60 items
	itr := m.Iterator()
	found := make(map[string]struct{})
	for itr.Next() {
		edge := itr.Edge().(*OCEdge)
		found[fmt.Sprintf("%s:%d", edge.label, edge.Properties["index"])] = struct{}{}
	}
	if len(found) != len(data) {
		t.Errorf("found: %v", found)
	}

	// Label-based iteration
	for _, label := range labels {
		itr = m.IteratorLabel(label)
		found = make(map[string]struct{})
		for itr.Next() {
			edge := itr.Edge().(*OCEdge)
			if edge.label != label {
				t.Errorf("Expecting %s got %+v", label, edge)
			}
			found[fmt.Sprint(edge.Properties["index"])] = struct{}{}
		}
		if len(found) != 10 {
			t.Errorf("10 entries were expected, got %v", found)
		}
	}

	itr = m.IteratorAnyLabel(NewStringSet("a", "c", "e", "g"))
	found = make(map[string]struct{})
	for itr.Next() {
		edge := itr.Edge().(*OCEdge)
		if edge.label != "a" && edge.label != "c" && edge.label != "e" {
			t.Errorf("Unexpected label: %s", edge.label)
		}
		found[fmt.Sprintf("%s:%d", edge.label, edge.Properties["index"])] = struct{}{}
	}
	if len(found) != 30 {
		t.Errorf("Expecting 30, got %v", found)
	}
}

func TestNodeMap(t *testing.T) {
	m := NodeMap{}
	labels := [][]string{{"a"}, {"b", "c", "d"}, {"e", "f"}}
	data := make(map[string]struct{})
	for _, l := range labels {
		for i := 0; i < 10; i++ {
			node := &OCNode{labels: NewStringSet(l...)}
			node.Properties = make(Properties)
			node.Properties["index"] = i
			m.Add(node)
			data[fmt.Sprintf("%d:%d", len(l), i)] = struct{}{}
		}
	}
	// itr: 30 items
	itr := m.Iterator()
	found := make(map[string]struct{})
	for itr.Next() {
		node := itr.Node().(*OCNode)
		found[fmt.Sprintf("%d:%d", len(node.labels), node.Properties["index"])] = struct{}{}
	}
	if len(found) != len(data) {
		t.Errorf("found: %v", found)
	}

	// Label-based iteration
	for _, label := range labels {
		itr = m.IteratorAllLabels(NewStringSet(label...))
		found = make(map[string]struct{})
		for itr.Next() {
			node := itr.Node().(*OCNode)
			if !node.labels.HasAll(label...) {
				t.Errorf("Expecting %v got %+v", label, node)
			}
			found[fmt.Sprint(node.Properties["index"])] = struct{}{}
		}
		if len(found) != 10 {
			t.Errorf("10 entries were expected, got %v", found)
		}
	}
}
