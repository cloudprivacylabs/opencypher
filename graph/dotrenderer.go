package graph

import (
	"fmt"
	"io"
)

// DOTRenderer renders a graph in Graphviz dot format
type DOTRenderer struct {
	// NodeRenderer renderes a node. If the node is to be excluded, returns false.
	NodeRenderer func(string, Node, io.Writer) (bool, error)
	// EdgeRenderer renders an edge. The from and to nodes are rendered
	// if this is called. If the edge is to be excluded, returns false
	EdgeRenderer func(fromID string, toID string, edge Edge, w io.Writer) (bool, error)
}

// RenderNode renders a node. If node renderer is not set, calls the default renderer
func (d DOTRenderer) RenderNode(ID string, node Node, w io.Writer) (bool, error) {
	if d.NodeRenderer == nil {
		return true, DefaultDOTNodeRender(ID, node, w)
	}
	return d.NodeRenderer(ID, node, w)
}

// RenderEdge renders an edge. If edge renderer is not set, call the default rendeded
func (d DOTRenderer) RenderEdge(fromID, toID string, edge Edge, w io.Writer) (bool, error) {
	if d.EdgeRenderer == nil {
		return true, DefaultDOTEdgeRender(fromID, toID, edge, w)
	}
	return d.EdgeRenderer(fromID, toID, edge, w)
}

// DefaultDOTNodeRender renders the node with the given ID. If the
// node has a label, it uses that label, otherwise node is not
// labeled.
func DefaultDOTNodeRender(ID string, node Node, w io.Writer) error {
	_, err := fmt.Fprintf(w, "  %s;\n", ID)
	return err
}

// DefaultDOTEdgeRender renders the edge with a label if there is
// one, or without a label if there is not a label.
func DefaultDOTEdgeRender(fromNode, toNode string, edge Edge, w io.Writer) error {
	lbl := edge.GetLabel()
	if len(lbl) != 0 {
		if _, err := fmt.Fprintf(w, "  %s -> %s [label=\"%s\"];\n", fromNode, toNode, lbl); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintf(w, "  %s -> %s;\n", fromNode, toNode); err != nil {
			return err
		}
	}
	return nil
}

func (d DOTRenderer) RenderNodesEdges(g Graph, out io.Writer) error {
	// Give nodes unique IDs for the graph
	nodeMap := map[Node]string{}
	x := 0
	for itr := g.GetNodes(); itr.Next(); {
		node := itr.Node()
		nodeId := fmt.Sprintf("n%d", x)
		rendered, err := d.RenderNode(nodeId, node, out)
		if err != nil {
			return err
		}
		if rendered {
			x++
			nodeMap[node] = nodeId
		}
	}
	for edgeItr := g.GetEdges(); edgeItr.Next(); {
		edge := edgeItr.Edge()
		fromNodeId, ok1 := nodeMap[edge.GetFrom()]
		toNodeId, ok2 := nodeMap[edge.GetTo()]
		if ok1 && ok2 {
			_, err := d.RenderEdge(fromNodeId, toNodeId, edge, out)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Render writes a DOT graph with the given name
func (d DOTRenderer) Render(g Graph, graphName string, out io.Writer) error {
	if _, err := fmt.Fprintf(out, "digraph %s {\n", graphName); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "rankdir=\"LR\";\n"); err != nil {
		return err
	}

	if err := d.RenderNodesEdges(g, out); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(out, "}\n"); err != nil {
		return err
	}
	return nil
}
