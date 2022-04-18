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
//	"fmt"
)

func logf(pattern string, args ...interface{}) {
	//fmt.Printf(pattern, args...)
}

type iterateNodes struct {
	itr         NodeIterator
	patternItem PatternItem
	result      Node
	initialized bool
}

func (processor *iterateNodes) init(ctx *MatchContext) error {
	if !processor.initialized {
		nodeFilter := processor.patternItem.getNodeFilter()
		processor.itr = &nodeIterator{
			&filterIterator{
				itr: processor.itr,
				filter: func(item interface{}) bool {
					return nodeFilter(item.(*OCNode))
				},
			},
		}
		nodes, err := processor.patternItem.isConstrainedNodes(ctx)
		if err != nil {
			return err
		}
		if nodes != nil {
			processor.itr = &nodeIterator{
				&filterIterator{
					itr: processor.itr,
					filter: func(item interface{}) bool {
						return nodes.Has(item.(*OCNode))
					},
				},
			}
		}
		processor.initialized = true
	}
	return nil
}

func (processor *iterateNodes) Run(ctx *MatchContext, next matchAccumulator) error {
	if err := processor.init(ctx); err != nil {
		return err
	}
	for processor.itr.Next() {
		processor.result = processor.itr.Node()
		ctx.recordStepResult(processor)
		logf("iterateNodes: %+v\n", processor.result)
		if err := next.Run(ctx); err != nil {
			return err
		}
		ctx.resetStepResult(processor)
	}
	return nil
}

func (processor *iterateNodes) GetPatternItem() PatternItem { return processor.patternItem }
func (processor *iterateNodes) GetResult() interface{}      { return processor.result }
func (processor *iterateNodes) IsNode()                     {}

type iterateEdges struct {
	itr         EdgeIterator
	patternItem PatternItem
	result      []Edge
	initialized bool
}

func (processor *iterateEdges) init(ctx *MatchContext) {
	if !processor.initialized {
		processor.initialized = true
		filterFunc := processor.patternItem.getEdgeFilter()
		processor.itr = &edgeIterator{
			&filterIterator{
				itr: processor.itr,
				filter: func(edge interface{}) bool {
					return filterFunc(edge.(*OCEdge))
				},
			},
		}
	}
}

func (processor *iterateEdges) Run(ctx *MatchContext, next matchAccumulator) error {
	processor.init(ctx)
	for processor.itr.Next() {
		processor.result = []Edge{processor.itr.Edge()}
		ctx.recordStepResult(processor)
		if err := next.Run(ctx); err != nil {
			return err
		}
		ctx.resetStepResult(processor)
	}
	return nil
}

func (processor *iterateEdges) GetPatternItem() PatternItem { return processor.patternItem }
func (processor *iterateEdges) GetResult() interface{}      { return processor.result }
func (processor *iterateEdges) IsEdge()                     {}

type iterateConnectedEdges struct {
	patternItem PatternItem
	source      planProcessor
	dir         EdgeDir
	result      []Edge
	edgeFilter  func(Edge) bool
	edgeItr     EdgeIterator
}

func newIterateConnectedEdges(source planProcessor, item PatternItem, dir EdgeDir) *iterateConnectedEdges {
	return &iterateConnectedEdges{
		patternItem: item,
		source:      source,
		dir:         dir,
		edgeFilter:  item.getEdgeFilter(),
	}
}

func (processor *iterateConnectedEdges) init(ctx *MatchContext) {
	if processor.edgeItr == nil {
		node := processor.source.GetResult().(Node)
		processor.edgeItr = &edgeIterator{
			&filterIterator{
				itr: node.GetEdgesWithAnyLabel(processor.dir, processor.patternItem.Labels),
				filter: func(item interface{}) bool {
					return processor.edgeFilter(item.(*OCEdge))
				},
			},
		}
	}
}

func (processor *iterateConnectedEdges) Run(ctx *MatchContext, next matchAccumulator) error {
	processor.init(ctx)
	if processor.edgeItr == nil {
		return nil
	}
	if processor.patternItem.Min == 1 && processor.patternItem.Max == 1 {
		for processor.edgeItr.Next() {
			processor.result = []Edge{processor.edgeItr.Edge()}
			ctx.recordStepResult(processor)
			logf("IterateConnectedEdges len=1 %+v\n", processor.result)
			if err := next.Run(ctx); err != nil {
				return err
			}
			ctx.resetStepResult(processor)
		}
		return nil
	}
	var err error
	CollectAllPaths(ctx.Graph, processor.edgeItr, processor.edgeFilter, processor.dir, processor.patternItem.Min, processor.patternItem.Max, func(path []Edge) bool {
		processor.result = path
		logf("IterateConnectedEdges len>1 %+v\n", processor.result)
		ctx.recordStepResult(processor)
		if err = next.Run(ctx); err != nil {
			return false
		}
		ctx.resetStepResult(processor)
		return true
	})
	return err
}

func (processor *iterateConnectedEdges) GetPatternItem() PatternItem { return processor.patternItem }
func (processor *iterateConnectedEdges) GetResult() interface{}      { return processor.result }
func (processor *iterateConnectedEdges) IsEdge()                     {}

const useFromNode = -1
const useToNode = 1

type iterateConnectedNodes struct {
	patternItem PatternItem
	source      planProcessor
	useNode     int
	result      Node
	nodeFilter  func(Node) bool
}

func newIterateConnectedNodes(source planProcessor, item PatternItem, useNode int) *iterateConnectedNodes {
	return &iterateConnectedNodes{
		patternItem: item,
		source:      source,
		useNode:     useNode,
		nodeFilter:  item.getNodeFilter(),
	}
}

func (processor *iterateConnectedNodes) Run(ctx *MatchContext, next matchAccumulator) error {
	edges := processor.source.GetResult().([]Edge)
	edge := edges[len(edges)-1]
	var node Node
	if processor.useNode == useToNode {
		node = edge.GetTo()
	} else {
		node = edge.GetFrom()
	}
	logf("Iterate connected nodes with node=%+v\n", node)
	if processor.nodeFilter(node) {

		constraints, err := processor.patternItem.isConstrainedNodes(ctx)
		if err != nil {
			return err
		}
		if constraints != nil {
			if !constraints.Has(node.(*OCNode)) {
				return nil
			}
		}

		processor.result = node
		ctx.recordStepResult(processor)
		logf("Iterate connected nodes goes deeper\n")
		if err := next.Run(ctx); err != nil {
			return err
		}
		ctx.resetStepResult(processor)
	}
	return nil
}

func (processor *iterateConnectedNodes) GetPatternItem() PatternItem { return processor.patternItem }
func (processor *iterateConnectedNodes) GetResult() interface{}      { return processor.result }
func (processor *iterateConnectedNodes) IsNode()                     {}
