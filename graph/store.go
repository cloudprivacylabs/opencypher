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
	NumNodes() int
	GetNodes() NodeIterator
	GetEdges() EdgeIterator
	NumEdges() int
	GetEdgesWithAnyLabel(set StringSet) EdgeIterator
}

type OCStore struct {
	allNodes NodeSet
	allEdges EdgeMap
}

func NewOCStore() *OCStore {
	return &OCStore{}
}

func (store *OCStore) AddNode(node *OCNode) {
	store.allNodes.Add(node)
}

func (store *OCStore) DetachNode(node *OCNode) {
	for _, e := range EdgeSlice(node.incoming.Iterator()) {
		store.RemoveEdge(e.(*OCEdge))
	}
	node.incoming.Clear()
	for _, e := range EdgeSlice(node.outgoing.Iterator()) {
		store.RemoveEdge(e.(*OCEdge))
	}
	node.outgoing.Clear()
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
	edge.to.incoming.Add(edge)
	edge.from.outgoing.Add(edge)
}

func (store *OCStore) disconnect(edge *OCEdge) {
	edge.to.incoming.Remove(edge)
	edge.from.outgoing.Remove(edge)
}
