package opencypher

import (
	"math"
	"time"

	"github.com/neo4j/neo4j-go-driver/neo4j"
)

func (expr *UnaryAddOrSubtractExpression) Evaluate(ctx *EvalContext) (Value, error) {
	if expr.constValue != nil {
		return *expr.constValue, nil
	}

	value, err := expr.Expr.Evaluate(ctx)
	if err != nil {
		return value, err
	}
	if value.Value == nil {
		return value, nil
	}
	if expr.Neg {
		if intValue, ok := value.Value.(int); ok {
			value.Value = -intValue
		} else if floatValue, ok := value.Value.(float64); ok {
			value.Value = -floatValue
		} else {
			return value, ErrInvalidUnaryOperation
		}
	}
	if value.Constant {
		expr.constValue = &value
	}
	return value, nil
}

func (expr *PowerOfExpression) Evaluate(ctx *EvalContext) (Value, error) {
	if expr.constValue != nil {
		return *expr.constValue, nil
	}
	var ret Value
	for i := range expr.Parts {
		val, err := expr.Parts[i].Evaluate(ctx)
		if err != nil {
			return val, err
		}
		if val.Value == nil {
			return Value{}, nil
		}
		if i == 0 {
			ret = val
		} else {
			var valValue float64
			if intValue, ok := val.Value.(int); ok {
				valValue = float64(intValue)
			} else if floatValue, ok := val.Value.(float64); ok {
				valValue = floatValue
			} else {
				return Value{}, ErrInvalidPowerOperation
			}
			if i, ok := ret.Value.(int); ok {
				ret.Value = math.Pow(float64(i), valValue)
			} else if f, ok := ret.Value.(float64); ok {
				ret.Value = math.Pow(f, valValue)
			} else {
				return Value{}, ErrInvalidPowerOperation
			}
			ret.Constant = ret.Constant && val.Constant
		}
	}
	if ret.Constant {
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
		return neo4j.DurationOf(a.Months()*b, a.Days()*b, a.Seconds()*b, a.Nanos()*int(b)), nil
	case '/':
		if b == 0 {
			return neo4j.Duration{}, ErrDivideByZero
		}
		return neo4j.DurationOf(a.Months()/b, a.Days()/b, a.Seconds()/b, a.Nanos()/int(b)), nil
	}
	return neo4j.Duration{}, ErrInvalidDurationOperation
}

func mulintdur(a int64, b neo4j.Duration, op rune) (neo4j.Duration, error) {
	switch op {
	case '*':
		return neo4j.DurationOf(b.Months()*a, b.Days()*a, b.Seconds()*a, b.Nanos()*int(a)), nil
	default:
		return neo4j.Duration{}, ErrInvalidDurationOperation
	}
}

func muldurfloat(a neo4j.Duration, b float64, op rune) (neo4j.Duration, error) {
	val := int64(b)
	switch op {
	case '*':
		return neo4j.DurationOf(int64(a.Months()*val), int64(a.Days()*val), int64(a.Seconds()*val), a.Nanos()*int(val)), nil
	case '/':
		if b == 0 {
			return neo4j.Duration{}, ErrDivideByZero
		}
		return neo4j.DurationOf(int64(a.Months()/val), int64(a.Days()/val), int64(a.Seconds()/val), a.Nanos()/int(val)), nil
	}
	return neo4j.Duration{}, ErrInvalidDurationOperation
}

func mulfloatdur(a float64, b neo4j.Duration, op rune) (neo4j.Duration, error) {
	val := int64(a)
	switch op {
	case '*':
		return neo4j.DurationOf(b.Months()*val, b.Days()*val, b.Seconds()*val, b.Nanos()*int(val)), nil
	default:
		return neo4j.Duration{}, ErrInvalidDurationOperation
	}
}

func (expr *MultiplyDivideModuloExpression) Evaluate(ctx *EvalContext) (Value, error) {
	if expr.constValue != nil {
		return *expr.constValue, nil
	}

	var ret Value
	var err error
	for i := range expr.Parts {
		var val Value
		val, err = expr.Parts[i].Expr.Evaluate(ctx)
		if err != nil {
			return val, err
		}
		if i == 0 {
			ret = val
		} else {
			if ret.Value == nil {
				return Value{}, nil
			}
			ret.Constant = ret.Constant && val.Constant
			switch result := ret.Value.(type) {
			case int:
				switch operand := val.Value.(type) {
				case int:
					ret.Value, err = mulintint(result, operand, expr.Parts[i].Op)
				case float64:
					ret.Value, err = mulintfloat(result, operand, expr.Parts[i].Op)
				case neo4j.Duration:
					ret.Value, err = mulintdur(int64(result), operand, expr.Parts[i].Op)
				default:
					err = ErrInvalidMultiplicativeOperation
				}
			case float64:
				switch operand := val.Value.(type) {
				case int:
					ret.Value, err = mulfloatint(result, operand, expr.Parts[i].Op)
				case float64:
					ret.Value, err = mulfloatfloat(result, operand, expr.Parts[i].Op)
				case neo4j.Duration:
					ret.Value, err = mulfloatdur(result, operand, expr.Parts[i].Op)
				default:
					err = ErrInvalidMultiplicativeOperation
				}
			case neo4j.Duration:
				switch operand := val.Value.(type) {
				case int:
					ret.Value, err = muldurint(result, int64(operand), expr.Parts[i].Op)
				case float64:
					ret.Value, err = muldurfloat(result, operand, expr.Parts[i].Op)
				default:
					err = ErrInvalidMultiplicativeOperation
				}
			default:
				err = ErrInvalidMultiplicativeOperation
			}
		}
	}
	if err != nil {
		return Value{}, err
	}
	if ret.Constant {
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
		return neo4j.DateOf(time.Date(t.Year(), t.Month()-time.Month(b.Months()), t.Day()-int(b.Days()), 0, 0, 0, 0, t.Location()))
	}
	return neo4j.DateOf(time.Date(t.Year(), t.Month()+time.Month(b.Months()), t.Day()+int(b.Days()), 0, 0, 0, 0, t.Location()))
}

func addtimedur(a neo4j.LocalTime, b neo4j.Duration, sub bool) neo4j.LocalTime {
	t := a.Time()
	if sub {
		return neo4j.LocalTimeOf(time.Date(1970, 1, 1, t.Hour(), t.Minute(), t.Second()-int(b.Seconds()), t.Nanosecond()-b.Nanos(), t.Location()))
	}
	return neo4j.LocalTimeOf(time.Date(1970, 1, 1, t.Hour(), t.Minute(), t.Second()+int(b.Seconds()), t.Nanosecond()+b.Nanos(), t.Location()))
}

func adddatetimedur(a neo4j.LocalDateTime, b neo4j.Duration, sub bool) neo4j.LocalDateTime {
	t := a.Time()
	if sub {
		return neo4j.LocalDateTimeOf(time.Date(t.Year(), t.Month()-time.Month(b.Months()), t.Day()-int(b.Days()), t.Hour(), t.Minute(), t.Second()-int(b.Seconds()), t.Nanosecond()-b.Nanos(), t.Location()))
	}
	return neo4j.LocalDateTimeOf(time.Date(t.Year(), t.Month()+time.Month(b.Months()), t.Day()+int(b.Days()), t.Hour(), t.Minute(), t.Second()+int(b.Seconds()), t.Nanosecond()+b.Nanos(), t.Location()))
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
		return neo4j.DurationOf(a.Months()-b.Months(), a.Days()-b.Days(), a.Seconds()-b.Seconds(), a.Nanos()-b.Nanos()), nil
	}
	return neo4j.DurationOf(a.Months()+b.Months(), a.Days()+b.Days(), a.Seconds()+b.Seconds(), a.Nanos()+b.Nanos()), nil
}

func addlistlist(a, b []Value) Value {
	arr := make([]Value, 0, len(a)+len(b))
	ret := Value{Constant: true}
	for _, x := range a {
		if !x.Constant {
			ret.Constant = false
		}
		arr = append(arr, x)
	}
	for _, x := range b {
		if !x.Constant {
			ret.Constant = false
		}
		arr = append(arr, x)
	}
	return ret
}

func (expr *AddOrSubtractExpression) Evaluate(ctx *EvalContext) (Value, error) {
	if expr.constValue != nil {
		return *expr.constValue, nil
	}

	var ret Value
	first := true

	accumulate := func(operand Value, sub bool) error {
		if first {
			first = false
			ret = operand
			return nil
		}
		ret.Constant = ret.Constant && operand.Constant
		var err error
		switch retValue := ret.Value.(type) {
		case int:
			switch operandValue := operand.Value.(type) {
			case int:
				ret.Value = addintint(retValue, operandValue, sub)
			case float64:
				ret.Value = addintfloat(retValue, operandValue, sub)
			default:
				err = ErrInvalidAdditiveOperation
			}
		case float64:
			switch operandValue := operand.Value.(type) {
			case int:
				ret.Value = addfloatint(retValue, operandValue, sub)
			case float64:
				ret.Value = addfloatfloat(retValue, operandValue, sub)
			default:
				err = ErrInvalidAdditiveOperation
			}
		case string:
			switch operandValue := operand.Value.(type) {
			case string:
				ret.Value, err = addstringstring(retValue, operandValue, sub)
			default:
				err = ErrInvalidAdditiveOperation
			}
		case neo4j.Duration:
			switch operandValue := operand.Value.(type) {
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
			switch operandValue := operand.Value.(type) {
			case neo4j.Duration:
				ret.Value = adddatedur(retValue, operandValue, sub)
			default:
				err = ErrInvalidAdditiveOperation
			}
		case neo4j.LocalTime:
			switch operandValue := operand.Value.(type) {
			case neo4j.Duration:
				ret.Value = addtimedur(retValue, operandValue, sub)
			default:
				err = ErrInvalidAdditiveOperation
			}
		case neo4j.LocalDateTime:
			switch operandValue := operand.Value.(type) {
			case neo4j.Duration:
				ret.Value = adddatetimedur(retValue, operandValue, sub)
			default:
				err = ErrInvalidAdditiveOperation
			}
		case []Value:
			if sub {
				return ErrInvalidAdditiveOperation
			}
			switch operandValue := operand.Value.(type) {
			case []Value:
				ret = addlistlist(retValue, operandValue)
			default:
				err = ErrInvalidAdditiveOperation
			}
		}
		return err
	}

	for i := range expr.Add {
		val, err := expr.Add[i].Evaluate(ctx)
		if err != nil {
			return Value{}, err
		}
		if err = accumulate(val, false); err != nil {
			return Value{}, err
		}
	}
	for i := range expr.Sub {
		val, err := expr.Add[i].Evaluate(ctx)
		if err != nil {
			return Value{}, err
		}
		if err = accumulate(val, true); err != nil {
			return Value{}, err
		}
	}
	if ret.Constant {
		expr.constValue = &ret
	}
	return ret, nil
}
