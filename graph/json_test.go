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
	"bytes"
	"encoding/json"
	"testing"
)

func TestEncode(t *testing.T) {
	g := NewOCGraph()
	j := JSON{}
	buf := bytes.Buffer{}
	if err := j.Encode(g, &buf); err != nil {
		t.Error(err)
	}
	if buf.String() != "{}" {
		t.Errorf("Wrong out: %s", buf.String())
	}
	n1 := g.NewNode([]string{"lbl"}, map[string]interface{}{"key": "value"})
	buf = bytes.Buffer{}
	if err := j.Encode(g, &buf); err != nil {
		t.Error(err)
	}
	if buf.String() != `{"nodes":[{"n":0,"labels":["lbl"],"properties":{"key":"value"}}]}` {
		t.Errorf("Wrong out: %s", buf.String())
	}

	n2 := g.NewNode([]string{"a", "b"}, nil)
	g.NewEdge(n1, n2, "edge", nil)
	buf = bytes.Buffer{}
	if err := j.Encode(g, &buf); err != nil {
		t.Error(err)
	}
	t.Log(buf.String())
	newg := NewOCGraph()
	if err := j.Decode(newg, json.NewDecoder(&buf)); err != nil {
		t.Errorf("Err: %v %s", err, buf.String())
	}

}
