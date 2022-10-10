package opencypher

import (
	"strings"
	"unicode"
)

func init() {
	globalFuncs["split"] = Function{
		Name:      "split",
		MinArgs:   2,
		MaxArgs:   2,
		ValueFunc: splitFunc,
	}
	globalFuncs["trim"] = Function{
		Name:      "trim",
		MinArgs:   1,
		MaxArgs:   1,
		ValueFunc: trimFunc,
	}
	globalFuncs["ltrim"] = Function{
		Name:      "ltrim",
		MinArgs:   1,
		MaxArgs:   1,
		ValueFunc: ltrimFunc,
	}
	globalFuncs["rtrim"] = Function{
		Name:      "rtrim",
		MinArgs:   1,
		MaxArgs:   1,
		ValueFunc: rtrimFunc,
	}
}

func splitFunc(ctx *EvalContext, args []Value) (Value, error) {
	if args[0].Get() == nil {
		return RValue{}, nil
	}
	str, err := ValueAsString(args[0])
	if err != nil {
		return nil, err
	}
	if args[0].Get() == nil {
		return RValue{}, nil
	}
	delim, err := ValueAsString(args[1])
	if err != nil {
		return nil, err
	}
	strs := strings.Split(str, delim)
	out := make([]Value, 0, len(strs))
	for _, x := range strs {
		out = append(out, RValue{Value: x})
	}
	return RValue{Value: out}, nil
}

func trimFunc(ctx *EvalContext, args []Value) (Value, error) {
	if args[0].Get() == nil {
		return RValue{}, nil
	}
	str, err := ValueAsString(args[0])
	if err != nil {
		return nil, err
	}
	return RValue{Value: strings.TrimSpace(str)}, nil
}

func ltrimFunc(ctx *EvalContext, args []Value) (Value, error) {
	if args[0].Get() == nil {
		return RValue{}, nil
	}
	str, err := ValueAsString(args[0])
	if err != nil {
		return nil, err
	}
	return RValue{Value: strings.TrimLeftFunc(str, unicode.IsSpace)}, nil
}

func rtrimFunc(ctx *EvalContext, args []Value) (Value, error) {
	if args[0].Get() == nil {
		return RValue{}, nil
	}
	str, err := ValueAsString(args[0])
	if err != nil {
		return nil, err
	}
	return RValue{Value: strings.TrimRightFunc(str, unicode.IsSpace)}, nil
}
