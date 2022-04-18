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
	"github.com/emirpasic/gods/maps/linkedhashmap"
	"github.com/emirpasic/gods/sets/linkedhashset"
)

// An EdgeMap stores edges indexed by edge label
type EdgeMap struct {
	// m[string]*linkedhashset
	m *linkedhashmap.Map
	n int
}

func (em *EdgeMap) Add(edge *OCEdge) {
	if em.m == nil {
		em.m = linkedhashmap.New()
	}

	var set *linkedhashset.Set
	v, found := em.m.Get(edge.label)
	if !found {
		set = linkedhashset.New()
		em.m.Put(edge.label, set)
	} else {
		set = v.(*linkedhashset.Set)
	}
	k := set.Size()
	set.Add(edge)
	if set.Size() > k {
		em.n++
	}
}

func (em EdgeMap) Remove(edge *OCEdge) {
	if em.m == nil {
		return
	}
	var set *linkedhashset.Set
	v, found := em.m.Get(edge.label)
	if !found {
		return
	}
	set = v.(*linkedhashset.Set)
	k := set.Size()
	set.Remove(edge)
	if set.Size() < k {
		em.n--
	}
	if set.Size() == 0 {
		em.m.Remove(edge.label)
	}
}

func (em EdgeMap) IsEmpty() bool {
	if em.m == nil {
		return true
	}
	if em.m.Size() == 0 {
		return true
	}
	return false
}

func (em EdgeMap) Len() int { return em.n }

type edgeMapIterator struct {
	labels  iteratorWithoutSize
	current EdgeIterator
	size    int
}

func (itr *edgeMapIterator) Next() bool {
	if itr.current != nil {
		if itr.current.Next() {
			return true
		}
		itr.current = nil
	}
	if itr.labels == nil {
		return false
	}
	if !itr.labels.Next() {
		return false
	}
	set := itr.labels.Value().(*linkedhashset.Set)
	setItr := set.Iterator()
	itr.current = &edgeIterator{withSize(&setItr, -1)}
	itr.current.Next()
	return true
}

func (itr *edgeMapIterator) Value() interface{} {
	return itr.current.Value()
}

func (itr *edgeMapIterator) Edge() Edge {
	return itr.current.Edge()
}

func (itr *edgeMapIterator) MaxSize() int { return itr.size }

func (em EdgeMap) Iterator() EdgeIterator {
	if em.m == nil {
		return &edgeIterator{&emptyIterator{}}
	}
	i := em.m.Iterator()
	return &edgeMapIterator{labels: &i, size: em.Len()}
}

func (em EdgeMap) IteratorLabel(label string) EdgeIterator {
	if em.m == nil {
		return &edgeIterator{&emptyIterator{}}
	}
	v, found := em.m.Get(label)
	if !found {
		return &edgeIterator{&emptyIterator{}}
	}
	set := v.(*linkedhashset.Set)
	i := set.Iterator()
	return &edgeIterator{withSize(&i, set.Size())}
}

func (em EdgeMap) IteratorAnyLabel(labels StringSet) EdgeIterator {
	if em.m == nil {
		return &edgeIterator{&emptyIterator{}}
	}
	strings := labels.Slice()
	return &edgeIterator{&funcIterator{
		iteratorFunc: func() Iterator {
			for len(strings) != 0 {
				v, found := em.m.Get(strings[0])
				strings = strings[1:]
				if !found {
					continue
				}
				itr := v.(*linkedhashset.Set).Iterator()
				return withSize(&itr, -1)
			}
			return nil
		},
	},
	}
}

// An NodeMap stores nodes indexed by node labels
type NodeMap struct {
	// m[string]*FastSet
	m        *linkedhashmap.Map
	nolabels FastSet
}

func (nm *NodeMap) Add(node *OCNode) {
	if nm.m == nil {
		nm.m = linkedhashmap.New()
	}
	if len(node.labels) == 0 {
		nm.nolabels.Add(node)
		return
	}

	var set *FastSet
	for label := range node.labels {
		v, found := nm.m.Get(label)
		if !found {
			set = &FastSet{}
			nm.m.Put(label, set)
		} else {
			set = v.(*FastSet)
		}
		set.Add(node)
	}
}

func (nm NodeMap) Remove(node *OCNode) {
	if nm.m == nil {
		return
	}
	if len(node.labels) == 0 {
		nm.nolabels.Remove(node)
		return
	}
	var set *FastSet
	for label := range node.labels {
		v, found := nm.m.Get(label)
		if !found {
			continue
		}
		set = v.(*FastSet)
		set.Remove(node)
		if set.Len() == 0 {
			nm.m.Remove(label)
		}
	}
}

func (nm NodeMap) IsEmpty() bool {
	if nm.m == nil {
		return true
	}
	if nm.m.Size() == 0 {
		return true
	}
	return false
}

type nodeMapIterator struct {
	labels     *linkedhashmap.Iterator
	seenLabels []string
	current    NodeIterator
}

func (itr *nodeMapIterator) Next() bool {
	if itr.current != nil {
		if itr.current.Next() {
			return true
		}
		itr.current = nil
	}
	if itr.labels == nil {
		return false
	}
	if !itr.labels.Next() {
		return false
	}
	itr.seenLabels = append(itr.seenLabels, itr.labels.Key().(string))
	set := itr.labels.Value().(*FastSet)
	setItr := set.Iterator()
	itr.current = &nodeIterator{withSize(setItr, -1)}
	itr.current.Next()
	return true
}

func (itr *nodeMapIterator) Value() interface{} {
	return itr.current.Value()
}

func (itr *nodeMapIterator) Node() Node {
	return itr.current.Node()
}

func (nm NodeMap) Iterator() NodeIterator {
	if nm.m == nil {
		return &nodeIterator{&emptyIterator{}}
	}
	i := nm.m.Iterator()

	nmIterator := &nodeMapIterator{labels: &i}
	return &nodeIterator{
		MultiIterator(
			&filterIterator{
				itr: withSize(nmIterator, -1),
				filter: func(node interface{}) bool {
					onode := node.(*OCNode)
					nSeen := 0
					for _, l := range nmIterator.seenLabels {
						if onode.labels.Has(l) {
							nSeen++
						}
					}
					return nSeen < 2
				},
			},
			nm.nolabels.Iterator(),
		),
	}
}

func (nm NodeMap) IteratorAllLabels(labels StringSet) NodeIterator {
	if nm.m == nil {
		return &nodeIterator{&emptyIterator{}}
	}
	// Find the smallest map element, iterate that
	var minSet *FastSet
	for label := range labels {
		v, found := nm.m.Get(label)
		if !found {
			return &nodeIterator{&emptyIterator{}}
		}
		mp := v.(*FastSet)
		if minSet == nil || minSet.Len() > mp.Len() {
			minSet = mp
		}
	}
	itr := minSet.Iterator()
	flt := &filterIterator{
		itr: withSize(itr, minSet.Len()),
		filter: func(item interface{}) bool {
			onode := item.(*OCNode)
			return onode.labels.HasAllSet(labels)
		},
	}
	return &nodeIterator{flt}
}
