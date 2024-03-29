package opencypher

import ()

func comparePrimitiveValues(v1, v2 interface{}) (int, error) {
	if v1 == nil {
		if v2 == nil {
			return 0, nil
		}
		return -1, nil
	}
	if v2 == nil {
		return 1, nil
	}
	switch value1 := v1.(type) {
	case bool:
		switch value2 := v2.(type) {
		case bool:
			if value1 == value2 {
				return 0, nil
			}
			if value1 {
				return 1, nil
			}
			return -1, nil
		}
	case int:
		switch value2 := v2.(type) {
		case int:
			return value1 - value2, nil
		case float64:
			if float64(value1) == value2 {
				return 0, nil
			}
			if float64(value1) < value2 {
				return -1, nil
			}
			return 1, nil
		}
	case float64:
		switch value2 := v2.(type) {
		case int:
			if value1 == float64(value2) {
				return 0, nil
			}
			if value1 < float64(value2) {
				return -1, nil
			}
			return 1, nil
		case float64:
			if value1 == value2 {
				return 0, nil
			}
			if value1 < value2 {
				return -1, nil
			}
			return 1, nil
		}
	case string:
		if str, ok := v2.(string); ok {
			if value1 == str {
				return 0, nil
			}
			if value1 < str {
				return -1, nil
			}
			return 1, nil
		}
	case Duration:
		if dur, ok := v2.(Duration); ok {
			if value1.Days == dur.Days && value1.Months == dur.Months && value1.Seconds == dur.Seconds && value1.Nanos == dur.Nanos {
				return 0, nil
			}
			if value1.Days < dur.Days {
				return -1, nil
			}
			if value1.Months < dur.Months {
				return -1, nil
			}
			if value1.Seconds < dur.Seconds {
				return -1, nil
			}
			if value1.Nanos < dur.Nanos {
				return -1, nil
			}
			return 1, nil
		}
	case Date:
		if date, ok := v2.(Date); ok {
			t1 := value1.Time()
			t2 := date.Time()
			if t1.Equal(t2) {
				return 0, nil
			}
			if t1.Before(t2) {
				return -1, nil
			}
			return 0, nil
		}
	case LocalTime:
		if date, ok := v2.(LocalTime); ok {
			t1 := value1.Time()
			t2 := date.Time()
			if t1.Equal(t2) {
				return 0, nil
			}
			if t1.Before(t2) {
				return -1, nil
			}
			return 0, nil
		}
	case LocalDateTime:
		if date, ok := v2.(LocalDateTime); ok {
			t1 := value1.Time()
			t2 := date.Time()
			if t1.Equal(t2) {
				return 0, nil
			}
			if t1.Before(t2) {
				return -1, nil
			}
			return 0, nil
		}
	}
	return 0, ErrInvalidComparison

}

func (expr comparisonExpression) Evaluate(ctx *EvalContext) (Value, error) {
	val, err := expr.first.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	if val.Get() == nil {
		return RValue{}, nil
	}
	ret := RValue{Value: val.Get(), Const: val.IsConst()}
	for i := range expr.second {
		second, err := expr.second[i].expr.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		if second.Get() == nil {
			return RValue{}, nil
		}
		result, err := comparePrimitiveValues(val.Get(), second.Get())
		if err != nil {
			return nil, err
		}
		switch expr.second[i].op {
		case "=":
			ret.Value = result == 0
		case "<>":
			ret.Value = result != 0
		case "<":
			ret.Value = result < 0
		case "<=":
			ret.Value = result <= 0
		case ">":
			ret.Value = result > 0
		case ">=":
			ret.Value = result >= 0
		}
		ret.Const = ret.Const && second.IsConst()
	}
	return ret, nil
}

func (expr notExpression) Evaluate(ctx *EvalContext) (Value, error) {
	val, err := expr.part.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	if val.Get() == nil {
		return RValue{}, nil
	}
	value, ok := val.Get().(bool)
	if !ok {
		return nil, ErrNotABooleanExpression
	}
	return RValue{Value: !value, Const: val.IsConst()}, nil
}

func (expr andExpression) Evaluate(ctx *EvalContext) (Value, error) {
	var ret RValue
	for i := range expr.parts {
		val, err := expr.parts[i].Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		if val.Get() == nil {
			return RValue{}, nil
		}
		if i == 0 {
			ret = RValue{Value: val.Get(), Const: val.IsConst()}
		} else {
			bval, ok := ret.Value.(bool)
			if !ok {
				return nil, ErrNotABooleanExpression
			}
			vval, ok := val.Get().(bool)
			if !ok {
				return nil, ErrNotABooleanExpression
			}
			ret.Const = ret.Const && val.IsConst()
			ret.Value = bval && vval
			if !bval || !vval {
				break
			}
		}
	}
	return ret, nil
}

func (expr xorExpression) Evaluate(ctx *EvalContext) (Value, error) {
	var ret RValue
	for i := range expr.parts {
		val, err := expr.parts[i].Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		if val.Get() == nil {
			return RValue{}, nil
		}
		if i == 0 {
			ret = RValue{Value: val.Get(), Const: val.IsConst()}
		} else {
			bval, ok := ret.Value.(bool)
			if !ok {
				return nil, ErrNotABooleanExpression
			}
			vval, ok := val.Get().(bool)
			if !ok {
				return nil, ErrNotABooleanExpression
			}
			ret.Const = ret.Const && val.IsConst()
			ret.Value = bval != vval
		}
	}
	return ret, nil
}

func (expr orExpression) Evaluate(ctx *EvalContext) (Value, error) {
	var ret RValue
	for i := range expr.parts {
		val, err := expr.parts[i].Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		if val.Get() == nil {
			return RValue{}, nil
		}
		if i == 0 {
			ret = RValue{Value: val.Get(), Const: val.IsConst()}
		} else {
			bval, ok := ret.Value.(bool)
			if !ok {
				return nil, ErrNotABooleanExpression
			}
			vval, ok := val.Get().(bool)
			if !ok {
				return nil, ErrNotABooleanExpression
			}
			ret.Const = ret.Const && val.IsConst()
			ret.Value = bval || vval
			if bval || vval {
				break
			}
		}
	}
	return ret, nil
}
