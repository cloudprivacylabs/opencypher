package opencypher

import (
	"fmt"
	"strings"

	"github.com/neo4j/neo4j-go-driver/neo4j"

	"github.com/cloudprivacylabs/opencypher/graph"
)

// Value represents a computer value. Possible data types it can contain are:
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
type Value struct {
	Value    interface{}
	Constant bool
}

// IsPrimitive returns true if the value is int, float64, bool,
// string, duration, date, datetime, localDateTime, or localTime
func (v Value) IsPrimitive() bool {
	switch v.Value.(type) {
	case int, float64, bool, string, neo4j.Duration, neo4j.Date, neo4j.LocalDateTime, neo4j.LocalTime:
		return true
	}
	return false
}

// AsBool returns the bool value, or if it is not bool, false,false
func (v Value) AsBool() (bool, bool) {
	if b, ok := v.Value.(bool); ok {
		return b, true
	}
	return false, false
}

func ValueOf(in interface{}) Value {
	switch v := in.(type) {
	case Value:
		return v
	case int8:
		return Value{Value: int(v)}
	case int16:
		return Value{Value: int(v)}
	case int32:
		return Value{Value: int(v)}
	case int64:
		return Value{Value: int(v)}
	case int:
		return Value{Value: v}
	case uint8:
		return Value{Value: int(v)}
	case uint16:
		return Value{Value: int(v)}
	case uint32:
		return Value{Value: int(v)}
	case string:
		return Value{Value: v}
	case bool:
		return Value{Value: v}
	case float64:
		return Value{Value: v}
	case float32:
		return Value{Value: float64(v)}
	case neo4j.Duration:
		return Value{Value: v}
	case neo4j.Date:
		return Value{Value: v}
	case neo4j.LocalDateTime:
		return Value{Value: v}
	case neo4j.LocalTime:
		return Value{Value: v}
	case graph.Node:
		return Value{Value: v}
	case []graph.Edge:
		return Value{Value: v}
	case []Value:
		return Value{Value: v}
	case map[string]Value:
		return Value{Value: v}
	case graph.StringSet:
		return Value{Value: v}
	}
	panic(fmt.Sprintf("Invalid value: %v %T", in, in))
}

// IsSame compares two values and decides if the two are the same
func (v Value) IsSame(v2 Value) bool {
	if v.IsPrimitive() {
		if v2.IsPrimitive() {
			eq, err := comparePrimitiveValues(v.Value, v2.Value)
			return err != nil && eq == 0
		}
		return false
	}

	switch val1 := v.Value.(type) {
	case []Value:
		val2, ok := v2.Value.([]Value)
		if !ok {
			return false
		}
		if len(val1) != len(val2) {
			return false
		}
		for i := range val1 {
			if !val1[i].IsSame(val2[i]) {
				return false
			}
		}
		return true

	case map[string]Value:
		val2, ok := v2.Value.(map[string]Value)
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
			if !v.IsSame(v2) {
				return false
			}
		}
		return true

	case graph.StringSet:
		val2, ok := v2.Value.(graph.StringSet)
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
		val2, ok := v2.Value.(graph.Node)
		if !ok {
			return false
		}
		return val1 == val2

	case []graph.Edge:
		val2, ok := v2.Value.([]graph.Edge)
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

func (v Value) Evaluate(ctx *EvalContext) (Value, error) { return v, nil }

func (v Value) String() string {
	if v.Value == nil {
		return "null"
	}
	if v.IsPrimitive() {
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
