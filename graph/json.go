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
	"encoding/json"
	"io"
	"strconv"
)

// Interner is a string interface, that can be as simple as a
// map[string]string, that is used to intern property keys
type Interner interface {
	// Intern should return the interned copy of the string
	Intern(string) string
}

// MapInterner is a basic interner that uses a map[string]string to intern strings
type MapInterner map[string]string

// Intern keeps the interned copy of s in the map
func (i MapInterner) Intern(s string) string {
	if k, exists := i[s]; exists {
		return k
	}
	i[s] = s
	return s
}

// JSON marshals/unmarshals a graph to/from JSON.
//
// The JSON unmarshaling uses a string Interner if it is set. This
// allows reusing the strings for repeated keys. Setting this interner
// to an implementation will reduce memory footprint of the graph. If
// letf uninitialized, the unmarshaler still uses an interner, but
// this interner will not be shared between different graphs
// unmarshaled using this unmarshaler.
//
// If PropertyMarshaler/PropertyUnmarshaler is set, these functions
// are called to marshal/unmarshal individual properties. These
// function can return an empty string for the key to skip the
// property.
//
// If MarshalEdgesSeparately is set, graph edges are marshaled under
// the "edges" key. If this is false, the graph edges are included
// under the source node.
//
//
// The JSON representation of nodes use a numeric node index. These
// indexes are used to refer to nodes in edges.
//
// With MarshalEdgesSeparately=false:
//
//     {
//        "nodes": [
//          {
//             "n": 0,
//             "labels": ["lbl1","lbl2"],
//             "properties": {
//               "key": "value"
//             },
//             "edges": [
//                {
//                    "to": 1,
//                    "label": "edgeLabel"
//                }
//             ]
//          },
//          {
//             "n": 1,
//             "labels": ["lbl1"]
//          }
//       ]
//     }
//
// With MarshalEdgesSeparately=true:
//
//     {
//        "nodes": [
//          {
//             "n": 0,
//             "labels": ["lbl1","lbl2"],
//             "properties": {
//               "key": "value"
//             }
//          },
//          {
//             "n": 1,
//             "labels": ["lbl1"]
//          }
//       ],
//       "edges": [
//            {
//                 "from": 0,
//                 "to": 1,
//                 "label": "edgeLabel"
//           }
//       ]
//     }
//
//
type JSON struct {
	Interner Interner

	// If false, dring marshaling edges are embedded in the source
	// nodes. If true, edges are marshaled separately as a JSON array.
	MarshalEdgesSeparately bool

	// PropertyUnmarshaler unmarshals a property value. The return
	// values are the key, value, and possible error. If the returned
	// key is empty, the property is not unmarshaled. If this is nil,
	// default json unmarshaler is used for property value.
	PropertyUnmarshaler func(key string, value json.RawMessage) (string, interface{}, error)

	// PropertyMarshaler marshals a property value. The return values
	// are the key, the marshaled value, and possible error. If the
	// returned key is empty, the property is not marshaled. If this is
	// nil, the default json marshaler is used for property value.
	PropertyMarshaler func(key string, value interface{}) (string, json.RawMessage, error)
}

// jsonNode contains the graph representation of a JSON node
type jsonNode struct {
	N          int                        `json:"n"`
	Labels     []string                   `json:"labels,omitempty"`
	Properties map[string]json.RawMessage `json:"properties,omitempty"`
	Edges      []jsonOutgoingEdge         `json:"edges,omitempty"`
}

// jsonEdge contains the graph representation of a JSON edge, using
// node indexes to address nodes
type jsonEdge struct {
	From       int                        `json:"from"`
	To         int                        `json:"to"`
	Label      string                     `json:"label,omitempty"`
	Properties map[string]json.RawMessage `json:"properties,omitempty"`
}

// jsonOutgoingEdge contains an edge included in a node
type jsonOutgoingEdge struct {
	To         int                        `json:"to"`
	Label      string                     `json:"label,omitempty"`
	Properties map[string]json.RawMessage `json:"properties,omitempty"`
}

var (
	objBegin      = []byte{'{'}
	objEnd        = []byte{'}'}
	nKey          = []byte(`"n":`)
	comma         = []byte{','}
	labelsKey     = []byte(`"labels":`)
	propertiesKey = []byte(`"properties":`)
	edgesKey      = []byte(`"edges":`)
	nodesKey      = []byte(`"nodes":`)
	arrBegin      = []byte{'['}
	arrEnd        = []byte{']'}
)

// Encode the graph in JSON
func (j JSON) Encode(g Graph, out io.Writer) error {
	nodeMap := make(map[*OCNode]int)

	marshalProperties := func(in map[string]interface{}) (map[string]json.RawMessage, error) {
		ret := make(map[string]json.RawMessage)
		for k, v := range in {
			if j.PropertyMarshaler == nil {
				d, _ := json.Marshal(v)
				ret[k] = d
			} else {
				k, d, err := j.PropertyMarshaler(k, v)
				if err != nil {
					return nil, err
				}
				if len(k) > 0 {
					ret[k] = d
				}
			}
		}
		return ret, nil
	}

	encodeEdge := func(edge *OCEdge, writeFrom bool) error {
		var e interface{}
		properties, err := marshalProperties(edge.Properties)
		if err != nil {
			return err
		}
		if writeFrom {
			e = jsonEdge{
				From:       nodeMap[edge.GetFrom().(*OCNode)],
				To:         nodeMap[edge.GetTo().(*OCNode)],
				Label:      edge.label,
				Properties: properties,
			}
		} else {
			e = jsonOutgoingEdge{
				To:         nodeMap[edge.GetTo().(*OCNode)],
				Label:      edge.label,
				Properties: properties,
			}
		}
		data, err := json.Marshal(e)
		if err != nil {
			return err
		}
		if _, err := out.Write(data); err != nil {
			return err
		}
		return nil
	}

	i := 0
	// Give each node an index
	for nodes := g.GetNodes(); nodes.Next(); {
		node := nodes.Node().(*OCNode)
		nodeMap[node] = i
		i++
	}

	if _, err := out.Write(objBegin); err != nil {
		return err
	}
	if len(nodeMap) > 0 {
		if _, err := out.Write(nodesKey); err != nil {
			return err
		}
		if _, err := out.Write(arrBegin); err != nil {
			return err
		}

		// Write nodes
		firstNode := true
		for node, ix := range nodeMap {
			if firstNode {
				firstNode = false
			} else {
				if _, err := out.Write(comma); err != nil {
					return err
				}
			}
			if _, err := out.Write(objBegin); err != nil {
				return err
			}
			if _, err := out.Write(nKey); err != nil {
				return err
			}
			s := strconv.Itoa(ix)
			if _, err := out.Write([]byte(s)); err != nil {
				return err
			}
			if node.labels.Len() > 0 {
				data, _ := json.Marshal(node.labels)
				if _, err := out.Write(comma); err != nil {
					return err
				}
				if _, err := out.Write(labelsKey); err != nil {
					return err
				}
				if _, err := out.Write(data); err != nil {
					return err
				}
			}
			if len(node.Properties) > 0 {
				if _, err := out.Write(comma); err != nil {
					return err
				}
				if _, err := out.Write(propertiesKey); err != nil {
					return err
				}
				data, err := json.Marshal(node.Properties)
				if err != nil {
					return err
				}
				if _, err := out.Write(data); err != nil {
					return err
				}
			}

			if j.MarshalEdgesSeparately {
				if _, err := out.Write(objEnd); err != nil {
					return err
				}
				continue
			}

			first := true
			for edges := node.GetEdges(OutgoingEdge); edges.Next(); {
				edge := edges.Edge().(*OCEdge)

				if first {
					if _, err := out.Write(comma); err != nil {
						return err
					}
					if _, err := out.Write(edgesKey); err != nil {
						return err
					}
					if _, err := out.Write(arrBegin); err != nil {
						return err
					}
					first = false
				} else {
					if _, err := out.Write(comma); err != nil {
						return err
					}
				}
				if err := encodeEdge(edge, false); err != nil {
					return err
				}
			}
			if !first {
				if _, err := out.Write(arrEnd); err != nil {
					return err
				}
			}
			if _, err := out.Write(objEnd); err != nil {
				return err
			}
		}
		if _, err := out.Write(arrEnd); err != nil {
			return err
		}
	}

	if j.MarshalEdgesSeparately {
		if _, err := out.Write(comma); err != nil {
			return err
		}
		if _, err := out.Write(edgesKey); err != nil {
			return err
		}
		if _, err := out.Write(arrBegin); err != nil {
			return err
		}
		first := true
		for edges := g.GetEdges(); edges.Next(); {
			edge := edges.Edge().(*OCEdge)
			if first {
				first = false
			} else {
				if _, err := out.Write(comma); err != nil {
					return err
				}
			}
			if err := encodeEdge(edge, true); err != nil {
				return err
			}
		}
		if _, err := out.Write(arrEnd); err != nil {
			return err
		}
	}

	if _, err := out.Write(objEnd); err != nil {
		return err
	}

	return nil
}

func (j JSON) unmarshalProperty(key string, value json.RawMessage) (string, interface{}, error) {
	if j.PropertyUnmarshaler == nil {
		var v interface{}
		if err := json.Unmarshal(value, &v); err != nil {
			return "", nil, err
		}
		return key, v, nil
	}
	return j.PropertyUnmarshaler(key, value)
}

func (j JSON) unmarshalProperties(input map[string]json.RawMessage) (map[string]interface{}, error) {
	if input == nil {
		return nil, nil
	}
	ret := make(map[string]interface{})
	for k, v := range input {
		k, v, err := j.unmarshalProperty(k, v)
		if err != nil {
			return nil, err
		}
		if len(k) > 0 {
			k = j.Interner.Intern(k)
			ret[k] = v
		}
	}
	return ret, nil
}

func (j JSON) addNode(g Graph, node jsonNode, nodeMap map[int]*OCNode, edgeQueue []jsonEdge) ([]jsonEdge, error) {
	p, err := j.unmarshalProperties(node.Properties)
	if err != nil {
		return nil, err
	}
	for i := range node.Labels {
		node.Labels[i] = j.Interner.Intern(node.Labels[i])
	}
	newNode := g.NewNode(node.Labels, p).(*OCNode)
	nodeMap[node.N] = newNode
	for _, edge := range node.Edges {
		to, ok := nodeMap[edge.To]
		if !ok {
			edgeQueue = append(edgeQueue, jsonEdge{
				From:       node.N,
				To:         edge.To,
				Label:      edge.Label,
				Properties: edge.Properties,
			})
		} else {
			if err := j.addEdge(g, newNode, to, edge.Label, edge.Properties); err != nil {
				return nil, err
			}
		}
	}
	return edgeQueue, nil
}

func (j JSON) addEdge(g Graph, from, to *OCNode, label string, properties map[string]json.RawMessage) error {
	p, err := j.unmarshalProperties(properties)
	if err != nil {
		return err
	}
	g.NewEdge(from, to, label, p)
	return nil
}

// Decode a graph in JSON
func (j JSON) Decode(g Graph, input *json.Decoder) error {
	if j.Interner == nil {
		j.Interner = make(MapInterner)
	}
	tok, err := input.Token()
	if err != nil {
		return err
	}
	if tok != json.Delim('{') {
		return &json.SyntaxError{Offset: input.InputOffset()}
	}
	depth := 0
	nodeMap := make(map[int]*OCNode)
	edgeQueue := make([]jsonEdge, 0)
	for {
		tok, err = input.Token()
		if err != nil {
			return err
		}
		if depth == 0 {
			if s, ok := tok.(string); ok {
				if s == "nodes" {
					var nodes []jsonNode
					if err := input.Decode(&nodes); err != nil {
						return err
					}
					for _, node := range nodes {
						if edgeQueue, err = j.addNode(g, node, nodeMap, edgeQueue); err != nil {
							return err
						}
					}
				} else if s == "edges" {
					var edges []jsonEdge
					if err := input.Decode(&edges); err != nil {
						return err
					}
					for _, edge := range edges {
						from, fromExists := nodeMap[edge.From]
						to, toExists := nodeMap[edge.To]
						if fromExists && toExists {
							if err := j.addEdge(g, from, to, edge.Label, edge.Properties); err != nil {
								return err
							}
						} else {
							edgeQueue = append(edgeQueue, edge)
						}
					}
				} else {
					depth = 1
				}
			} else if tok == json.Delim('}') {
				return nil
			}
		} else {
			switch tok {
			case json.Delim('{'), json.Delim('['):
				depth++
			case json.Delim('}'), json.Delim(']'):
				depth--
			default:
				depth--
			}
		}
	}
}
