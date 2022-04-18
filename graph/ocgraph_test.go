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
	"testing"
)

func TestGraphCRUD(t *testing.T) {
	g := NewOCGraph()
	nodes := make([]Node, 0)
	for i := 0; i < 10; i++ {
		nodes = append(nodes, g.NewNode([]string{fmt.Sprint(i)}, nil))
	}
	for i := 0; i < len(nodes)-1; i++ {
		g.NewEdge(nodes[i], nodes[i+1], "e", nil)
	}

	if len(NodeSlice(g.GetNodes())) != len(nodes) {
		t.Errorf("Wrong node count")
	}
	if g.NumNodes() != len(nodes) {
		t.Errorf("Wrong numNodes")
	}
	nodes[2].DetachAndRemove()
	if len(NodeSlice(g.GetNodes())) != len(nodes)-1 {
		t.Errorf("Wrong node count")
	}
	if g.NumNodes() != len(nodes)-1 {
		t.Errorf("Wrong numNodes")
	}

}
