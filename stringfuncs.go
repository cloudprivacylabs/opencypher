package opencypher

import (
	"strings"
	"unicode"
)

func init() {
	globalFuncs["left"] = Function{
		Name:      "left",
		MinArgs:   2,
		MaxArgs:   2,
		ValueFunc: leftFunc,
	}
	globalFuncs["right"] = Function{
		Name:      "right",
		MinArgs:   2,
		MaxArgs:   2,
		ValueFunc: rightFunc,
	}
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
	globalFuncs["substring"] = Function{
		Name:      "substring",
		MinArgs:   2,
		MaxArgs:   3,
		ValueFunc: substringFunc,
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

func rightFunc(ctx *EvalContext, args []Value) (Value, error) {
	if args[0].Get() == nil || args[1].Get() == nil {
		return RValue{}, nil
	}
	s, err := ValueAsString(args[0])
	if err != nil {
		return nil, err
	}
	i, err := ValueAsInt(args[1])
	if err != nil {
		return nil, err
	}
	if i < 0 || i > len(s) {
		return RValue{Value: ""}, nil
	}
	return RValue{Value: s[i:]}, nil
}

func leftFunc(ctx *EvalContext, args []Value) (Value, error) {
	if args[0].Get() == nil || args[1].Get() == nil {
		return RValue{}, nil
	}
	s, err := ValueAsString(args[0])
	if err != nil {
		return nil, err
	}
	i, err := ValueAsInt(args[1])
	if err != nil {
		return nil, err
	}
	if i < 0 || i > len(s) {
		return RValue{Value: ""}, nil
	}
	return RValue{Value: s[:i]}, nil
}

func substringFunc(ctx *EvalContext, args []Value) (Value, error) {
	if args[0].Get() == nil || args[1].Get() == nil {
		return RValue{}, nil
	}
	s, err := ValueAsString(args[0])
	if err != nil {
		return nil, err
	}
	start, err := ValueAsInt(args[1])
	if err != nil {
		return nil, err
	}
	max := -1
	if len(args) == 3 {
		if args[2].Get() == nil {
			return RValue{}, nil
		}
		max, err = ValueAsInt(args[2])
		if err != nil {
			return nil, err
		}
	}
	if start < 0 || start > len(s) {
		return RValue{Value: ""}, nil
	}
	if max > len(s) || max == -1 {
		max = len(s)
	}
	return RValue{Value: s[start:max]}, nil
}
