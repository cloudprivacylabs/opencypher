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

// Store keeps the nodes, edges, and the connections of a graph
type Store interface {
	AddNode(node *OCNode)
	DetachNode(node *OCNode)
	DetachRemoveNode(node *OCNode)
	AddEdge(edge *OCEdge)
	RemoveEdge(edge *OCEdge)
	GetNodeEdges(node *OCNode, dir EdgeDir) EdgeIterator
	GetNodeEdgesWithLabel(node *OCNode, dir EdgeDir, label string) EdgeIterator
	GetNodeEdgesWithAnyLabel(node *OCNode, dir EdgeDir, set StringSet) EdgeIterator
	NumNodes() int
	GetNodes() NodeIterator
	GetEdges() EdgeIterator
	NumEdges() int
	GetEdgesWithAnyLabel(set StringSet) EdgeIterator
}

type OCStore struct {
	allNodes NodeSet
	allEdges EdgeMap

	incoming map[*OCNode]EdgeMap
	outgoing map[*OCNode]EdgeMap
}

func NewOCStore() *OCStore {
	return &OCStore{
		incoming: make(map[*OCNode]EdgeMap),
		outgoing: make(map[*OCNode]EdgeMap),
	}
}

func (store *OCStore) AddNode(node *OCNode) {
	store.allNodes.Add(node)
}

func (store *OCStore) DetachNode(node *OCNode) {
	for _, e := range EdgeSlice(store.incoming[node].Iterator()) {
		store.RemoveEdge(e.(*OCEdge))
	}
	delete(store.incoming, node)
	for _, e := range EdgeSlice(store.outgoing[node].Iterator()) {
		store.RemoveEdge(e.(*OCEdge))
	}
	delete(store.outgoing, node)
}

func (store *OCStore) DetachRemoveNode(node *OCNode) {
	store.DetachNode(node)
	store.allNodes.Remove(node)
}

func (store *OCStore) AddEdge(edge *OCEdge) {
	store.AddNode(edge.from)
	store.AddNode(edge.to)
	store.allEdges.Add(edge)
	store.connect(edge)
}

func (store *OCStore) RemoveEdge(edge *OCEdge) {
	store.disconnect(edge)
	store.allEdges.Remove(edge)
}

func (store *OCStore) GetNodeEdges(node *OCNode, dir EdgeDir) EdgeIterator {
	if dir == IncomingEdge {
		return store.incoming[node].Iterator()
	}
	return store.outgoing[node].Iterator()
}

func (store *OCStore) GetNodeEdgesWithLabel(node *OCNode, dir EdgeDir, label string) EdgeIterator {
	if dir == IncomingEdge {
		return store.incoming[node].IteratorLabel(label)
	}
	return store.outgoing[node].IteratorLabel(label)
}

func (store *OCStore) GetNodeEdgesWithAnyLabel(node *OCNode, dir EdgeDir, set StringSet) EdgeIterator {
	if dir == IncomingEdge {
		if len(set) == 0 {
			return store.incoming[node].Iterator()
		}
		return store.incoming[node].IteratorAnyLabel(set)
	}
	if len(set) == 0 {
		return store.outgoing[node].Iterator()
	}
	return store.outgoing[node].IteratorAnyLabel(set)
}

func (store *OCStore) NumNodes() int {
	return store.allNodes.Len()
}

func (store *OCStore) GetNodes() NodeIterator {
	return store.allNodes.Iterator()
}

func (store *OCStore) GetEdges() EdgeIterator {
	return store.allEdges.Iterator()
}

func (store *OCStore) NumEdges() int {
	return store.allEdges.Len()
}

func (store *OCStore) GetEdgesWithAnyLabel(set StringSet) EdgeIterator {
	return store.allEdges.IteratorAnyLabel(set)
}

func (store *OCStore) connect(edge *OCEdge) {
	in := store.incoming[edge.to]
	in.Add(edge)
	store.incoming[edge.to] = in
	out := store.outgoing[edge.from]
	out.Add(edge)
	store.outgoing[edge.from] = out
}

func (store *OCStore) disconnect(edge *OCEdge) {
	in := store.incoming[edge.to]
	in.Remove(edge)
	if in.IsEmpty() {
		delete(store.incoming, edge.to)
	}
	out := store.outgoing[edge.from]
	out.Remove(edge)
	if out.IsEmpty() {
		delete(store.outgoing, edge.from)
	}
}
