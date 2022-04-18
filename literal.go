package opencypher

import ()

func (literal IntLiteral) Evaluate(ctx *EvalContext) (Value, error) {
	return Value{
		Value:    int(literal),
		Constant: true,
	}, nil
}

func (literal BooleanLiteral) Evaluate(ctx *EvalContext) (Value, error) {
	return Value{
		Value:    bool(literal),
		Constant: true,
	}, nil
}

func (literal DoubleLiteral) Evaluate(ctx *EvalContext) (Value, error) {
	return Value{
		Value:    float64(literal),
		Constant: true,
	}, nil
}

func (literal StringLiteral) Evaluate(ctx *EvalContext) (Value, error) {
	return Value{
		Value:    string(literal),
		Constant: true,
	}, nil
}

func (literal NullLiteral) Evaluate(ctx *EvalContext) (Value, error) {
	return Value{
		Constant: true,
	}, nil
}

func (lst *ListLiteral) Evaluate(ctx *EvalContext) (Value, error) {
	if lst.constValue != nil {
		return *lst.constValue, nil
	}
	ret := make([]Value, 0, len(lst.Values))
	var val Value
	for i := range lst.Values {
		v, err := lst.Values[i].Evaluate(ctx)
		if err != nil {
			return Value{}, err
		}
		if i == 0 {
			val.Constant = v.Constant
		} else {
			val.Constant = val.Constant && v.Constant
		}
		ret = append(ret, v)
	}
	val.Value = ret
	if val.Constant {
		lst.constValue = &val
	}
	return val, nil
}

func (mp *MapLiteral) Evaluate(ctx *EvalContext) (Value, error) {
	if mp.constValue != nil {
		return *mp.constValue, nil
	}
	var val Value
	ret := make(map[string]Value)
	for i := range mp.KeyValues {
		keyStr := mp.KeyValues[i].Key
		if len(keyStr) == 0 {
			return Value{}, ErrInvalidMapKey
		}
		value, err := mp.KeyValues[i].Value.Evaluate(ctx)
		if err != nil {
			return Value{}, err
		}
		ret[keyStr] = value
		if i == 0 {
			val.Constant = value.Constant
		} else {
			val.Constant = val.Constant && value.Constant
		}
	}
	val.Value = ret
	if val.Constant {
		mp.constValue = &val
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
		i := v.Value.(int)
		from = &i
	}
	if r.To != nil {
		v, err = r.To.Evaluate(ctx)
		if err != nil {
			return
		}
		i := v.Value.(int)
		to = &i
	}
	return
}
