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
	"fmt"
)

type OCEdge struct {
	from, to *OCNode
	label    string
	Properties
	id int
}

func (edge *OCEdge) GetID() int       { return edge.id }
func (edge *OCEdge) GetGraph() Graph  { return edge.from.graph }
func (edge *OCEdge) GetLabel() string { return edge.label }
func (edge *OCEdge) GetFrom() Node    { return edge.from }
func (edge *OCEdge) GetTo() Node      { return edge.to }

func (edge *OCEdge) SetLabel(label string) {
	edge.from.graph.SetEdgeLabel(edge, label)
}

func (edge *OCEdge) SetProperty(key string, value interface{}) {
	edge.from.graph.SetEdgeProperty(edge, key, value)
}

func (edge *OCEdge) RemoveProperty(key string) {
	edge.from.graph.RemoveEdgeProperty(edge, key)
}

// Remove an edge
func (edge *OCEdge) Remove() {
	edge.from.graph.RemoveEdge(edge)
}

func (edge *OCEdge) String() string {
	return fmt.Sprintf("[:%s %s]", edge.label, edge.Properties)
}
