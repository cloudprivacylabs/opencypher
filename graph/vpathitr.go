// Copyright 2021 Cloud Privacy Labs, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package graph

// CollectAllPaths iterates the variable length paths that have the
// edges in firstLeg. For each edge, it calls the edgeFilter
// function. If the edge is accepted, it recursively descends and
// calls accumulator.AddPath for each discovered path until AddPath
// returns false
func CollectAllPaths(graph Graph, fromNode Node, firstLeg EdgeIterator, edgeFilter func(Edge) bool, dir EdgeDir, min, max int, accumulator func([]Edge, Node) bool) {
	var recurse func([]Edge, Node) bool

	isLoop := func(node Node, edges []Edge) bool {
		for _, e := range edges {
			if e.GetFrom() == node {
				return true
			}
		}
		if len(edges) > 0 {
			return edges[len(edges)-1].GetTo() == node
		}
		return false
	}

	recurse = func(prefix []Edge, lastNode Node) bool {

		var endNode Node
		switch dir {
		case OutgoingEdge:
			endNode = prefix[len(prefix)-1].GetTo()
		case IncomingEdge:
			endNode = prefix[len(prefix)-1].GetFrom()
		case AnyEdge:
			if prefix[len(prefix)-1].GetTo() == lastNode {
				endNode = prefix[len(prefix)-1].GetFrom()
			} else {
				endNode = prefix[len(prefix)-1].GetTo()
			}
		}

		if (min == -1 || len(prefix) >= min) && (max == -1 || len(prefix) <= max) {
			// A valid path
			entry := make([]Edge, len(prefix))
			copy(entry, prefix)
			if !accumulator(entry, endNode) {
				return false
			}
		}

		if max != -1 && len(prefix) >= max {
			return true
		}

		if isLoop(endNode, prefix[:len(prefix)-1]) {
			return true
		}
		itr := &edgeIterator{
			&filterIterator{
				itr: endNode.GetEdges(dir),
				filter: func(item interface{}) bool {
					return edgeFilter(item.(*OCEdge))
				},
			},
		}
		for itr.Next() {
			edge := itr.Edge()
			if !recurse(append(prefix, edge.(*OCEdge)), endNode) {
				return false
			}
		}
		return true
	}

	for firstLeg.Next() {
		edge := firstLeg.Edge()
		if !recurse([]Edge{edge}, fromNode) {
			break
		}
	}
}
