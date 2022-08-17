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

import (
	"github.com/emirpasic/gods/trees/btree"
)

// A setTree is a B-Tree of linkedhashsets
type setTree struct {
	tree *btree.Tree
}

func (s *setTree) add(key interface{}, id int, item interface{}) {
	if s.tree == nil {
		s.tree = btree.NewWith(16, ComparePropertyValue)
	}
	v, found := s.tree.Get(key)
	if !found {
		v = NewFastSet()
		s.tree.Put(key, v)
	}
	set := v.(*FastSet)
	set.Add(id, item)
}

func (s setTree) remove(key interface{}, id int, item interface{}) {
	if s.tree == nil {
		return
	}
	v, found := s.tree.Get(key)
	if !found {
		return
	}
	set := v.(*FastSet)
	set.Remove(id, item)
	if set.Size() == 0 {
		s.tree.Remove(key)
	}
}

// find returns the iterator and expected size.
func (s setTree) find(key interface{}) Iterator {
	if s.tree == nil {
		return emptyIterator{}
	}
	v, found := s.tree.Get(key)
	if !found {
		return emptyIterator{}
	}
	set := v.(*FastSet)
	itr := set.Iterator()
	return withSize(itr, set.Size())
}

func (s setTree) valueItr() Iterator {
	if s.tree == nil {
		return emptyIterator{}
	}
	treeItr := s.tree.Iterator()
	return &funcIterator{
		iteratorFunc: func() Iterator {
			if !treeItr.Next() {
				return nil
			}
			set := treeItr.Value().(*FastSet)
			itr := set.Iterator()
			return withSize(itr, set.Size())
		},
	}
}

type graphIndex struct {
	nodesByLabel NodeMap

	nodeProperties map[string]*setTree
	edgeProperties map[string]*setTree
}

func newGraphIndex() graphIndex {
	return graphIndex{
		nodeProperties: make(map[string]*setTree),
		edgeProperties: make(map[string]*setTree),
	}
}

// NodePropertyIndex sets up an index for the given node property
func (g *graphIndex) NodePropertyIndex(propertyName string, graph Graph) {
	_, exists := g.nodeProperties[propertyName]
	if exists {
		return
	}
	index := &setTree{}
	g.nodeProperties[propertyName] = index
	// Reindex
	for nodes := graph.GetNodes(); nodes.Next(); {
		node := nodes.Node().(*OCNode)
		value, ok := node.Properties[propertyName]
		if ok {
			index.add(value, node.id, node)
		}
	}
}

func (g *graphIndex) isNodePropertyIndexed(propertyName string) *setTree {
	return g.nodeProperties[propertyName]
}

func (g *graphIndex) isEdgePropertyIndexed(propertyName string) *setTree {
	return g.edgeProperties[propertyName]
}

// GetIteratorForNodeProperty returns an iterator for the given
// key/value, and the max size of the resultset. If no index found,
// returns nil,-1
func (g *graphIndex) GetIteratorForNodeProperty(key string, value interface{}) NodeIterator {
	index, found := g.nodeProperties[key]
	if !found {
		return nil
	}
	itr := index.find(value)
	return &nodeIterator{itr}
}

// NodesWithProperty returns an iterator that will go through the
// nodes that has the property
func (g *graphIndex) NodesWithProperty(key string) NodeIterator {
	index, found := g.nodeProperties[key]
	if !found {
		return nil
	}
	return &nodeIterator{index.valueItr()}
}

// EdgesWithProperty returns an iterator that will go through the
// edges that has the property
func (g *graphIndex) EdgesWithProperty(key string) EdgeIterator {
	index, found := g.edgeProperties[key]
	if !found {
		return nil
	}
	return &edgeIterator{index.valueItr()}
}

func (g *graphIndex) addNodeToIndex(node *OCNode) {
	g.nodesByLabel.Add(node)

	for k, v := range node.Properties {
		index, found := g.nodeProperties[k]
		if !found {
			continue
		}
		index.add(v, node.id, node)
	}
}

func (g *graphIndex) removeNodeFromIndex(node *OCNode) {
	g.nodesByLabel.Remove(node)

	for k, v := range node.Properties {
		index, found := g.nodeProperties[k]
		if !found {
			continue
		}
		index.remove(v, node.id, node)
	}
}

// EdgePropertyIndex sets up an index for the given edge property
func (g *graphIndex) EdgePropertyIndex(propertyName string, graph Graph) {
	_, exists := g.edgeProperties[propertyName]
	if exists {
		return
	}
	index := &setTree{}
	g.edgeProperties[propertyName] = index
	// Reindex
	for edges := graph.GetEdges(); edges.Next(); {
		edge := edges.Edge().(*OCEdge)
		value, ok := edge.Properties[propertyName]
		if ok {
			index.add(value, edge.id, edge)
		}
	}
}

func (g *graphIndex) addEdgeToIndex(edge *OCEdge) {
	for k, v := range edge.Properties {
		index, found := g.edgeProperties[k]
		if !found {
			continue
		}
		index.add(v, edge.id, edge)
	}
}

func (g *graphIndex) removeEdgeFromIndex(edge *OCEdge) {
	for k, v := range edge.Properties {
		index, found := g.edgeProperties[k]
		if !found {
			continue
		}
		index.remove(v, edge.id, edge)
	}
}

// GetIteratorForEdgeProperty returns an iterator for the given
// key/value, and the max size of the resultset. If no index found,
// returns nil,-1
func (g *graphIndex) GetIteratorForEdgeProperty(key string, value interface{}) EdgeIterator {
	index, found := g.edgeProperties[key]
	if !found {
		return nil
	}
	itr := index.find(value)
	return &edgeIterator{itr}
}
