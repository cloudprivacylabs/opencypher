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

type OCGraph struct {
	index    graphIndex
	allNodes NodeSet
	allEdges EdgeMap
	idBase   int
}

func NewOCGraph() *OCGraph {
	return &OCGraph{
		index: graphIndex{},
	}
}

// NewNode creates a new node with the given labels and properties
func (g *OCGraph) NewNode(labels []string, properties map[string]interface{}) Node {
	node := g.NewOCNode(labels, properties)
	return node
}

// NewOCNode creates a new node with the given labels and properties
func (g *OCGraph) NewOCNode(labels []string, properties map[string]interface{}) *OCNode {
	node := &OCNode{labels: NewStringSet(labels...), Properties: Properties(properties), graph: g}
	node.id = g.idBase
	g.idBase++
	g.addNode(node)
	return node
}

func (g *OCGraph) cloneNode(node *OCNode, cloneProperty func(string, interface{}) interface{}) *OCNode {
	newNode := &OCNode{
		labels:     node.labels.Clone(),
		Properties: node.Properties.clone(cloneProperty),
		graph:      g,
	}
	newNode.id = g.idBase
	g.idBase++
	g.addNode(newNode)
	return newNode
}

func (g *OCGraph) addNode(node *OCNode) {
	g.allNodes.Add(node)
	g.index.addNodeToIndex(node)
}

func (g *OCGraph) SetNodeLabels(node *OCNode, labels StringSet) {
	g.index.nodesByLabel.Replace(node, node.GetLabels(), labels)
	node.labels = labels.Clone()
}

func (g *OCGraph) SetNodeProperty(node *OCNode, key string, value interface{}) {
	if node.Properties == nil {
		node.Properties = make(Properties)
	}
	oldValue, exists := node.Properties[key]
	nix := g.index.isNodePropertyIndexed(key)
	if nix != nil && exists {
		nix.remove(oldValue, node.id, node)
	}
	node.Properties[key] = value
	if nix != nil {
		nix.add(value, node.id, node)
	}
}

func (g *OCGraph) RemoveNodeProperty(node *OCNode, key string) {
	if node.Properties == nil {
		return
	}
	value, exists := node.Properties[key]
	if !exists {
		return
	}
	nix := g.index.isNodePropertyIndexed(key)
	if nix != nil {
		nix.remove(value, node.id, node)
	}
	delete(node.Properties, key)
}

func (g *OCGraph) DetachRemoveNode(node *OCNode) {
	g.DetachNode(node)
	g.allNodes.Remove(node)
	g.index.removeNodeFromIndex(node)
}

func (g *OCGraph) DetachNode(node *OCNode) {
	for _, e := range EdgeSlice(node.incoming.Iterator()) {
		edge := e.(*OCEdge)
		g.disconnect(edge)
		g.allEdges.Remove(edge)
		g.index.removeEdgeFromIndex(edge)
	}
	node.incoming.Clear()
	for _, e := range EdgeSlice(node.outgoing.Iterator()) {
		edge := e.(*OCEdge)
		g.disconnect(edge)
		g.allEdges.Remove(edge)
		g.index.removeEdgeFromIndex(edge)
	}
	node.outgoing.Clear()
}

func (g *OCGraph) NewEdge(from, to Node, label string, properties map[string]interface{}) Edge {
	ofrom := from.(*OCNode)
	oto := to.(*OCNode)
	if ofrom.graph != g {
		panic("from node is not in graph")
	}
	if oto.graph != g {
		panic("to node is not in graph")
	}
	newEdge := &OCEdge{
		from:       ofrom,
		to:         oto,
		label:      label,
		Properties: Properties(properties),
		id:         g.idBase,
	}
	g.idBase++
	g.allEdges.Add(newEdge)
	g.connect(newEdge)
	g.index.addEdgeToIndex(newEdge)
	return newEdge
}

func (g *OCGraph) cloneEdge(from, to Node, edge *OCEdge, cloneProperty func(string, interface{}) interface{}) *OCEdge {
	ofrom := from.(*OCNode)
	oto := to.(*OCNode)
	if ofrom.graph != g {
		panic("from node is not in graph")
	}
	if oto.graph != g {
		panic("to node is not in graph")
	}
	newEdge := &OCEdge{
		from:       ofrom,
		to:         oto,
		label:      edge.label,
		Properties: edge.Properties.clone(cloneProperty),
		id:         g.idBase,
	}
	g.idBase++
	g.allEdges.Add(newEdge)
	g.connect(newEdge)
	g.index.addEdgeToIndex(newEdge)
	return newEdge
}

func (g *OCGraph) connect(edge *OCEdge) {
	edge.to.incoming.Add(edge)
	edge.from.outgoing.Add(edge)
}

func (g *OCGraph) disconnect(edge *OCEdge) {
	edge.to.incoming.Remove(edge)
	edge.from.outgoing.Remove(edge)
}

func (g *OCGraph) SetEdgeLabel(edge *OCEdge, label string) {
	g.disconnect(edge)
	edge.label = label
	g.connect(edge)
}

func (g *OCGraph) RemoveEdge(edge *OCEdge) {
	g.disconnect(edge)
	g.allEdges.Remove(edge)
}

func (g *OCGraph) SetEdgeProperty(edge *OCEdge, key string, value interface{}) {
	if edge.Properties == nil {
		edge.Properties = make(Properties)
	}
	oldValue, exists := edge.Properties[key]
	nix := g.index.isEdgePropertyIndexed(key)
	if nix != nil && exists {
		nix.remove(oldValue, edge.id, edge)
	}
	edge.Properties[key] = value
	if nix != nil {
		nix.add(value, edge.id, edge)
	}
}

func (g *OCGraph) RemoveEdgeProperty(edge *OCEdge, key string) {
	if edge.Properties == nil {
		return
	}
	oldValue, exists := edge.Properties[key]
	if !exists {
		return
	}
	nix := g.index.isEdgePropertyIndexed(key)
	if nix != nil {
		nix.remove(oldValue, edge.id, edge)
	}
	delete(edge.Properties, key)
}

func (g *OCGraph) NumNodes() int {
	return g.allNodes.Len()
}

func (g *OCGraph) NumEdges() int {
	return g.allEdges.Len()
}

func (g *OCGraph) GetNodes() NodeIterator {
	return g.allNodes.Iterator()
}

func (g *OCGraph) GetNodesWithAllLabels(labels StringSet) NodeIterator {
	return g.index.nodesByLabel.IteratorAllLabels(labels)
}

func (g *OCGraph) GetEdges() EdgeIterator {
	return g.allEdges.Iterator()
}

func (g *OCGraph) GetEdgesWithAnyLabel(set StringSet) EdgeIterator {
	return g.allEdges.IteratorAnyLabel(set)
}

// FindNodes returns an iterator that will iterate through all the
// nodes that have all of the given labels and properties. If
// allLabels is nil or empty, it does not look at the labels. If
// properties is nil or empty, it does not look at the properties
func (g *OCGraph) FindNodes(allLabels StringSet, properties map[string]interface{}) NodeIterator {
	if len(allLabels) == 0 && len(properties) == 0 {
		// Return all nodes
		return g.GetNodes()
	}

	var nodesByLabelItr NodeIterator
	if len(allLabels) > 0 {
		nodesByLabelItr = g.index.nodesByLabel.IteratorAllLabels(allLabels)
	}
	// Select the iterator with minimum max size
	nodesByLabelSize := nodesByLabelItr.MaxSize()
	propertyIterators := make(map[string]NodeIterator)
	if len(properties) > 0 {
		for k, v := range properties {
			itr := g.index.GetIteratorForNodeProperty(k, v)
			if itr == nil {
				continue
			}
			propertyIterators[k] = itr
		}
	}
	var minimumPropertyItrKey string
	minPropertySize := -1
	for k, itr := range propertyIterators {
		maxSize := itr.MaxSize()
		if maxSize == -1 {
			continue
		}
		if minPropertySize == -1 || minPropertySize > maxSize {
			minPropertySize = maxSize
			minimumPropertyItrKey = k
		}
	}

	nodeFilterFunc := GetNodeFilterFunc(allLabels, properties)
	// Iterate the minimum iterator, with a filter
	if nodesByLabelSize != -1 && (minPropertySize == -1 || minPropertySize > nodesByLabelSize) {
		// Iterate by node label
		// build a filter from properties
		return &nodeIterator{
			&filterIterator{
				itr: nodesByLabelItr,
				filter: func(item interface{}) bool {
					return nodeFilterFunc(item.(Node))
				},
			},
		}
	}
	if minPropertySize != -1 {
		// Iterate by property
		return &nodeIterator{
			&filterIterator{
				itr: propertyIterators[minimumPropertyItrKey],
				filter: func(item interface{}) bool {
					return nodeFilterFunc(item.(Node))
				},
			},
		}
	}
	// Iterate all
	return g.GetNodes()
}

// AddEdgePropertyIndex adds an index for the given edge property
func (g *OCGraph) AddEdgePropertyIndex(propertyName string) {
	g.index.EdgePropertyIndex(propertyName, g)
}

// AddNodePropertyIndex adds an index for the given node property
func (g *OCGraph) AddNodePropertyIndex(propertyName string) {
	g.index.NodePropertyIndex(propertyName, g)
}

// GetNodesWithProperty returns an iterator for the nodes that has the property
func (g *OCGraph) GetNodesWithProperty(property string) NodeIterator {
	itr := g.index.NodesWithProperty(property)
	if itr != nil {
		return itr
	}
	return &nodeIterator{&filterIterator{
		itr: g.GetNodes(),
		filter: func(v interface{}) bool {
			wp, ok := v.(Node)
			if !ok {
				return false
			}
			_, exists := wp.GetProperty(property)
			return exists
		},
	}}
}

// GetEdgesWithProperty returns an iterator for the edges that has the property
func (g *OCGraph) GetEdgesWithProperty(property string) EdgeIterator {
	itr := g.index.EdgesWithProperty(property)
	if itr != nil {
		return itr
	}
	return &edgeIterator{&filterIterator{
		itr: g.GetEdges(),
		filter: func(v interface{}) bool {
			wp, ok := v.(Edge)
			if !ok {
				return false
			}
			_, exists := wp.GetProperty(property)
			return exists
		},
	}}

}

// FindEdges returns an iterator that will iterate through all the
// edges whose label is in the given labels and have all the
// properties. If labels is nil or empty, it does not look at the
// labels. If properties is nil or empty, it does not look at the
// properties
func (g *OCGraph) FindEdges(labels StringSet, properties map[string]interface{}) EdgeIterator {
	if len(labels) == 0 && len(properties) == 0 {
		// Return all edges
		return g.GetEdges()
	}

	var edgesByLabelItr EdgeIterator
	if len(labels) > 0 {
		edgesByLabelItr = g.GetEdgesWithAnyLabel(labels)
	}
	// Select the iterator with minimum max size
	edgesByLabelSize := edgesByLabelItr.MaxSize()
	propertyIterators := make(map[string]EdgeIterator)
	if len(properties) > 0 {
		for k, v := range properties {
			itr := g.index.GetIteratorForEdgeProperty(k, v)
			if itr == nil {
				continue
			}
			propertyIterators[k] = itr
		}
	}
	var minimumPropertyItrKey string
	minPropertySize := -1
	for k, itr := range propertyIterators {
		maxSize := itr.MaxSize()
		if maxSize == -1 {
			continue
		}
		if minPropertySize == -1 || minPropertySize > maxSize {
			minPropertySize = maxSize
			minimumPropertyItrKey = k
		}
	}

	edgeFilterFunc := GetEdgeFilterFunc(labels, properties)
	// Iterate the minimum iterator, with a filter
	if edgesByLabelSize != -1 && (minPropertySize == -1 || minPropertySize > edgesByLabelSize) {
		// Iterate by edge label
		// build a filter from properties
		return &edgeIterator{
			&filterIterator{
				itr: edgesByLabelItr,
				filter: func(item interface{}) bool {
					return edgeFilterFunc(item.(Edge))
				},
			},
		}
	}
	if minPropertySize != -1 {
		// Iterate by property
		return &edgeIterator{
			&filterIterator{
				itr: propertyIterators[minimumPropertyItrKey],
				filter: func(item interface{}) bool {
					return edgeFilterFunc(item.(Edge))
				},
			},
		}
	}
	// Iterate all
	return g.GetEdges()
}

// GetNodeFilterFunc returns a function that can be used to pass
// nodes that have all the specified labels, with correct property
// values
func GetNodeFilterFunc(labels StringSet, properties map[string]interface{}) func(Node) bool {
	return func(node Node) (cmp bool) {
		onode := node.(*OCNode)
		if len(labels) > 0 {
			if !onode.labels.HasAllSet(labels) {
				return false
			}
		}
		defer func() {
			if r := recover(); r != nil {
				cmp = false
			}
		}()
		for k, v := range properties {
			nodeValue, exists := onode.GetProperty(k)
			if !exists {
				if v != nil {
					return false
				}
			}
			if ComparePropertyValue(v, nodeValue) != 0 {
				return false
			}
		}
		return true
	}
}

// GetEdgeFilterFunc returns a function that can be used to pass edges
// that have at least one of the specified labels, with correct
// property values
func GetEdgeFilterFunc(labels StringSet, properties map[string]interface{}) func(Edge) bool {
	return func(edge Edge) (cmp bool) {
		oedge := edge.(*OCEdge)
		if len(labels) > 0 {
			if !labels.Has(oedge.label) {
				return false
			}
		}
		defer func() {
			if r := recover(); r != nil {
				cmp = false
			}
		}()
		for k, v := range properties {
			edgeValue, exists := oedge.GetProperty(k)
			if !exists {
				if v != nil {
					return false
				}
			}
			if ComparePropertyValue(v, edgeValue) != 0 {
				return false
			}
		}
		return true
	}
}

type WithProperties interface {
	GetProperty(key string) (interface{}, bool)
}

func buildPropertyFilterFunc(key string, value interface{}) func(WithProperties) bool {
	return func(properties WithProperties) (cmp bool) {
		pvalue, exists := properties.GetProperty(key)
		if !exists {
			return value == nil
		}

		defer func() {
			if r := recover(); r != nil {
				cmp = false
			}
		}()
		return ComparePropertyValue(value, pvalue) == 0
	}
}
