package opencypher

import (
	"math"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func (expr *unaryAddOrSubtractExpression) Evaluate(ctx *EvalContext) (Value, error) {
	if expr.constValue != nil {
		return expr.constValue, nil
	}

	value, err := expr.expr.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	// If the value is  an lvalue, preserve lvalue status
	if !expr.neg {
		return value, nil
	}
	// From now on, it is an rvalue
	ret := RValue{Const: value.IsConst()}
	val := value.Get()
	if val == nil {
		return ret, nil
	}
	if intValue, ok := val.(int); ok {
		ret.Value = -intValue
	} else if floatValue, ok := val.(float64); ok {
		ret.Value = -floatValue
	} else {
		return ret, ErrInvalidUnaryOperation
	}
	if ret.IsConst() {
		expr.constValue = &ret
	}
	return ret, nil
}

func (expr *powerOfExpression) Evaluate(ctx *EvalContext) (Value, error) {
	if expr.constValue != nil {
		return expr.constValue, nil
	}
	val, err := expr.parts[0].Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	if len(expr.parts) == 1 {
		return val, nil
	}
	// ret is an rvalue
	ret := RValue{
		Value: val.Get(),
		Const: val.IsConst(),
	}
	for i := 1; 1 < len(expr.parts); i++ {
		val, err := expr.parts[i].Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		v := val.Get()
		if v == nil {
			return RValue{}, nil
		}
		var valValue float64
		if intValue, ok := v.(int); ok {
			valValue = float64(intValue)
		} else if floatValue, ok := v.(float64); ok {
			valValue = floatValue
		} else {
			return RValue{}, ErrInvalidPowerOperation
		}
		if i, ok := ret.Value.(int); ok {
			ret.Value = math.Pow(float64(i), valValue)
		} else if f, ok := ret.Value.(float64); ok {
			ret.Value = math.Pow(f, valValue)
		} else {
			return RValue{}, ErrInvalidPowerOperation
		}
		ret.Const = ret.Const && val.IsConst()
	}
	if ret.Const {
		expr.constValue = &ret
	}
	return ret, nil
}

func mulintint(a, b int, op rune) (int, error) {
	switch op {
	case '*':
		return a * b, nil
	case '/':
		if b == 0 {
			return 0, ErrDivideByZero
		}
		return a / b, nil
	}
	if b == 0 {
		return 0, ErrDivideByZero
	}
	return a % b, nil
}

func mulintfloat(a int, b float64, op rune) (float64, error) {
	switch op {
	case '*':
		return float64(a) * b, nil
	case '/':
		if b == 0 {
			return 0, ErrDivideByZero
		}
		return float64(a) / b, nil
	}
	if b == 0 {
		return 0, ErrDivideByZero
	}
	return math.Mod(float64(a), b), nil
}

func mulfloatint(a float64, b int, op rune) (float64, error) {
	switch op {
	case '*':
		return a * float64(b), nil
	case '/':
		if b == 0 {
			return 0, ErrDivideByZero
		}
		return a / float64(b), nil
	}
	if b == 0 {
		return 0, ErrDivideByZero
	}
	return math.Mod(a, float64(b)), nil
}

func mulfloatfloat(a, b float64, op rune) (float64, error) {
	switch op {
	case '*':
		return a * b, nil
	case '/':
		if b == 0 {
			return 0, ErrDivideByZero
		}
		return a / b, nil
	}
	if b == 0 {
		return 0, ErrDivideByZero
	}
	return math.Mod(a, b), nil
}

func muldurint(a neo4j.Duration, b int64, op rune) (neo4j.Duration, error) {
	switch op {
	case '*':
		return neo4j.DurationOf(a.Months*b, a.Days*b, a.Seconds*b, a.Nanos*int(b)), nil
	case '/':
		if b == 0 {
			return neo4j.Duration{}, ErrDivideByZero
		}
		return neo4j.DurationOf(a.Months/b, a.Days/b, a.Seconds/b, a.Nanos/int(b)), nil
	}
	return neo4j.Duration{}, ErrInvalidDurationOperation
}

func mulintdur(a int64, b neo4j.Duration, op rune) (neo4j.Duration, error) {
	switch op {
	case '*':
		return neo4j.DurationOf(b.Months*a, b.Days*a, b.Seconds*a, b.Nanos*int(a)), nil
	default:
		return neo4j.Duration{}, ErrInvalidDurationOperation
	}
}

func muldurfloat(a neo4j.Duration, b float64, op rune) (neo4j.Duration, error) {
	val := int64(b)
	switch op {
	case '*':
		return neo4j.DurationOf(int64(a.Months*val), int64(a.Days*val), int64(a.Seconds*val), a.Nanos*int(val)), nil
	case '/':
		if b == 0 {
			return neo4j.Duration{}, ErrDivideByZero
		}
		return neo4j.DurationOf(int64(a.Months/val), int64(a.Days/val), int64(a.Seconds/val), a.Nanos/int(val)), nil
	}
	return neo4j.Duration{}, ErrInvalidDurationOperation
}

func mulfloatdur(a float64, b neo4j.Duration, op rune) (neo4j.Duration, error) {
	val := int64(a)
	switch op {
	case '*':
		return neo4j.DurationOf(b.Months*val, b.Days*val, b.Seconds*val, b.Nanos*int(val)), nil
	default:
		return neo4j.Duration{}, ErrInvalidDurationOperation
	}
}

func (expr *multiplyDivideModuloExpression) Evaluate(ctx *EvalContext) (Value, error) {
	if expr.constValue != nil {
		return expr.constValue, nil
	}
	if len(expr.parts) == 1 {
		v, err := expr.parts[0].expr.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		if v.IsConst() {
			expr.constValue = v
		}
		return v, err
	}
	// Multiple parts, cannot be an lvalue
	var ret RValue
	var err error
	for i := range expr.parts {
		var val Value
		val, err = expr.parts[i].expr.Evaluate(ctx)
		if err != nil {
			return val, err
		}
		if i == 0 {
			ret.Value = val.Get()
			ret.Const = val.IsConst()
		} else {
			if ret.Value == nil {
				return RValue{}, nil
			}
			ret.Const = ret.Const && val.IsConst()
			switch result := ret.Value.(type) {
			case int:
				switch operand := val.Get().(type) {
				case int:
					ret.Value, err = mulintint(result, operand, expr.parts[i].op)
				case float64:
					ret.Value, err = mulintfloat(result, operand, expr.parts[i].op)
				case neo4j.Duration:
					ret.Value, err = mulintdur(int64(result), operand, expr.parts[i].op)
				default:
					err = ErrInvalidMultiplicativeOperation
				}
			case float64:
				switch operand := val.Get().(type) {
				case int:
					ret.Value, err = mulfloatint(result, operand, expr.parts[i].op)
				case float64:
					ret.Value, err = mulfloatfloat(result, operand, expr.parts[i].op)
				case neo4j.Duration:
					ret.Value, err = mulfloatdur(result, operand, expr.parts[i].op)
				default:
					err = ErrInvalidMultiplicativeOperation
				}
			case neo4j.Duration:
				switch operand := val.Get().(type) {
				case int:
					ret.Value, err = muldurint(result, int64(operand), expr.parts[i].op)
				case float64:
					ret.Value, err = muldurfloat(result, operand, expr.parts[i].op)
				default:
					err = ErrInvalidMultiplicativeOperation
				}
			default:
				err = ErrInvalidMultiplicativeOperation
			}
		}
	}
	if err != nil {
		return nil, err
	}
	if ret.Const {
		expr.constValue = &ret
	}
	return ret, nil
}

func addintint(a int, b int, sub bool) int {
	if sub {
		return a - b
	}
	return a + b
}

func addintfloat(a int, b float64, sub bool) float64 {
	if sub {
		return float64(a) - b
	}
	return float64(a) + b
}

func addfloatint(a float64, b int, sub bool) float64 {
	if sub {
		return a - float64(b)
	}
	return a + float64(b)
}

func addfloatfloat(a float64, b float64, sub bool) float64 {
	if sub {
		return a - b
	}
	return a + b
}

func addstringstring(a string, b string, sub bool) (string, error) {
	if sub {
		return "", ErrInvalidStringOperation
	}
	return a + b, nil
}

func adddatedur(a neo4j.Date, b neo4j.Duration, sub bool) neo4j.Date {
	t := a.Time()
	if sub {
		return neo4j.DateOf(time.Date(t.Year(), t.Month()-time.Month(b.Months), t.Day()-int(b.Days), 0, 0, 0, 0, t.Location()))
	}
	return neo4j.DateOf(time.Date(t.Year(), t.Month()+time.Month(b.Months), t.Day()+int(b.Days), 0, 0, 0, 0, t.Location()))
}

func addtimedur(a neo4j.LocalTime, b neo4j.Duration, sub bool) neo4j.LocalTime {
	t := a.Time()
	if sub {
		return neo4j.LocalTimeOf(time.Date(1970, 1, 1, t.Hour(), t.Minute(), t.Second()-int(b.Seconds), t.Nanosecond()-b.Nanos, t.Location()))
	}
	return neo4j.LocalTimeOf(time.Date(1970, 1, 1, t.Hour(), t.Minute(), t.Second()+int(b.Seconds), t.Nanosecond()+b.Nanos, t.Location()))
}

func adddatetimedur(a neo4j.LocalDateTime, b neo4j.Duration, sub bool) neo4j.LocalDateTime {
	t := a.Time()
	if sub {
		return neo4j.LocalDateTimeOf(time.Date(t.Year(), t.Month()-time.Month(b.Months), t.Day()-int(b.Days), t.Hour(), t.Minute(), t.Second()-int(b.Seconds), t.Nanosecond()-b.Nanos, t.Location()))
	}
	return neo4j.LocalDateTimeOf(time.Date(t.Year(), t.Month()+time.Month(b.Months), t.Day()+int(b.Days), t.Hour(), t.Minute(), t.Second()+int(b.Seconds), t.Nanosecond()+b.Nanos, t.Location()))
}

func adddurdate(a neo4j.Duration, b neo4j.Date, sub bool) (neo4j.Date, error) {
	if sub {
		return neo4j.Date{}, ErrInvalidDateOperation
	}
	return adddatedur(b, a, false), nil
}

func adddurtime(a neo4j.Duration, b neo4j.LocalTime, sub bool) (neo4j.LocalTime, error) {
	if sub {
		return neo4j.LocalTime{}, ErrInvalidDateOperation
	}
	return addtimedur(b, a, false), nil
}

func adddurdatetime(a neo4j.Duration, b neo4j.LocalDateTime, sub bool) (neo4j.LocalDateTime, error) {
	if sub {
		return neo4j.LocalDateTime{}, ErrInvalidDateOperation
	}
	return adddatetimedur(b, a, false), nil
}

func adddurdur(a neo4j.Duration, b neo4j.Duration, sub bool) (neo4j.Duration, error) {
	if sub {
		return neo4j.DurationOf(a.Months-b.Months, a.Days-b.Days, a.Seconds-b.Seconds, a.Nanos-b.Nanos), nil
	}
	return neo4j.DurationOf(a.Months+b.Months, a.Days+b.Days, a.Seconds+b.Seconds, a.Nanos+b.Nanos), nil
}

func addlistlist(a, b []Value) RValue {
	arr := make([]Value, 0, len(a)+len(b))
	ret := RValue{Const: true}
	for _, x := range a {
		if !x.IsConst() {
			ret.Const = false
		}
		arr = append(arr, x)
	}
	for _, x := range b {
		if !x.IsConst() {
			ret.Const = false
		}
		arr = append(arr, x)
	}
	return ret
}

func (expr *addOrSubtractExpression) Evaluate(ctx *EvalContext) (Value, error) {
	if expr.constValue != nil {
		return expr.constValue, nil
	}
	if len(expr.add) == 1 && len(expr.sub) == 0 {
		ret, err := expr.add[0].Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		if ret.IsConst() {
			expr.constValue = ret
		}
		return ret, nil
	}

	var ret RValue
	first := true

	accumulate := func(operand Value, sub bool) error {
		if first {
			first = false
			ret.Value = operand.Get()
			ret.Const = operand.IsConst()
			return nil
		}
		if ret.Value == nil || operand.Get() == nil {
			ret.Value = nil
			ret.Const = false
			return nil
		}
		ret.Const = ret.Const && operand.IsConst()
		var err error
		switch retValue := ret.Value.(type) {
		case int:
			switch operandValue := operand.Get().(type) {
			case int:
				ret.Value = addintint(retValue, operandValue, sub)
			case float64:
				ret.Value = addintfloat(retValue, operandValue, sub)
			default:
				err = ErrInvalidAdditiveOperation
			}
		case float64:
			switch operandValue := operand.Get().(type) {
			case int:
				ret.Value = addfloatint(retValue, operandValue, sub)
			case float64:
				ret.Value = addfloatfloat(retValue, operandValue, sub)
			default:
				err = ErrInvalidAdditiveOperation
			}
		case string:
			switch operandValue := operand.Get().(type) {
			case string:
				ret.Value, err = addstringstring(retValue, operandValue, sub)
			default:
				err = ErrInvalidAdditiveOperation
			}
		case neo4j.Duration:
			switch operandValue := operand.Get().(type) {
			case neo4j.Duration:
				ret.Value, err = adddurdur(retValue, operandValue, sub)
			case neo4j.Date:
				ret.Value, err = adddurdate(retValue, operandValue, sub)
			case neo4j.LocalTime:
				ret.Value, err = adddurtime(retValue, operandValue, sub)
			case neo4j.LocalDateTime:
				ret.Value, err = adddurdatetime(retValue, operandValue, sub)
			default:
				err = ErrInvalidAdditiveOperation
			}
		case neo4j.Date:
			switch operandValue := operand.Get().(type) {
			case neo4j.Duration:
				ret.Value = adddatedur(retValue, operandValue, sub)
			default:
				err = ErrInvalidAdditiveOperation
			}
		case neo4j.LocalTime:
			switch operandValue := operand.Get().(type) {
			case neo4j.Duration:
				ret.Value = addtimedur(retValue, operandValue, sub)
			default:
				err = ErrInvalidAdditiveOperation
			}
		case neo4j.LocalDateTime:
			switch operandValue := operand.Get().(type) {
			case neo4j.Duration:
				ret.Value = adddatetimedur(retValue, operandValue, sub)
			default:
				err = ErrInvalidAdditiveOperation
			}
		case []Value:
			if sub {
				return ErrInvalidAdditiveOperation
			}
			switch operandValue := operand.Get().(type) {
			case []Value:
				ret = addlistlist(retValue, operandValue)
			default:
				err = ErrInvalidAdditiveOperation
			}
		}
		return err
	}

	for i := range expr.add {
		val, err := expr.add[i].Evaluate(ctx)
		if err != nil {
			return RValue{}, err
		}
		if err = accumulate(val, false); err != nil {
			return RValue{}, err
		}
	}
	for i := range expr.sub {
		val, err := expr.sub[i].Evaluate(ctx)
		if err != nil {
			return RValue{}, err
		}
		if err = accumulate(val, true); err != nil {
			return RValue{}, err
		}
	}
	if ret.Const {
		expr.constValue = &ret
	}
	return ret, nil
}
