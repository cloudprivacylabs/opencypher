package opencypher

import (
	"fmt"
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
	globalFuncs["print"] = Function{
		Name:      "print",
		MinArgs:   0,
		MaxArgs:   -1,
		ValueFunc: printFunc,
	}
}

func splitFunc(ctx *EvalContext, args []Value) (Value, error) {
	if args[0].Get() == nil {
		return RValue{}, nil
	}
	str, err := ValueAsString(args[0])
	if err != nil {
		return nil, fmt.Errorf("In split: %w", err)
	}
	if args[1].Get() == nil {
		return RValue{}, nil
	}
	delim, err := ValueAsString(args[1])
	if err != nil {
		return nil, fmt.Errorf("In string: %w", err)
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
		return nil, fmt.Errorf("In trim: %w", err)
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
		return nil, fmt.Errorf("In right: %s", err)
	}
	i, err := ValueAsInt(args[1])
	if err != nil {
		return nil, fmt.Errorf("In right: %s", err)
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
		return nil, fmt.Errorf("In left: %w, %v", err, args[0])
	}
	i, err := ValueAsInt(args[1])
	if err != nil {
		return nil, fmt.Errorf("In left: %w", err)
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
		return nil, fmt.Errorf("In substring: %w", err)
	}
	start, err := ValueAsInt(args[1])
	if err != nil {
		return nil, fmt.Errorf("In substring: %w", err)
	}
	max := -1
	if len(args) == 3 {
		if args[2].Get() == nil {
			return RValue{}, nil
		}
		max, err = ValueAsInt(args[2])
		if err != nil {
			return nil, fmt.Errorf("In substring: %w", err)
		}
		if max < 0 {
			return RValue{}, nil
		}
	}
	if start < 0 || start > len(s) {
		return RValue{Value: ""}, nil
	}
	end := start
	if max == -1 {
		end = len(s)
	} else {
		end = start + max
	}

	if end > len(s) {
		end = len(s)
	}
	return RValue{Value: s[start:end]}, nil
}

func printFunc(ctx *EvalContext, args []Value) (Value, error) {
	for _, arg := range args {
		fmt.Print(arg.Get())
		fmt.Print(" ")
	}
	fmt.Println()
	return RValue{}, nil
}
