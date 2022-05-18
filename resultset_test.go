package opencypher

import (
	"testing"
)

func TestCartesianProduct(t *testing.T) {
	r := ResultSet{Rows: []map[string]Value{
		{"a": RValue{Value: 0}, "b": RValue{Value: 1}, "c": RValue{Value: 2}},
		{"a": RValue{Value: 10}, "b": RValue{Value: 11}, "c": RValue{Value: 12}},
		{"a": RValue{Value: 20}, "b": RValue{Value: 21}, "c": RValue{Value: 22}},
	}}
	expected := [][3]int{
		{0, 1, 2},
		{10, 1, 2},
		{20, 1, 2},
		{0, 11, 2},
		{10, 11, 2},
		{20, 11, 2},
		{0, 21, 2},
		{10, 21, 2},
		{20, 21, 2},
		{0, 1, 12},
		{10, 1, 12},
		{20, 1, 12},
		{0, 11, 12},
		{10, 11, 12},
		{20, 11, 12},
		{0, 21, 12},
		{10, 21, 12},
		{20, 21, 12},
		{0, 1, 22},
		{10, 1, 22},
		{20, 1, 22},
		{0, 11, 22},
		{10, 11, 22},
		{20, 11, 22},
		{0, 21, 22},
		{10, 21, 22},
		{20, 21, 22},
	}

	n := 0
	r.CartesianProduct(func(m map[string]Value) bool {
		n++
		v := [3]int{m["a"].Get().(int),
			m["b"].Get().(int),
			m["c"].Get().(int),
		}
		for _, x := range expected {
			if x == v {
				return true
			}
		}
		t.Errorf("Not found: %v", m)
		return false
	})
	if n != len(expected) {
		t.Errorf("Got %d results", n)
	}
}
