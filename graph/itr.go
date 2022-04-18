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

type emptyIterator struct{}

func (emptyIterator) Next() bool         { return false }
func (emptyIterator) Value() interface{} { return nil }
func (emptyIterator) MaxSize() int       { return 0 }

// filterIterator filters the items of the underlying iterator
type filterIterator struct {
	itr     Iterator
	filter  func(interface{}) bool
	current interface{}
}

func (itr *filterIterator) Next() bool {
	for itr.itr.Next() {
		itr.current = itr.itr.Value()
		if itr.filter(itr.current) {
			return true
		}
		itr.current = nil
	}
	return false
}

func (itr *filterIterator) Value() interface{} {
	return itr.current
}

func (itr *filterIterator) MaxSize() int { return itr.itr.MaxSize() }

// makeUniqueIterator returns a filter iterator that will filter out duplicates
func makeUniqueIterator(itr Iterator) Iterator {
	seenItems := make(map[interface{}]struct{})
	return &filterIterator{
		itr: itr,
		filter: func(item interface{}) bool {
			if _, seen := seenItems[item]; seen {
				return false
			}
			seenItems[item] = struct{}{}
			return true
		},
	}
}

// funcIterator iterates through a set of underlying iterators obtained from a function
type funcIterator struct {
	// Returns a new iterator every time it is called. When returns nil, iteration stops
	iteratorFunc func() Iterator
	current      Iterator
}

func (itr *funcIterator) Next() bool {
	for {
		if itr.current != nil {
			if itr.current.Next() {
				return true
			}
			itr.current = nil
		}
		itr.current = itr.iteratorFunc()
		if itr.current == nil {
			return false
		}
	}
}

func (itr *funcIterator) Value() interface{} {
	return itr.current.Value()
}

func (itr *funcIterator) MaxSize() int { return -1 }

// MultiIterator returns an iterator that contatenates all the given iterators
func MultiIterator(iterators ...Iterator) Iterator {
	return &funcIterator{
		iteratorFunc: func() Iterator {
			if len(iterators) == 0 {
				return nil
			}
			ret := iterators[0]
			iterators = iterators[1:]
			return ret
		},
	}
}

// nodeIterator is a type-safe iterator for nodes
type nodeIterator struct {
	Iterator
}

func (n *nodeIterator) Node() Node {
	return n.Value().(Node)
}

// edgeIterator is a type-safe iterator for edges
type edgeIterator struct {
	Iterator
}

func (n *edgeIterator) Edge() Edge {
	return n.Value().(Edge)
}

type arrEdgeIterator struct {
	edges   []Edge
	current Edge
}

func (n *arrEdgeIterator) Next() bool {
	if len(n.edges) == 0 {
		return false
	}
	n.current = n.edges[0]
	n.edges = n.edges[1:]
	return true
}

func (n *arrEdgeIterator) Value() interface{} {
	return n.current
}

func (n *arrEdgeIterator) Edge() Edge {
	return n.current
}

func (n *arrEdgeIterator) MaxSize() int {
	return len(n.edges)
}

func NewEdgeIterator(edges ...Edge) EdgeIterator {
	return &arrEdgeIterator{edges: edges}
}

type iteratorWithoutSize interface {
	Next() bool
	Value() interface{}
}

type iteratorWithSize struct {
	itr  iteratorWithoutSize
	size int
}

func (i *iteratorWithSize) Next() bool         { return i.itr.Next() }
func (i *iteratorWithSize) Value() interface{} { return i.itr.Value() }
func (i *iteratorWithSize) MaxSize() int       { return i.size }

func withSize(itr iteratorWithoutSize, size int) Iterator {
	return &iteratorWithSize{
		itr:  itr,
		size: size}
}

type listIterator struct {
	next, current *list.Element
	size          int
}

func (l *listIterator) Next() bool {
	l.current = l.next
	if l.next != nil {
		l.next = l.next.Next()
	}
	return l.current != nil
}

func (l *listIterator) Value() interface{} {
	return l.current.Value
}

func (l *listIterator) MaxSize() int { return l.size }
