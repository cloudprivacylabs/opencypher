package opencypher

import (
	"errors"
	"strconv"
	"strings"

	"github.com/cloudprivacylabs/opencypher/graph"
)

const (
	c_const  uint = 0x001
	c_lvalue uint = 0x002
)

var (
	ErrDivideByZero                   = errors.New("Divide by zero")
	ErrInvalidUnaryOperation          = errors.New("Invalid unary operation")
	ErrInvalidPowerOperation          = errors.New("Invalid power operation")
	ErrInvalidMultiplicativeOperation = errors.New("Invalid multiplicative operation")
	ErrInvalidDurationOperation       = errors.New("Invalid duration operation")
	ErrOperationWithNull              = errors.New("Operation with null")
	ErrInvalidStringOperation         = errors.New("Invalid string operation")
	ErrInvalidDateOperation           = errors.New("Invalid date operation")
	ErrInvalidAdditiveOperation       = errors.New("Invalid additive operation")
	ErrInvalidComparison              = errors.New("Invalid comparison")
	ErrInvalidListIndex               = errors.New("Invalid list index")
	ErrNotAList                       = errors.New("Not a list")
	ErrNotABooleanExpression          = errors.New("Not a boolean expression")
	ErrMapKeyNotString                = errors.New("Map key is not a string")
	ErrInvalidMapKey                  = errors.New("Invalid map key")
	ErrNotAStringSet                  = errors.New("Not a string set")
	ErrIntValueRequired               = errors.New("Int value required")
	ErrExpectingResultSet             = errors.New("Expecting a result set")
	ErrPropertiesParameterExpected    = errors.New("Parameter value cannot be used for properties")
	ErrPropertiesExpected             = errors.New("Value cannot be used for properties")
	ErrValueDoesNotHaveProperties     = errors.New("Value does not have properties")
)

func (expr Parameter) Evaluate(ctx *EvalContext) (Value, error) {
	return ctx.GetParameter(string(expr))
}

func (expr StringListNullOperatorExpression) Evaluate(ctx *EvalContext) (Value, error) {
	val, err := expr.PropertyOrLabels.Evaluate(ctx)
	if err != nil {
		return Value{}, err
	}
	for _, part := range expr.Parts {
		val, err = part.evaluate(ctx, val)
		if err != nil {
			return Value{}, err
		}
	}
	return val, nil
}

func (expr StringListNullOperatorExpressionPart) evaluate(ctx *EvalContext, inputValue Value) (Value, error) {
	switch {
	case expr.IsNull != nil:
		if *expr.IsNull {
			return Value{Value: inputValue.Value == nil}, nil
		}
		return Value{Value: inputValue.Value != nil}, nil

	case expr.ListIndex != nil:
		listValue, ok := inputValue.Value.([]Value)
		if !ok {
			if inputValue.Value != nil {
				return Value{}, ErrNotAList
			}
		}
		indexValue, err := expr.ListIndex.Evaluate(ctx)
		if err != nil {
			return Value{}, err
		}
		if indexValue.Value == nil {
			return Value{}, nil
		}
		intValue, ok := indexValue.Value.(int)
		if !ok {
			return Value{}, ErrInvalidListIndex
		}
		if listValue == nil {
			return Value{}, nil
		}
		if intValue >= 0 {
			if intValue >= len(listValue) {
				return Value{}, nil
			}
			return listValue[intValue], nil
		}
		index := len(listValue) + intValue
		if index < 0 {
			return Value{}, nil
		}
		return listValue[index], nil

	case expr.ListIn != nil:
		listValue, err := expr.ListIn.Evaluate(ctx)
		if err != nil {
			return Value{}, err
		}
		list, ok := listValue.Value.([]Value)
		if ok {
			if listValue.Value != nil {
				return Value{}, ErrNotAList
			}
		}
		if inputValue.Value == nil {
			return Value{}, nil
		}
		hasNull := false
		for _, elem := range list {
			if elem.Value == nil {
				hasNull = true
			} else {
				v, err := comparePrimitiveValues(inputValue.Value, elem.Value)
				if err != nil {
					return Value{}, err
				}
				if v == 0 {
					return Value{Value: true}, nil
				}
			}
		}
		if hasNull {
			return Value{}, nil
		}
		return Value{Value: false}, nil

	case expr.ListRange != nil:
		constant := inputValue.Constant
		listValue, ok := inputValue.Value.([]Value)
		if !ok {
			if inputValue.Value != nil {
				return Value{}, ErrNotAList
			}
		}
		from, err := expr.ListRange.First.Evaluate(ctx)
		if err != nil {
			return Value{}, err
		}
		if from.Value == nil {
			return Value{}, nil
		}
		if !from.Constant {
			constant = false
		}
		fromi, ok := from.Value.(int)
		if !ok {
			return Value{}, ErrInvalidListIndex
		}
		to, err := expr.ListRange.Second.Evaluate(ctx)
		if err != nil {
			return Value{}, err
		}
		if to.Value == nil {
			return Value{}, nil
		}
		if !to.Constant {
			constant = false
		}
		toi, ok := to.Value.(int)
		if !ok {
			return Value{}, ErrInvalidListIndex
		}
		if fromi < 0 || toi < 0 {
			return Value{}, ErrInvalidListIndex
		}
		if fromi >= len(listValue) {
			fromi = len(listValue) - 1
		}
		if toi >= len(listValue) {
			toi = len(listValue) - 1
		}
		if fromi > toi {
			fromi = toi
		}
		arr := make([]Value, 0, toi-fromi)
		for i := fromi; i < toi; i++ {
			if !listValue[i].Constant {
				constant = false
			}
			arr = append(arr, listValue[i])
		}
		return Value{Value: arr, Constant: constant}, nil
	}
	return expr.String.evaluate(ctx, inputValue)
}

func (expr StringOperatorExpression) evaluate(ctx *EvalContext, inputValue Value) (Value, error) {
	inputStrValue, ok := inputValue.Value.(string)
	if !ok {
		return Value{}, ErrInvalidStringOperation
	}
	exprValue, err := expr.Expr.Evaluate(ctx)
	if err != nil {
		return Value{}, err
	}
	strValue, ok := exprValue.Value.(string)
	if !ok {
		return Value{}, ErrInvalidStringOperation
	}
	if expr.Operator == "STARTS" {
		return Value{Value: strings.HasPrefix(inputStrValue, strValue)}, nil
	}
	if expr.Operator == "ENDS" {
		return Value{Value: strings.HasSuffix(inputStrValue, strValue)}, nil
	}
	return Value{Value: strings.Contains(inputStrValue, strValue)}, nil
}

func (pl PropertyOrLabelsExpression) Evaluate(ctx *EvalContext) (Value, error) {
	val, err := pl.Atom.Evaluate(ctx)
	if err != nil {
		return Value{}, err
	}
	if pl.NodeLabels != nil {
		gobj, ok := val.Value.(graph.StringSet)
		if !ok {
			return Value{}, ErrNotAStringSet
		}
		for _, label := range *pl.NodeLabels {
			str := label.String()
			gobj.Add(str)
		}
		val.Value = gobj
	}
	type withProperty interface {
		GetProperty(string) (interface{}, bool)
	}
	type withNativeValue interface {
		GetNativeValue() interface{}
	}
	for _, property := range pl.PropertyLookup {
		if val.Value == nil {
			return Value{}, nil
		}
		if wp, ok := val.Value.(withProperty); ok {
			prop, ok := wp.GetProperty(property.String())
			if !ok {
				return Value{}, nil
			}
			if n, ok := prop.(withNativeValue); ok {
				val = ValueOf(n.GetNativeValue())
			} else {
				val = ValueOf(prop)
			}
		} else {
			return Value{}, ErrValueDoesNotHaveProperties
		}
	}
	return val, nil
}

func (f *FunctionInvocation) Evaluate(ctx *EvalContext) (Value, error) {
	if f.function == nil {
		fn, err := ctx.GetFunction(f.Name)
		if err != nil {
			return Value{}, err
		}
		f.function = fn
	}
	args := f.args
	if args == nil {
		args = make([]Evaluatable, 0, len(f.Args))
		isConst := false

		for a := range f.Args {
			v, err := f.Args[a].Evaluate(ctx)
			if err != nil {
				return Value{}, err
			}
			if a == 0 {
				isConst = v.Constant
			} else if !v.Constant {
				isConst = false
			}
			args = append(args, v)
		}
		if isConst {
			f.args = args
		}
	}
	return f.function(ctx, args)
}

func (cs Case) Evaluate(ctx *EvalContext) (Value, error) {
	var testValue Value
	if cs.Test != nil {
		v, err := cs.Test.Evaluate(ctx)
		if err != nil {
			return Value{}, err
		}
		testValue = v
	}
	for _, alternative := range cs.Alternatives {
		when, err := alternative.When.Evaluate(ctx)
		if err != nil {
			return Value{}, err
		}
		if cs.Test != nil {
			result, err := comparePrimitiveValues(testValue, when)
			if err != nil {
				return Value{}, err
			}
			if result == 0 {
				return alternative.Then.Evaluate(ctx)
			}
		} else {
			boolValue, ok := when.Value.(bool)
			if !ok {
				return Value{}, ErrNotABooleanExpression
			}
			if boolValue {
				return alternative.Then.Evaluate(ctx)
			}
		}
	}
	if cs.Default != nil {
		return cs.Default.Evaluate(ctx)
	}
	return Value{}, nil
}

func (v Variable) Evaluate(ctx *EvalContext) (Value, error) {
	return ctx.GetVar(string(v))
}

// Evaluate a regular query, which is a single query with an optional
// union list
func (query RegularQuery) Evaluate(ctx *EvalContext) (Value, error) {
	result, err := query.SingleQuery.Evaluate(ctx)
	if err != nil {
		return Value{}, err
	}
	resultSet, ok := result.Value.(ResultSet)
	if !ok {
		return Value{}, ErrExpectingResultSet
	}
	for _, u := range query.Unions {
		newResult, err := u.SingleQuery.Evaluate(ctx)
		if err != nil {
			return Value{}, err
		}
		newResultSet, ok := newResult.Value.(ResultSet)
		if !ok {
			return Value{}, ErrExpectingResultSet
		}
		if err := resultSet.Union(newResultSet, u.All); err != nil {
			return Value{}, err
		}
	}
	return Value{Value: resultSet}, nil
}

func (query SinglePartQuery) Evaluate(ctx *EvalContext) (Value, error) {
	if len(query.Update) > 0 {
		panic("Updating Query is not implemented")
	}

	ret := ResultSet{}
	if len(query.Read) > 0 {
		results := ResultSet{}
		for _, r := range query.Read {
			rs, err := r.GetResults(ctx)
			if err != nil {
				return Value{}, err
			}
			results.Add(rs)
		}
		if query.Return == nil {
			return Value{}, nil
		}
		// Keys keep the order of each key in the result set
		for _, item := range results.Rows {
			val, err := query.Return.Projection.Items.Project(ctx, item)
			if err != nil {
				return Value{}, err
			}
			ret.Rows = append(ret.Rows, val)
		}
		return Value{Value: ret}, nil
	}
	if query.Return != nil {
		val, err := query.Return.Projection.Items.Project(ctx, nil)
		if err != nil {
			return Value{}, err
		}
		ret.Rows = append(ret.Rows, val)
		return Value{Value: ret}, nil
	}
	return Value{}, nil
}

func (prj ProjectionItems) Project(ctx *EvalContext, values map[string]Value) (map[string]Value, error) {
	ret := make(map[string]Value)
	if prj.All {
		for k, v := range values {
			ret[k] = v
		}
		return ret, nil
	}

	for k, v := range values {
		ctx.SetVar(k, v)
	}
	for i, item := range prj.Items {
		result, err := item.Expr.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		var varName string
		if item.Var != nil {
			varName = string(*item.Var)
		} else {
			varName = strconv.Itoa(i + 1)
		}
		ret[varName] = result
	}
	return ret, nil
}

func (unwind Unwind) GetResults(ctx *EvalContext) (ResultSet, error)      { panic("Unimplemented") }
func (ls ListComprehension) Evaluate(ctx *EvalContext) (Value, error)     { panic("Unimplemented") }
func (p PatternComprehension) Evaluate(ctx *EvalContext) (Value, error)   { panic("Unimplemented") }
func (flt FilterAtom) Evaluate(ctx *EvalContext) (Value, error)           { panic("Unimplemented") }
func (rel RelationshipsPattern) Evaluate(ctx *EvalContext) (Value, error) { panic("Unimplemented") }
func (cnt CountAtom) Evaluate(ctx *EvalContext) (Value, error)            { panic("Unimplemented") }
func (mq MultiPartQuery) Evaluate(ctx *EvalContext) (Value, error)        { panic("Unimplemented") }
