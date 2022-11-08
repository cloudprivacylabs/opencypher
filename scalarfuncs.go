package opencypher

import (
	"fmt"
	"strconv"
)

func init() {
	globalFuncs["toInteger"] = Function{
		Name:      "toInteger",
		MinArgs:   1,
		MaxArgs:   1,
		ValueFunc: toIntegerFunc,
	}
	globalFuncs["toIntegerOrNull"] = Function{
		Name:      "toIntegerOrNull",
		MinArgs:   1,
		MaxArgs:   1,
		ValueFunc: toIntegerFunc,
	}
	globalFuncs["toFloatOrNull"] = Function{
		Name:      "toFloatOrNull",
		MinArgs:   1,
		MaxArgs:   1,
		ValueFunc: toFloatOrNullFunc,
	}
	globalFuncs["toFloat"] = Function{
		Name:      "toFloat",
		MinArgs:   1,
		MaxArgs:   1,
		ValueFunc: toFloatFunc,
	}
}

func toIntegerOrNullFunc(ctx *EvalContext, args []Value) (Value, error) {
	v := args[0].Get()
	if v == nil {
		return RValue{}, nil
	}
	if i, ok := v.(int); ok {
		return RValue{Value: i}, nil
	}
	if s, ok := v.(string); ok {
		if i, err := strconv.Atoi(s); err == nil {
			return RValue{Value: i}, nil
		}
		return RValue{}, nil
	}
	if f, ok := v.(float64); ok {
		return RValue{Value: int(f)}, nil
	}
	if b, ok := v.(bool); ok {
		if b {
			return RValue{Value: 1}, nil
		}
		return RValue{Value: 0}, nil
	}
	return RValue{}, nil
}

func toIntegerFunc(ctx *EvalContext, args []Value) (Value, error) {
	v := args[0].Get()
	if v == nil {
		return RValue{}, nil
	}
	if i, ok := v.(int); ok {
		return RValue{Value: i}, nil
	}
	if s, ok := v.(string); ok {
		i, err := strconv.Atoi(s)
		if err == nil {
			return RValue{Value: i}, nil
		}
		return RValue{}, err
	}
	if f, ok := v.(float64); ok {
		return RValue{Value: int(f)}, nil
	}
	if b, ok := v.(bool); ok {
		if b {
			return RValue{Value: 1}, nil
		}
		return RValue{Value: 0}, nil
	}
	return nil, fmt.Errorf("Invalid integer value: %v", v)
}

func toFloatOrNullFunc(ctx *EvalContext, args []Value) (Value, error) {
	v := args[0].Get()
	if v == nil {
		return RValue{}, nil
	}
	if i, ok := v.(int); ok {
		return RValue{Value: float64(i)}, nil
	}
	if s, ok := v.(string); ok {
		if i, err := strconv.ParseFloat(s, 64); err == nil {
			return RValue{Value: i}, nil
		}
		return RValue{}, nil
	}
	if f, ok := v.(float64); ok {
		return RValue{Value: f}, nil
	}
	return RValue{}, nil
}

func toFloatFunc(ctx *EvalContext, args []Value) (Value, error) {
	v := args[0].Get()
	if v == nil {
		return RValue{}, nil
	}
	if i, ok := v.(int); ok {
		return RValue{Value: float64(i)}, nil
	}
	if s, ok := v.(string); ok {
		i, err := strconv.ParseFloat(s, 64)
		if err == nil {
			return RValue{Value: i}, nil
		}
		return RValue{}, err
	}
	if f, ok := v.(float64); ok {
		return RValue{Value: f}, nil
	}
	return RValue{}, fmt.Errorf("Invalid float value: %v", v)
}
