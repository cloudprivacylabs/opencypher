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
	index  *graphIndex
	store  OCStore
	idBase int
}

func NewOCGraph() *OCGraph {
	return &OCGraph{
		store: *NewOCStore(),
		index: &graphIndex{},
	}
}

// NewNode creates a new node with the given labels and properties
func (g *OCGraph) NewNode(labels []string, properties map[string]interface{}) Node {
	return g.NewOCNode(labels, properties)
}

// NewOCNode creates a new node with the given labels and properties
func (g *OCGraph) NewOCNode(labels []string, properties map[string]interface{}) *OCNode {
	node := &OCNode{id: g.idBase, labels: NewStringSet(labels...), Properties: Properties(properties), graph: g}
	g.idBase++
	g.addNode(node)
	return node
}

func (g *OCGraph) addNode(node *OCNode) {
	g.store.AddNode(node)
	if g.index != nil {
		g.index.addNodeToIndex(node)
	}
}

func (g *OCGraph) SetNodeLabels(node *OCNode, labels StringSet) {
	if g.index != nil {
		g.index.removeNodeFromIndex(node)
	}
	node.labels = labels.Clone()
	if g.index != nil {
		g.index.addNodeToIndex(node)
	}
}

func (g *OCGraph) SetNodeProperty(node *OCNode, key string, value interface{}) {
	if node.Properties == nil {
		node.Properties = make(Properties)
	}
	indexed := g.index.IsNodePropertyIndexed(key)
	if indexed {
		g.index.removeNodeFromIndex(node)
	}
	node.Properties[key] = value
	if indexed {
		g.index.addNodeToIndex(node)
	}
}

func (g *OCGraph) RemoveNodeProperty(node *OCNode, key string) {
	if node.Properties == nil {
		return
	}
	if g.index.IsNodePropertyIndexed(key) {
		g.index.removeNodeFromIndex(node)
	}
	delete(node.Properties, key)
}

func (g *OCGraph) DetachRemoveNode(node *OCNode) {
	g.store.DetachRemoveNode(node)
	if g.index != nil {
		g.index.removeNodeFromIndex(node)
	}
}

func (g *OCGraph) DetachNode(node *OCNode) {
	g.store.DetachNode(node)
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
		id:         g.idBase,
		from:       ofrom,
		to:         oto,
		label:      label,
		Properties: Properties(properties),
		graph:      g,
	}
	g.idBase++
	g.store.AddEdge(newEdge)
	return newEdge
}

func (g *OCGraph) connect(edge *OCEdge) {
	g.store.connect(edge)
}

func (g *OCGraph) disconnect(edge *OCEdge) {
	g.store.disconnect(edge)
}

func (g *OCGraph) SetEdgeLabel(edge *OCEdge, label string) {
	g.disconnect(edge)
	edge.label = label
	g.connect(edge)
}

func (g *OCGraph) RemoveEdge(edge *OCEdge) {
	g.store.RemoveEdge(edge)
}

func (g *OCGraph) SetEdgeProperty(edge *OCEdge, key string, value interface{}) {
	if edge.Properties == nil {
		edge.Properties = make(Properties)
	}
	indexed := g.index.IsEdgePropertyIndexed(key)
	if indexed {
		g.index.removeEdgeFromIndex(edge)
	}
	edge.Properties[key] = value
	if indexed {
		g.index.addEdgeToIndex(edge)
	}
}

func (g *OCGraph) RemoveEdgeProperty(edge *OCEdge, key string) {
	if edge.Properties == nil {
		return
	}
	if g.index.IsEdgePropertyIndexed(key) {
		g.index.removeEdgeFromIndex(edge)
	}
	delete(edge.Properties, key)
}

func (g *OCGraph) NumNodes() int {
	return g.store.NumNodes()
}

func (g *OCGraph) NumEdges() int {
	return g.store.NumEdges()
}

func (g *OCGraph) GetNodes() NodeIterator {
	return g.store.GetNodes()
}

func (g *OCGraph) GetNodesWithAllLabels(labels StringSet) NodeIterator {
	return g.index.nodesByLabel.IteratorAllLabels(labels)
}

func (g *OCGraph) GetEdges() EdgeIterator {
	return g.store.GetEdges()
}

func (g *OCGraph) GetEdgesWithAnyLabel(set StringSet) EdgeIterator {
	return g.store.GetEdgesWithAnyLabel(set)
}

func (g *OCGraph) GetNodeEdges(node Node, dir EdgeDir) EdgeIterator {
	onode := node.(*OCNode)
	return g.store.GetNodeEdges(onode, dir)
}

func (g *OCGraph) GetNodeEdgesWithLabel(node Node, dir EdgeDir, label string) EdgeIterator {
	onode := node.(*OCNode)
	return g.store.GetNodeEdgesWithLabel(onode, dir, label)
}

func (g *OCGraph) GetNodeEdgesWithAnyLabel(node Node, dir EdgeDir, set StringSet) EdgeIterator {
	onode := node.(*OCNode)
	return g.store.GetNodeEdgesWithAnyLabel(onode, dir, set)
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
	return func(node Node) bool {
		onode := node.(*OCNode)
		if len(labels) > 0 {
			if !onode.labels.HasAllSet(labels) {
				return false
			}
		}
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
	return func(edge Edge) bool {
		oedge := edge.(*OCEdge)
		if len(labels) > 0 {
			if !labels.Has(oedge.label) {
				return false
			}
		}
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
	return func(properties WithProperties) bool {
		pvalue, exists := properties.GetProperty(key)
		if !exists {
			return value == nil
		}
		return ComparePropertyValue(value, pvalue) == 0
	}
}
