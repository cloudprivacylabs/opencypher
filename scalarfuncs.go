package opencypher

import (
	"strconv"
)

func init() {
	globalFuncs["toInteger"] = Function{
		Name:      "toInteger",
		MinArgs:   1,
		MaxArgs:   1,
		ValueFunc: toIntegerFunc,
	}
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
