package opencypher

import (
	"fmt"
	"strings"

	"github.com/neo4j/neo4j-go-driver/neo4j"

	"github.com/cloudprivacylabs/opencypher/graph"
)

// Value represents a computed value. Possible data types it can contain are:
//
//   primitives:
//    int
//    float64
//    bool
//    string
//    neo4j.Duration
//    neo4j.Date
//    neo4j.LocalDateTime
//    neo4j.LocalTime
//
//  composites:
//    []Value
//    map[string]Value
//    graph.StringSet
//    Node
//    []Edge
//    ResultSet

type Value interface {
	Evaluatable
	Get() interface{}
	IsConst() bool
}

// LValue is a pointer to a value
type LValue struct {
	getter func() interface{}
	setter func(interface{})
}

func (v LValue) Get() interface{}                     { return v.getter() }
func (v LValue) Set(val interface{})                  { v.setter(val) }
func (LValue) IsConst() bool                          { return false }
func (v LValue) Evaluate(*EvalContext) (Value, error) { return v, nil }

// NewLValue returns an LValue from the given value
func NewLValue(v Value) LValue {
	l, ok := v.(LValue)
	if ok {
		return l
	}
	r := v.(RValue)
	return LValue{
		getter: func() interface{} {
			return r.Value
		},
		setter: func(toValue interface{}) {
			r.Value = toValue
		},
	}
}

// RValue is a value
type RValue struct {
	Value interface{}
	Const bool
}

func (v RValue) Get() interface{}                     { return v.Value }
func (v RValue) IsConst() bool                        { return v.Const }
func (v RValue) Evaluate(*EvalContext) (Value, error) { return v, nil }

// IsValuePrimitive returns true if the value is int, float64, bool,
// string, duration, date, datetime, localDateTime, or localTime
func IsValuePrimitive(v Value) bool {
	switch v.Get().(type) {
	case int, float64, bool, string, neo4j.Duration, neo4j.Date, neo4j.LocalDateTime, neo4j.LocalTime:
		return true
	}
	return false
}

// ValueAsBool returns the bool value, or if it is not bool, false,false
func ValueAsBool(v Value) (bool, bool) {
	if b, ok := v.Get().(bool); ok {
		return b, true
	}
	return false, false
}

func ValueOf(in interface{}) Value {
	if in == nil {
		return RValue{}
	}
	switch v := in.(type) {
	case Value:
		return v
	case int8:
		return RValue{Value: int(v)}
	case int16:
		return RValue{Value: int(v)}
	case int32:
		return RValue{Value: int(v)}
	case int64:
		return RValue{Value: int(v)}
	case int:
		return RValue{Value: v}
	case uint8:
		return RValue{Value: int(v)}
	case uint16:
		return RValue{Value: int(v)}
	case uint32:
		return RValue{Value: int(v)}
	case string:
		return RValue{Value: v}
	case bool:
		return RValue{Value: v}
	case float64:
		return RValue{Value: v}
	case float32:
		return RValue{Value: float64(v)}
	case neo4j.Duration:
		return RValue{Value: v}
	case neo4j.Date:
		return RValue{Value: v}
	case neo4j.LocalDateTime:
		return RValue{Value: v}
	case neo4j.LocalTime:
		return RValue{Value: v}
	case graph.Node:
		return RValue{Value: v}
	case []graph.Edge:
		return RValue{Value: v}
	case []Value:
		return RValue{Value: v}
	case map[string]Value:
		return RValue{Value: v}
	case graph.StringSet:
		return RValue{Value: v}
	}
	panic(fmt.Sprintf("Invalid value: %v %T", in, in))
}

// IsValueSame compares two values and decides if the two are the same
func IsValueSame(v, v2 Value) bool {
	if IsValuePrimitive(v) {
		if IsValuePrimitive(v2) {
			eq, err := comparePrimitiveValues(v.Get(), v2.Get())
			return err != nil && eq == 0
		}
		return false
	}

	switch val1 := v.Get().(type) {
	case []Value:
		val2, ok := v2.Get().([]Value)
		if !ok {
			return false
		}
		if len(val1) != len(val2) {
			return false
		}
		for i := range val1 {
			if !IsValueSame(val1[i], val2[i]) {
				return false
			}
		}
		return true

	case map[string]Value:
		val2, ok := v2.Get().(map[string]Value)
		if !ok {
			return false
		}
		if len(val1) != len(val2) {
			return false
		}
		for k, v := range val1 {
			v2, ok := val2[k]
			if !ok {
				return false
			}
			if !IsValueSame(v, v2) {
				return false
			}
		}
		return true

	case graph.StringSet:
		val2, ok := v2.Get().(graph.StringSet)
		if !ok {
			return false
		}
		if len(val1) != len(val2) {
			return false
		}
		for k := range val1 {
			if !val2.Has(k) {
				return false
			}
		}
		return true

	case graph.Node:
		val2, ok := v2.Get().(graph.Node)
		if !ok {
			return false
		}
		return val1 == val2

	case []graph.Edge:
		val2, ok := v2.Get().([]graph.Edge)
		if !ok {
			return false
		}
		if len(val1) != len(val2) {
			return false
		}
		for i, x := range val1 {
			if x != val2[i] {
				return false
			}
		}
		return true
	}
	return false
}

func (v RValue) String() string {
	if v.Value == nil {
		return "null"
	}
	if IsValuePrimitive(v) {
		return fmt.Sprint(v.Value)
	}
	switch val := v.Value.(type) {
	case []Value:
		return fmt.Sprint(val)
	case map[string]Value:
		result := make([]string, 0)
		for k, v := range val {
			result = append(result, fmt.Sprintf("%s: %s", k, v))
		}
		return fmt.Sprintf("{%s}", strings.Join(result, " "))
	}
	return fmt.Sprint(v.Value)
}
