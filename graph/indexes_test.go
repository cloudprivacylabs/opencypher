package graph

import (
	"fmt"
	"testing"
)

func TestNodeIndex(t *testing.T) {
	g := NewOCGraph()
	g.index.NodePropertyIndex("index", g)
	labels := []string{"a", "b", "c", "d", "e", "f"}
	data := make(map[string]struct{})
	for _, l := range labels {
		for i := 0; i < 10; i++ {
			g.NewNode([]string{l}, map[string]interface{}{"index": fmt.Sprint(i)})
			data[fmt.Sprintf("%s:%d", l, i)] = struct{}{}
		}
	}
	itr := g.index.GetIteratorForNodeProperty("index", "0")
	if size := itr.MaxSize(); size != 6 {
		t.Errorf("Expecting 6, got %d", size)
	}
	foundLabel := make(map[string]struct{})
	for itr.Next() {
		foundLabel[itr.Node().GetLabels().Slice()[0]] = struct{}{}
	}
	for _, l := range labels {
		if _, found := foundLabel[l]; !found {
			t.Errorf("Not found: %s", l)
		}
	}
}
