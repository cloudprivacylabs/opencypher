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

type OCEdge struct {
	id       int
	from, to *OCNode
	label    string
	Properties
	graph *OCGraph
}

func (edge *OCEdge) GetGraph() Graph  { return edge.graph }
func (edge *OCEdge) GetLabel() string { return edge.label }
func (edge *OCEdge) GetFrom() Node    { return edge.from }
func (edge *OCEdge) GetTo() Node      { return edge.to }

func (edge *OCEdge) SetLabel(label string) {
	edge.graph.SetEdgeLabel(edge, label)
}

func (edge *OCEdge) SetProperty(key string, value interface{}) {
	edge.graph.SetEdgeProperty(edge, key, value)
}

func (edge *OCEdge) RemoveProperty(key string) {
	edge.graph.RemoveEdgeProperty(edge, key)
}

// Remove an edge
func (edge *OCEdge) Remove() {
	edge.graph.RemoveEdge(edge)
}
