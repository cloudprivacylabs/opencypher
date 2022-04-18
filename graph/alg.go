package graph

import ()

// Sources finds all the source nodes in the graph
func SourcesItr(graph Graph) NodeIterator {
	return &nodeIterator{
		&filterIterator{
			itr: graph.GetNodes(),
			filter: func(item interface{}) bool {
				node := item.(Node)
				if edges := node.GetEdges(IncomingEdge); edges.Next() {
					return false
				}
				return true
			},
		},
	}
}

// Sources finds all the source nodes in the graph
func Sources(graph Graph) []Node {
	return NodeSlice(SourcesItr(graph))
}

// CheckIsomoprhism checks to see if graphs given are equal as defined
// by the edge equivalence and node equivalence functions. The
// nodeEquivalenceFunction will be called for nodes whose labels are
// the same. The edgeEquivalenceFunction will be called for edges
// connecting equivalent nodes with the same labels.
//
// Node isomorphism check will fail if one node is equivalent to
// multiple nodes
func CheckIsomorphism(g1, g2 Graph, nodeEquivalenceFunc func(n1, n2 Node) bool, edgeEquivalenceFunc func(e1, e2 Edge) bool) bool {
	// Map of nodes1 -> nodes2
	nodeMapping1_2 := make(map[Node]Node)
	// Map of nodes2 -> nodes1
	nodeMapping2_1 := make(map[Node]Node)

	if g1.NumNodes() != g2.NumNodes() || g1.NumEdges() != g2.NumEdges() {
		return false
	}

	for nodes := g1.GetNodes(); nodes.Next(); {
		node1 := nodes.Node()
		for nodes2 := g2.GetNodes(); nodes2.Next(); {
			node2 := nodes2.Node()
			if node1.GetLabels().IsEqual(node2.GetLabels()) {
				if nodeEquivalenceFunc(node1, node2) {
					if _, ok := nodeMapping1_2[node1]; ok {
						return false
					}
					nodeMapping1_2[node1] = node2
					if _, ok := nodeMapping2_1[node2]; ok {
						return false
					}
					nodeMapping2_1[node2] = node1
				}
			}
		}
	}
	if len(nodeMapping1_2) != g1.NumNodes() {
		return false
	}
	// Node equivalences are established, now check edge equivalences
	for node1, node2 := range nodeMapping1_2 {
		// node1 and node2 are equivalent. Now we check if equivalent edges go to equivalent nodes
		edges1 := EdgeSlice(node1.GetEdges(OutgoingEdge))
		edges2 := EdgeSlice(node2.GetEdges(OutgoingEdge))
		// There must be same number of edges
		if len(edges1) != len(edges2) {
			return false
		}
		// Find equivalent edges
		edgeMap := make(map[Edge]Edge)
		for _, edge1 := range edges1 {
			found := false
			for _, edge2 := range edges2 {
				toNode1 := nodeMapping1_2[edge1.GetTo()]
				toNode2 := edge2.GetTo()
				if toNode1 == toNode2 {
					if edge1.GetLabel() == edge2.GetLabel() &&
						edgeEquivalenceFunc(edge1, edge2) {
						if found {
							// Multiple edges match
							return false
						}
						edgeMap[edge1] = edge2
						found = true
					}
				}
			}
			if !found {
				return false
			}
		}
		if len(edgeMap) != len(edges1) {
			return false
		}
	}
	return true
}

// ForEachNode iterates through all the nodes of g until predicate
// returns false or all nodes are processed.
func ForEachNode(g Graph, predicate func(Node) bool) bool {
	for nodes := g.GetNodes(); nodes.Next(); {
		if !predicate(nodes.Node()) {
			return false
		}
	}
	return true
}
