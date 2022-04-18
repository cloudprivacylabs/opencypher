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

import ()

type ErrNodeVariableExpected string

func (e ErrNodeVariableExpected) Error() string {
	return "Node variable expected: " + string(e)
}

type ErrEdgeVariableExpected string

func (e ErrEdgeVariableExpected) Error() string {
	return "Edge variable expected:" + string(e)
}

// Pattern contains pattern items, with even numbered elements
// corresponding to nodes, and odd numbered elements corresponding to
// edges
type Pattern []PatternItem

// A PatternSymbol contains either nodes, or edges.
type PatternSymbol struct {
	Nodes *NodeSet
	Edges *EdgeSet
}

// A PatternItem can be a node or an edge element of a pattern
type PatternItem struct {
	Labels     StringSet
	Properties map[string]interface{}
	// Min=-1 and Max=-1 for variable length
	Min       int
	Max       int
	Backwards bool
	// Name of the variable associated with this processing node. If the
	// name is defined, it is used to constrain values. If not, it is
	// used to store values
	Name string
}

func (p PatternItem) getEdgeFilter() func(Edge) bool {
	return GetEdgeFilterFunc(p.Labels, p.Properties)
}

func (p PatternItem) getNodeFilter() func(Node) bool {
	return GetNodeFilterFunc(p.Labels, p.Properties)
}

// Returns the set of nodes constraining the pattern item. That is,
// the set of nodes in the symbols
func (p PatternItem) isConstrainedNodes(ctx *MatchContext) (*NodeSet, error) {
	if len(p.Name) == 0 {
		return nil, nil
	}
	// Is this a symbol created in this pattern?
	sym, exists := ctx.LocalSymbols[p.Name]
	if exists {
		if sym.Edges != nil {
			return nil, ErrNodeVariableExpected(p.Name)
		}
		return sym.Nodes, nil
	}
	sym, exists = ctx.Symbols[p.Name]
	if exists {
		if sym.Edges != nil {
			return nil, ErrNodeVariableExpected(p.Name)
		}
		return sym.Nodes, nil
	}
	return nil, nil
}

// Returns the set of edges constraining the pattern item. That is,
// the set of edges in the symbols
func (p PatternItem) isConstrainedEdges(ctx *MatchContext) (*EdgeSet, error) {
	if len(p.Name) == 0 {
		return nil, nil
	}
	// Is this a symbol created in this pattern?
	sym, exists := ctx.LocalSymbols[p.Name]
	if exists {
		if sym.Nodes != nil {
			return nil, ErrEdgeVariableExpected(p.Name)
		}
		return sym.Edges, nil
	}
	sym, exists = ctx.Symbols[p.Name]
	if exists {
		if sym.Nodes != nil {
			return nil, ErrEdgeVariableExpected(p.Name)
		}
		return sym.Edges, nil
	}
	return nil, nil
}

func (p PatternItem) estimateNodeSize(gr Graph, symbols map[string]*PatternSymbol) (Iterator, int) {
	g := gr.(*OCGraph)
	max := -1
	var ret Iterator
	if len(p.Labels) > 0 {
		itr := g.index.nodesByLabel.IteratorAllLabels(p.Labels)
		if sz := itr.MaxSize(); sz != -1 {
			max = sz
			ret = itr
		}
	}
	if len(p.Properties) > 0 {
		for k, v := range p.Properties {
			itr := g.index.GetIteratorForNodeProperty(k, v)
			if itr == nil {
				continue
			}
			maxSize := itr.MaxSize()
			if maxSize == -1 {
				continue
			}
			if max == -1 || maxSize < max {
				max = maxSize
				ret = itr
			}
		}
	}
	if len(p.Name) > 0 {
		sym, ok := symbols[p.Name]
		if ok {
			if max == -1 || sym.Nodes.Len() < max {
				max = sym.Nodes.Len()
				ret = sym.Nodes.Iterator()
			}
		}
	}
	if ret == nil {
		ret = g.GetNodes()
		if sz := ret.MaxSize(); sz != -1 {
			max = sz
		}
	}
	return ret, max
}

func (p PatternItem) estimateEdgeSize(gr Graph, symbols map[string]*PatternSymbol) (Iterator, int) {
	g := gr.(*OCGraph)
	max := -1
	var ret Iterator

	allEdges := func() (Iterator, int) {
		ret := g.GetEdges()
		max := -1
		if sz := ret.MaxSize(); sz != -1 {
			max = sz
		}
		return ret, max
	}

	if p.Min > 1 || p.Max > 1 || p.Min == -1 || p.Max == -1 {
		return g.GetEdges(), -1
	}

	if len(p.Labels) > 0 {
		itr := g.GetEdgesWithAnyLabel(p.Labels)
		if sz := itr.MaxSize(); sz != -1 {
			max = sz
			ret = itr
		}
	}
	if len(p.Properties) > 0 {
		for k, v := range p.Properties {
			itr := g.index.GetIteratorForEdgeProperty(k, v)
			if itr == nil {
				continue
			}
			maxSize := itr.MaxSize()
			if maxSize == -1 {
				continue
			}
			if maxSize < max {
				max = maxSize
				ret = itr
			}
		}
	}
	if len(p.Name) > 0 {
		sym, ok := symbols[p.Name]
		if ok {
			if max == -1 || sym.Edges.Len() < max {
				max = sym.Edges.Len()
				ret = sym.Edges.Iterator()
			}
		}
	}
	if ret == nil {
		return allEdges()
	}
	return ret, max
}

func (p *PatternSymbol) Add(item interface{}) bool {
	switch k := item.(type) {
	case *OCNode:
		p.AddNode(k)

	case *OCEdge:
		if p.Edges == nil {
			p.Edges = &EdgeSet{}
		}
		p.Edges.Add(k)

	case []Edge:
		p.AddPath(k)
	}
	return true
}

func (p *PatternSymbol) AddNode(item Node) {
	if p.Nodes == nil {
		p.Nodes = &NodeSet{}
	}
	p.Nodes.Add(item.(*OCNode))
}

func (p *PatternSymbol) AddPath(path []Edge) {
	if p.Edges == nil {
		p.Edges = &EdgeSet{}
	}
	for _, x := range path {
		p.Edges.Add(x.(*OCEdge))
	}
}

func (p *PatternSymbol) NodeSlice() []Node {
	if p.Nodes != nil {
		return p.Nodes.Slice()
	}
	return nil
}

func (p *PatternSymbol) EdgeSlice() []Edge {
	if p.Edges != nil {
		return p.Edges.Slice()
	}
	return nil
}

type MatchPlan struct {
	steps []planProcessor
}

type planProcessor interface {
	Run(*MatchContext, matchAccumulator) error
	GetResult() interface{}
	GetPatternItem() PatternItem
}

type MatchAccumulator interface {
	// path is either a Node or []Edge, the matching path symbols
	// contains the current values for each symbol. The values of the
	// map is either Node or []Edge
	StoreResult(ctx *MatchContext, path interface{}, symbols map[string]interface{})
}

type MatchContext struct {
	Graph Graph
	// These are symbols that are used as constraints in the matching process.
	Symbols map[string]*PatternSymbol

	// localSymbols are symbols defined in the pattern.
	LocalSymbols map[string]*PatternSymbol
}

// If the current step has a local symbol, it will be recorded in the context
func (ctx *MatchContext) recordStepResult(step planProcessor) {
	name := step.GetPatternItem().Name
	if len(name) == 0 {
		return
	}
	if _, global := ctx.Symbols[name]; global {
		return
	}
	result := &PatternSymbol{}
	result.Add(step.GetResult())
	ctx.LocalSymbols[name] = result
}

// resetStepResult will remove the step's local symbol from the context
func (ctx *MatchContext) resetStepResult(step planProcessor) {
	name := step.GetPatternItem().Name
	if len(name) == 0 {
		return
	}
	if _, global := ctx.Symbols[name]; global {
		return
	}
	delete(ctx.LocalSymbols, name)
}

func (pattern Pattern) Run(graph Graph, symbols map[string]*PatternSymbol, result MatchAccumulator) error {
	plan, err := pattern.GetPlan(graph, symbols)
	if err != nil {
		return err
	}
	logf("Starting plan run\n")
	return plan.Run(graph, symbols, result)
}

func (pattern Pattern) FindPaths(graph Graph, symbols map[string]*PatternSymbol) (DefaultMatchAccumulator, error) {
	acc := DefaultMatchAccumulator{}
	if err := pattern.Run(graph, symbols, &acc); err != nil {
		return acc, err
	}
	return acc, nil
}

// FindNodes runs the pattern with the given symbols, and returns all the head nodes found
func (pattern Pattern) FindNodes(graph Graph, symbols map[string]*PatternSymbol) ([]Node, error) {
	acc := DefaultMatchAccumulator{}
	if err := pattern.Run(graph, symbols, &acc); err != nil {
		return nil, err
	}
	return acc.GetHeadNodes(), nil
}

func (pattern Pattern) getFastestElement(graph Graph, symbols map[string]*PatternSymbol) (Iterator, int) {
	maxSize := -1
	index := 0
	var itr Iterator
	for i := range pattern {
		sz := -1
		var t Iterator
		if (i % 2) == 0 {
			t, sz = pattern[i].estimateNodeSize(graph, symbols)
		} else {
			t, sz = pattern[i].estimateEdgeSize(graph, symbols)
		}
		if sz != -1 {
			if maxSize == -1 || sz < maxSize {
				maxSize = sz
				index = i
				itr = t
			}
		}
	}
	return itr, index
}

func (pattern Pattern) GetSymbolNames() StringSet {
	ret := NewStringSet()
	for _, p := range pattern {
		if len(p.Name) > 0 {
			ret.Add(p.Name)
		}
	}
	return ret
}

// GetPlan returns a match execution plan
func (pattern Pattern) GetPlan(graph Graph, symbols map[string]*PatternSymbol) (MatchPlan, error) {
	itr, index := pattern.getFastestElement(graph, symbols)
	plan := MatchPlan{}
	processors := make([]planProcessor, len(pattern))
	if (index % 2) == 0 {
		// start with a node
		processors[index] = &iterateNodes{itr: itr.(NodeIterator), patternItem: pattern[index]}
		plan.steps = append(plan.steps, processors[index])
		// Go forward
		for i := index + 1; i < len(pattern); i++ {
			if (i % 2) == 1 {
				// There is a node before this edge.
				if pattern[i].Backwards {
					// n<--
					processors[i] = newIterateConnectedEdges(processors[i-1], pattern[i], IncomingEdge)
				} else {
					// n-->
					processors[i] = newIterateConnectedEdges(processors[i-1], pattern[i], OutgoingEdge)
				}
			} else {
				// There is an edge before this node, and that determines the direction
				if pattern[i-1].Backwards {
					// <--n
					processors[i] = newIterateConnectedNodes(processors[i-1], pattern[i], useFromNode)
				} else {
					// -->n
					processors[i] = newIterateConnectedNodes(processors[i-1], pattern[i], useToNode)
				}
			}
			plan.steps = append(plan.steps, processors[i])
		}
		// Go backwards
		for i := index - 1; i >= 0; i-- {
			if (i % 2) == 1 {
				// There is a node after this edge
				if pattern[i].Backwards {
					// <--n
					processors[i] = newIterateConnectedEdges(processors[i+1], pattern[i], OutgoingEdge)
				} else {
					// -->n
					processors[i] = newIterateConnectedEdges(processors[i+1], pattern[i], IncomingEdge)
				}
			} else {
				// There is an edge after this node, and that determines the direction
				if pattern[i+1].Backwards {
					// n<--
					processors[i] = newIterateConnectedNodes(processors[i+1], pattern[i], useToNode)
				} else {
					// n-->
					processors[i] = newIterateConnectedNodes(processors[i+1], pattern[i], useFromNode)
				}
			}
			plan.steps = append(plan.steps, processors[i])
		}
	} else {
		// start with an edge
		processors[index] = &iterateEdges{itr: itr.(EdgeIterator), patternItem: pattern[index]}
		plan.steps = append(plan.steps, processors[index])
		// Go forward
		for i := index + 1; i < len(pattern); i++ {
			if (i % 2) == 1 {
				// There is a node before this edge
				if pattern[i].Backwards {
					// n<--
					processors[i] = newIterateConnectedEdges(processors[i-1], pattern[i], IncomingEdge)
				} else {
					// n-->
					processors[i] = newIterateConnectedEdges(processors[i-1], pattern[i], OutgoingEdge)
				}
			} else {
				// There is an edge before this node, and that determines the direction
				if pattern[i-1].Backwards {
					processors[i] = newIterateConnectedNodes(processors[i-1], pattern[i], useFromNode)
				} else {
					processors[i] = newIterateConnectedNodes(processors[i-1], pattern[i], useToNode)
				}
			}
			plan.steps = append(plan.steps, processors[i])
		}
		// Go backwards
		for i := index - 1; i >= 0; i-- {
			if (i % 2) == 1 {
				// There is a node after this edge
				if pattern[i].Backwards {
					// <--n
					processors[i] = newIterateConnectedEdges(processors[i+1], pattern[i], OutgoingEdge)
				} else {
					// -->n
					processors[i] = newIterateConnectedEdges(processors[i+1], pattern[i], IncomingEdge)
				}
			} else {
				// There is an edge after this node, and that determines the direction
				if pattern[i+1].Backwards {
					processors[i] = newIterateConnectedNodes(processors[i+1], pattern[i], useToNode)
				} else {
					processors[i] = newIterateConnectedNodes(processors[i+1], pattern[i], useFromNode)
				}
			}
			plan.steps = append(plan.steps, processors[i])
		}
	}
	return plan, nil
}

type matchAccumulator interface {
	Run(*MatchContext) error
}

type nextAccumulator struct {
	run  planProcessor
	next matchAccumulator
}

func (n nextAccumulator) Run(ctx *MatchContext) error {
	return n.run.Run(ctx, n.next)
}

type resultAccumulator struct {
	acc  MatchAccumulator
	plan MatchPlan
}

// Capture the current results
func (n *resultAccumulator) Run(ctx *MatchContext) error {
	n.acc.StoreResult(ctx, n.plan.GetCurrentPath(), n.plan.CaptureSymbolValues())
	return nil
}

// GetCurrentPath returns the current path recoded in the stages of the pattern. The result is either a single node, or a path
func (plan MatchPlan) GetCurrentPath() interface{} {
	if len(plan.steps) == 1 {
		return plan.steps[0].GetResult()
	}
	out := make([]Edge, 0)
	for i := range plan.steps {
		if edges, ok := plan.steps[i].GetResult().([]Edge); ok {
			out = append(out, edges...)
		}
	}
	return out
}

// CaptureSymbolValues captures the current symbol values as nodes or []Edges
func (plan MatchPlan) CaptureSymbolValues() map[string]interface{} {
	ret := make(map[string]interface{})
	for _, step := range plan.steps {
		if len(step.GetPatternItem().Name) > 0 {
			if _, exists := ret[step.GetPatternItem().Name]; !exists {
				ret[step.GetPatternItem().Name] = step.GetResult()
			}
		}
	}
	return ret
}

type DefaultMatchAccumulator struct {
	// Each element of the paths is either a Node or []Edge
	Paths   []interface{}
	Symbols []map[string]interface{}
}

func (acc *DefaultMatchAccumulator) StoreResult(_ *MatchContext, path interface{}, symbols map[string]interface{}) {
	acc.Paths = append(acc.Paths, path)
	acc.Symbols = append(acc.Symbols, symbols)
}

// Returns the unique nodes in the accumulator that start a path
func (acc *DefaultMatchAccumulator) GetHeadNodes() []Node {
	ret := make(map[Node]struct{})
	for _, x := range acc.Paths {
		if n, ok := x.(Node); ok {
			ret[n] = struct{}{}
		} else if e, ok := x.([]Edge); ok {
			ret[e[0].GetFrom()] = struct{}{}
		}
	}
	arr := make([]Node, 0, len(ret))
	for x := range ret {
		arr = append(arr, x)
	}
	return arr
}

// Returns the unique nodes in the accumulator that ends a path
func (acc *DefaultMatchAccumulator) GetTailNodes() []Node {
	ret := make(map[Node]struct{})
	for _, x := range acc.Paths {
		if n, ok := x.(Node); ok {
			ret[n] = struct{}{}
		} else if e, ok := x.([]Edge); ok {
			ret[e[len(e)-1].GetTo()] = struct{}{}
		}
	}
	arr := make([]Node, 0, len(ret))
	for x := range ret {
		arr = append(arr, x)
	}
	return arr
}

func (plan MatchPlan) Run(graph Graph, symbols map[string]*PatternSymbol, result MatchAccumulator) error {
	ctx := &MatchContext{
		Graph:        graph,
		Symbols:      symbols,
		LocalSymbols: make(map[string]*PatternSymbol),
	}

	res := resultAccumulator{acc: result, plan: plan}
	acc := matchAccumulator(&res)
	for i := len(plan.steps) - 1; i > 0; i-- {
		acc = nextAccumulator{
			run:  plan.steps[i],
			next: acc,
		}
	}
	return plan.steps[0].Run(ctx, acc)
}
