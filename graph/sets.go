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
	"container/list"
)

// A FastSet is a set of objects with constant-time
// insertion/deletion, with iterator support
type FastSet struct {
	m map[int]*list.Element
	l *list.List
}

func NewFastSet() *FastSet {
	return &FastSet{}
}

func (f FastSet) Len() int  { return len(f.m) }
func (f FastSet) Size() int { return len(f.m) }

func (f *FastSet) Add(id int, item interface{}) {
	if f.m == nil {
		f.m = make(map[int]*list.Element)
		f.l = list.New()
	}
	_, exists := f.m[id]
	if exists {
		return
	}
	el := f.l.PushBack(item)
	f.m[id] = el
}

func (f *FastSet) Remove(id int, item interface{}) {
	if f.m == nil {
		return
	}
	el := f.m[id]
	if el == nil {
		return
	}
	delete(f.m, id)
	f.l.Remove(el)
}

func (f FastSet) Has(id int) bool {
	if f.m == nil {
		return false
	}
	_, exists := f.m[id]
	return exists
}

func (f FastSet) Iterator() Iterator {
	if f.m == nil {
		return emptyIterator{}
	}
	return &listIterator{next: f.l.Front(), size: f.Len()}
}

type NodeSet struct {
	set FastSet
}

func (set *NodeSet) Add(node *OCNode) {
	set.set.Add(node.id, node)
}

func (set NodeSet) Remove(node *OCNode) {
	set.set.Remove(node.id, node)
}

func (set NodeSet) Has(node *OCNode) bool {
	return set.set.Has(node.id)
}

func (set NodeSet) Len() int {
	return set.set.Len()
}

func (set NodeSet) Iterator() NodeIterator {
	i := set.set.Iterator()
	return &nodeIterator{withSize(i, set.set.Len())}
}

func (set NodeSet) Slice() []Node {
	return NodeSlice(set.Iterator())
}

// EdgeSet keeps an unordered set of edges
type EdgeSet struct {
	set FastSet
}

func (set *EdgeSet) Add(edge *OCEdge) {
	set.set.Add(edge.id, edge)
}

func (set EdgeSet) Remove(edge *OCEdge) {
	set.set.Remove(edge.id, edge)
}

func (set EdgeSet) Len() int {
	return set.set.Len()
}

func (set EdgeSet) Iterator() EdgeIterator {
	i := set.set.Iterator()
	return &edgeIterator{withSize(i, set.set.Len())}
}

func (set EdgeSet) Slice() []Edge {
	return EdgeSlice(set.Iterator())
}
