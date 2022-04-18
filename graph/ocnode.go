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
	"strings"
)

type OCNode struct {
	id     int
	labels StringSet
	Properties
	graph *OCGraph
}

func (node *OCNode) GetGraph() Graph      { return node.graph }
func (node *OCNode) GetLabels() StringSet { return node.labels.Clone() }

// Returns an edge iterator for incoming or outgoing edges
func (node *OCNode) GetEdges(dir EdgeDir) EdgeIterator {
	return node.graph.GetNodeEdges(node, dir)
}

// Returns an edge iterator for incoming or outgoing edges with the given label
func (node *OCNode) GetEdgesWithLabel(dir EdgeDir, label string) EdgeIterator {
	return node.graph.GetNodeEdgesWithLabel(node, dir, label)
}

// Returns an edge iterator for incoming or outgoingn edges that has the given labels
func (node *OCNode) GetEdgesWithAnyLabel(dir EdgeDir, labels StringSet) EdgeIterator {
	return node.graph.GetNodeEdgesWithAnyLabel(node, dir, labels)
}

func (node *OCNode) SetLabels(labels StringSet) {
	node.graph.SetNodeLabels(node, labels)
}

func (node *OCNode) SetProperty(key string, value interface{}) {
	node.graph.SetNodeProperty(node, key, value)
}

func (node *OCNode) RemoveProperty(key string) {
	node.graph.RemoveNodeProperty(node, key)
}

// Remove all connected edges, and remove the node
func (node *OCNode) DetachAndRemove() {
	node.graph.DetachRemoveNode(node)
}

// Remove all connected edges
func (node *OCNode) Detach() {
	node.graph.DetachNode(node)
}

func (node OCNode) String() string {
	labels := strings.Join(node.labels.Slice(), ":")
	if len(node.labels) > 0 {
		labels = ":" + labels
	}
	return fmt.Sprintf("(%s %s)", labels, node.Properties)
}
