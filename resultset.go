package opencypher

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"unicode"

	"github.com/cloudprivacylabs/lpg"
)

var ErrRowsHaveDifferentSizes = errors.New("Rows have different sizes")
var ErrIncompatibleCells = errors.New("Incompatible result set cells")

// ResultSet is a table of values
type ResultSet struct {
	Nodes lpg.NodeSet
	Edges lpg.EdgeSet

	Rows []map[string]Value
}

func NewResultSet() *ResultSet {
	return &ResultSet{
		Nodes: *lpg.NewNodeSet(),
		Edges: *lpg.NewEdgeSet(),
	}
}

func isCompatibleValue(v1, v2 Value) bool {
	if v1.Get() == nil {
		if v2.Get() == nil {
			return true
		}
		return false
	}
	if v2.Get() == nil {
		return false
	}
	return reflect.TypeOf(v1.Get()) == reflect.TypeOf(v2.Get())
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
			if !IsValueSame(r[i], row[i]) {
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

// Column returns a column of results as a value
func (r *ResultSet) Column(key string) []Value {
	ret := make([]Value, 0)
	for _, row := range r.Rows {
		val, exists := row[key]
		if !exists {
			continue
		}
		ret = append(ret, val)
	}
	return ret
}

// Append the row to the resultset.
func (r *ResultSet) Append(row map[string]Value) error {
	r.Rows = append(r.Rows, row)
	for _, v := range row {
		switch val := v.Get().(type) {
		case *lpg.Node:
			r.Nodes.Add(val)
		case []*lpg.Edge:
			for _, edge := range val {
				r.Edges.Add(edge)
			}
		case *lpg.Edge:
			r.Edges.Add(val)
		}
	}
	return nil
}

func (r *ResultSet) AddPath(node *lpg.Node, edges []*lpg.Edge) {
	if node != nil {
		r.Nodes.Add(node)
	}
	for _, e := range edges {
		r.Edges.Add(e)
	}
}

func (r *ResultSet) Add(rs ResultSet) {
	r.Rows = append(r.Rows, rs.Rows...)
	for itr := rs.Nodes.Iterator(); itr.Next(); {
		r.Nodes.Add(itr.Node())
	}
	for itr := rs.Edges.Iterator(); itr.Next(); {
		r.Edges.Add(itr.Edge())
	}
}

// Union adds the src resultset to this. If all is set, it adds all rows, otherwise, it adds unique rows
func (r *ResultSet) Union(src ResultSet, all bool) error {
	for _, sourceRow := range src.Rows {
		appnd := all
		if !appnd && r.find(sourceRow) == -1 {
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

// CartesianProduct calls f with all permutations of rows until f
// returns false. The map passed to f is reused, so copy if you need a
// copy of it.
func (r ResultSet) CartesianProduct(f func(map[string]Value) bool) bool {
	if len(r.Rows) == 0 {
		return true
	}
	ctr := make([]int, len(r.Rows[0]))
	keys := make([]string, 0, len(r.Rows[0]))
	for k := range r.Rows[0] {
		keys = append(keys, k)
	}
	data := make(map[string]Value)
	for {
		for k, i := range ctr {
			key := keys[k]
			data[key] = r.Rows[i][key]
		}
		if !f(data) {
			return false
		}
		carry := false
		for i := range ctr {
			ctr[i]++
			if ctr[i] >= len(r.Rows) {
				ctr[i] = 0
				carry = true
				continue
			}
			carry = false
			break
		}
		if carry {
			break
		}
	}
	return true
}

// CartesianProuduct builds the product of all the resultsets
func CartesianProduct(resultsets []ResultSet, all bool, filter func(map[string]Value) bool) ResultSet {
	ctr := make([]int, len(resultsets))
	result := *NewResultSet()
	for {
		data := make(map[string]Value)
		for i := range ctr {
			var row map[string]Value
			if ctr[i] >= len(resultsets[i].Rows) {
				if !all {
					return *NewResultSet()
				}
			} else {
				row = resultsets[i].Rows[ctr[i]]
			}
			for k, v := range row {
				data[k] = v
			}
		}
		if filter(data) {
			result.Rows = append(result.Rows, data)
		}
		carry := false
		for i := range ctr {
			ctr[i]++
			if ctr[i] >= len(resultsets[i].Rows) {
				ctr[i] = 0
				carry = true
				continue
			}
			carry = false
			break
		}
		if carry {
			break
		}
	}
	return result
}

// IsNamedResult checks if the given name is a symbol name (i.e. it does not start with a number)
func IsNamedResult(name string) bool {
	for _, r := range name {
		if unicode.IsDigit(r) {
			return false
		}
		return true
	}
	return false
}
