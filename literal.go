package opencypher

import (
	"strings"
)

func (literal IntLiteral) Evaluate(ctx *EvalContext) (Value, error) {
	return RValue{
		Value: int(literal),
		Const: true,
	}, nil
}

func (literal BooleanLiteral) Evaluate(ctx *EvalContext) (Value, error) {
	return RValue{
		Value: bool(literal),
		Const: true,
	}, nil
}

func (literal DoubleLiteral) Evaluate(ctx *EvalContext) (Value, error) {
	return RValue{
		Value: float64(literal),
		Const: true,
	}, nil
}

func (literal StringLiteral) Evaluate(ctx *EvalContext) (Value, error) {
	return RValue{
		Value: string(literal),
		Const: true,
	}, nil
}

func (literal NullLiteral) Evaluate(ctx *EvalContext) (Value, error) {
	return RValue{
		Const: true,
	}, nil
}

func (lst *ListLiteral) Evaluate(ctx *EvalContext) (Value, error) {
	if lst.constValue != nil {
		return lst.constValue, nil
	}
	ret := make([]Value, 0, len(lst.Values))
	var val RValue
	for i := range lst.Values {
		v, err := lst.Values[i].Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		if i == 0 {
			val.Const = v.IsConst()
		} else {
			val.Const = val.Const && v.IsConst()
		}
		ret = append(ret, v)
	}
	val.Value = ret
	if val.IsConst() {
		lst.constValue = val
	}
	return val, nil
}

func (mp *MapLiteral) Evaluate(ctx *EvalContext) (Value, error) {
	if mp.constValue != nil {
		return mp.constValue, nil
	}
	var val RValue
	ret := make(map[string]Value)
	for i := range mp.KeyValues {
		keyStr := mp.KeyValues[i].Key
		if len(keyStr) == 0 {
			return nil, ErrInvalidMapKey
		}
		value, err := mp.KeyValues[i].Value.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		ret[keyStr] = value
		if i == 0 {
			val.Const = value.IsConst()
		} else {
			val.Const = val.Const && value.IsConst()
		}
	}
	val.Value = ret
	if val.IsConst() {
		mp.constValue = val
	}
	return val, nil
}

func (r *RangeLiteral) Evaluate(ctx *EvalContext) (from, to *int, err error) {
	var v Value
	if r.From != nil {
		v, err = r.From.Evaluate(ctx)
		if err != nil {
			return
		}
		i := v.Get().(int)
		from = &i
	}
	if r.To != nil {
		v, err = r.To.Evaluate(ctx)
		if err != nil {
			return
		}
		i := v.Get().(int)
		to = &i
	}
	return
}

// EscapeLabelLiteral escape a literal that can be used as a label. It
// returns `s`
func EscapeLabelLiteral(s string) string {
	return "`" + s + "`"
}

// EscapePropertyKeyLiteral escapes a literal that can be used as a
// property key. Returns `s`
func EscapePropertyKeyLiteral(s string) string {
	return "`" + s + "`"
}

// EscapeStringLiteral returns "s" where backslashes and quotes in s
// are escaped
func EscapeStringLiteral(s string) string {
	bld := strings.Builder{}
	bld.WriteRune('"')
	for _, c := range s {
		if c == '\\' {
			bld.WriteRune('\\')
			bld.WriteRune('\\')
		} else if c == '"' {
			bld.WriteRune('\\')
			bld.WriteRune('"')
		} else {
			bld.WriteRune(c)
		}
	}
	bld.WriteRune('"')
	return bld.String()
}
