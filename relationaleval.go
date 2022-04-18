package opencypher

import (
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

func comparePrimitiveValues(v1, v2 interface{}) (int, error) {
	if v1 == nil || v2 == nil {
		return 0, ErrOperationWithNull
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
	case neo4j.Duration:
		if dur, ok := v2.(neo4j.Duration); ok {
			if value1.Days() == dur.Days() && value1.Months() == dur.Months() && value1.Seconds() == dur.Seconds() && value1.Nanos() == dur.Nanos() {
				return 0, nil
			}
			if value1.Days() < dur.Days() {
				return -1, nil
			}
			if value1.Months() < dur.Months() {
				return -1, nil
			}
			if value1.Seconds() < dur.Seconds() {
				return -1, nil
			}
			if value1.Nanos() < dur.Nanos() {
				return -1, nil
			}
			return 1, nil
		}
	case neo4j.Date:
		if date, ok := v2.(neo4j.Date); ok {
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
	case neo4j.LocalTime:
		if date, ok := v2.(neo4j.LocalTime); ok {
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
	case neo4j.LocalDateTime:
		if date, ok := v2.(neo4j.LocalDateTime); ok {
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

func (expr ComparisonExpression) Evaluate(ctx *EvalContext) (Value, error) {
	val, err := expr.First.Evaluate(ctx)
	if err != nil {
		return Value{}, err
	}

	if val.Value == nil {
		return Value{}, nil
	}
	for i := range expr.Second {
		second, err := expr.Second[i].Expr.Evaluate(ctx)
		if err != nil {
			return Value{}, err
		}
		if second.Value == nil {
			return Value{}, nil
		}
		result, err := comparePrimitiveValues(val.Value, second.Value)
		if err != nil {
			return Value{}, err
		}
		switch expr.Second[i].Op {
		case "=":
			val.Value = result == 0
		case "<>":
			val.Value = result != 0
		case "<":
			val.Value = result < 0
		case "<=":
			val.Value = result <= 0
		case ">":
			val.Value = result > 0
		case ">=":
			val.Value = result >= 0
		}
		val.Constant = val.Constant && second.Constant
	}
	return val, nil
}

func (expr NotExpression) Evaluate(ctx *EvalContext) (Value, error) {
	val, err := expr.Part.Evaluate(ctx)
	if err != nil {
		return Value{}, err
	}
	if val.Value == nil {
		return Value{}, nil
	}
	value, ok := val.Value.(bool)
	if !ok {
		return Value{}, ErrNotABooleanExpression
	}
	val.Value = !value
	return val, nil
}

func (expr AndExpression) Evaluate(ctx *EvalContext) (Value, error) {
	var ret Value
	for i := range expr.Parts {
		val, err := expr.Parts[i].Evaluate(ctx)
		if err != nil {
			return Value{}, err
		}
		if val.Value == nil {
			return Value{}, nil
		}
		if i == 0 {
			ret = val
		} else {
			bval, ok := ret.Value.(bool)
			if !ok {
				return Value{}, ErrNotABooleanExpression
			}
			vval, ok := val.Value.(bool)
			if !ok {
				return Value{}, ErrNotABooleanExpression
			}
			ret.Constant = ret.Constant && val.Constant
			ret.Value = bval && vval
			if !bval || !vval {
				break
			}
		}
	}
	return ret, nil
}

func (expr XorExpression) Evaluate(ctx *EvalContext) (Value, error) {
	var ret Value
	for i := range expr.Parts {
		val, err := expr.Parts[i].Evaluate(ctx)
		if err != nil {
			return Value{}, err
		}
		if val.Value == nil {
			return Value{}, nil
		}
		if i == 0 {
			ret = val
		} else {
			bval, ok := ret.Value.(bool)
			if !ok {
				return Value{}, ErrNotABooleanExpression
			}
			vval, ok := val.Value.(bool)
			if !ok {
				return Value{}, ErrNotABooleanExpression
			}
			ret.Constant = ret.Constant && val.Constant
			ret.Value = bval != vval
		}
	}
	return ret, nil
}

func (expr OrExpression) Evaluate(ctx *EvalContext) (Value, error) {
	var ret Value
	for i := range expr.Parts {
		val, err := expr.Parts[i].Evaluate(ctx)
		if err != nil {
			return Value{}, err
		}
		if val.Value == nil {
			return Value{}, nil
		}
		if i == 0 {
			ret = val
		} else {
			bval, ok := ret.Value.(bool)
			if !ok {
				return Value{}, ErrNotABooleanExpression
			}
			vval, ok := val.Value.(bool)
			if !ok {
				return Value{}, ErrNotABooleanExpression
			}
			ret.Constant = ret.Constant && val.Constant
			ret.Value = bval || vval
			if bval || vval {
				break
			}
		}
	}
	return ret, nil
}
