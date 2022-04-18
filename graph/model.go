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

// An Iterator iterates the items of a collection
type Iterator interface {
	// Next moves to the next item in the iterator, and returns true if
	// move was successful. If there are no next items remaining, returns false
	Next() bool

	// Value returns the current item in the iterator. This is undefined
	// before the first call to Next, and after Next returns false.
	Value() interface{}

	// MaxSize returns an estimation of maximum number of elements. If unknown, returns -1
	MaxSize() int
}

// NodeIterator iterates nodes of an underlying list
type NodeIterator interface {
	Iterator
	// Returns the current node
	Node() Node
}

// NodeSlice reads all the remaining items of a node iterator and returns them in a slice
func NodeSlice(in NodeIterator) []Node {
	ret := make([]Node, 0)
	for in.Next() {
		ret = append(ret, in.Node())
	}
	return ret
}

// EdgeIterator iterates the edges of an underlying list
type EdgeIterator interface {
	Iterator
	// Returns the current edge
	Edge() Edge
}

// EdgeSlice reads all the remaining items of an edge iterator and returns them in a slice
func EdgeSlice(in EdgeIterator) []Edge {
	ret := make([]Edge, 0)
	for in.Next() {
		ret = append(ret, in.Edge())
	}
	return ret
}

// TargetNodes returns the target nodes of all edges
func TargetNodes(in EdgeIterator) []Node {
	set := make(map[Node]struct{})
	for in.Next() {
		set[in.Edge().GetTo()] = struct{}{}
	}
	ret := make([]Node, 0, len(set))
	for x := range set {
		ret = append(ret, x)
	}
	return ret
}

// SourceNodes returns the source nodes of all edges
func SourceNodes(in EdgeIterator) []Node {
	set := make(map[Node]struct{})
	for in.Next() {
		set[in.Edge().GetFrom()] = struct{}{}
	}
	ret := make([]Node, 0, len(set))
	for x := range set {
		ret = append(ret, x)
	}
	return ret
}

// EdgeDir is used to show edge direction
type EdgeDir int

// Incoming and outgoing edge direction constants
const (
	IncomingEdge EdgeDir = -1
	OutgoingEdge EdgeDir = 1
)

// Node represents a graph node. A node is owned by a graph
type Node interface {
	GetGraph() Graph

	GetLabels() StringSet
	SetLabels(StringSet)

	GetProperty(string) (interface{}, bool)
	SetProperty(string, interface{})
	RemoveProperty(string)

	// Iterate properties until function returns false. Returns false if
	// function returns false, true if all properties are iterated (may be none)
	ForEachProperty(func(string, interface{}) bool) bool

	// Remove all connected edges, and remove the node
	DetachAndRemove()
	// Remove all connected edges
	Detach()

	// Returns an edge iterator for incoming or outgoing edges
	GetEdges(EdgeDir) EdgeIterator
	// Returns an edge iterator for incoming or outgoing edges with the given label
	GetEdgesWithLabel(EdgeDir, string) EdgeIterator
	// Returns an edge iterator for incoming or outgoingn edges that has the given labels
	GetEdgesWithAnyLabel(EdgeDir, StringSet) EdgeIterator
}

// NextNodesWith returns the nodes reachable from source with the given label at one step
func NextNodesWith(source Node, label string) []Node {
	return TargetNodes(source.GetEdgesWithLabel(OutgoingEdge, label))
}

// PrevNodesWith returns the nodes reachable from source with the given label at one step
func PrevNodesWith(source Node, label string) []Node {
	return SourceNodes(source.GetEdgesWithLabel(IncomingEdge, label))
}

type Edge interface {
	GetGraph() Graph

	GetLabel() string
	SetLabel(string)

	GetFrom() Node
	GetTo() Node

	GetProperty(string) (interface{}, bool)
	SetProperty(string, interface{})
	RemoveProperty(string)
	// Iterate properties until function returns false. Returns false if
	// function returns false, true if all properties are iterated (may be none)
	ForEachProperty(func(string, interface{}) bool) bool

	// Remove an edge
	Remove()
}

// A Graph is a collection of nodes and edges.
type Graph interface {
	NewNode(labels []string, properties map[string]interface{}) Node

	// NewEdge will add the nodes to the graph if they are not in the graph, and connect them
	NewEdge(from, to Node, label string, properties map[string]interface{}) Edge

	GetNodes() NodeIterator
	GetNodesWithAllLabels(StringSet) NodeIterator
	GetNodesWithProperty(string) NodeIterator
	NumNodes() int

	GetEdges() EdgeIterator
	GetEdgesWithAnyLabel(StringSet) EdgeIterator
	GetEdgesWithProperty(string) EdgeIterator
	NumEdges() int
}
