package opencypher

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/cloudprivacylabs/opencypher/graph"
)

var ErrRowsHaveDifferentSizes = errors.New("Rows have different sizes")
var ErrIncompatibleCells = errors.New("Incompatible result set cells")

// ResultSet is a table of values
type ResultSet struct {
	Nodes graph.NodeSet
	Edges graph.EdgeSet

	Rows []map[string]Value
}

func isCompatibleValue(v1, v2 Value) bool {
	if v1.Value == nil {
		if v2.Value == nil {
			return true
		}
		return false
	}
	if v2.Value == nil {
		return false
	}
	return reflect.TypeOf(v1.Value) == reflect.TypeOf(v2.Value)
}

// Check if all row cells are compatible
func isCompatibleRow(row1, row2 map[string]Value) error {
	if len(row1) != len(row2) {
		return ErrRowsHaveDifferentSizes
	}
	for k := range row1 {
		if !isCompatibleValue(row1[k], row2[k]) {
			return ErrIncompatibleCells
		}
	}
	return nil
}

func (r *ResultSet) find(row map[string]Value) int {
	for index, r := range r.Rows {
		if len(r) != len(row) {
			break
		}
		found := true
		for i := range r {
			if !r[i].IsSame(row[i]) {
				found = false
				break
			}
		}
		if found {
			return index
		}
	}
	return -1
}

// Append the row to the resultset.
func (r *ResultSet) Append(row map[string]Value) error {
	r.Rows = append(r.Rows, row)
	for _, v := range row {
		switch val := v.Value.(type) {
		case graph.Node:
			r.Nodes.Add(val.(*graph.OCNode))
		case []graph.Edge:
			for _, edge := range val {
				r.Edges.Add(edge.(*graph.OCEdge))
			}
		case graph.Edge:
			r.Edges.Add(val.(*graph.OCEdge))
		}
	}
	return nil
}

func (r *ResultSet) AddPath(node graph.Node, edges []graph.Edge) {
	if node != nil {
		r.Nodes.Add(node.(*graph.OCNode))
	}
	for _, e := range edges {
		r.Edges.Add(e.(*graph.OCEdge))
	}
}

func (r *ResultSet) Add(rs ResultSet) {
	r.Rows = append(r.Rows, rs.Rows...)
	for itr := rs.Nodes.Iterator(); itr.Next(); {
		r.Nodes.Add(itr.Node().(*graph.OCNode))
	}
	for itr := rs.Edges.Iterator(); itr.Next(); {
		r.Edges.Add(itr.Edge().(*graph.OCEdge))
	}
}

// Union adds the src resultset to this. If all is set, it adds all rows, otherwise, it adds unique rows
func (r *ResultSet) Union(src ResultSet, all bool) error {
	for _, sourceRow := range src.Rows {
		appnd := all
		if !appnd && r.find(sourceRow) != -1 {
			appnd = true
		}
		if appnd {
			if err := r.Append(sourceRow); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r ResultSet) String() string {
	out := bytes.Buffer{}
	for _, row := range r.Rows {
		for k, v := range row {
			io.WriteString(&out, fmt.Sprintf("%s: %s ", k, v))
		}
		io.WriteString(&out, "\n")
	}
	return out.String()
}
