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
	ErrNotAnLValue                    = errors.New("Not and lvalue")
)

type ErrInvalidIndirection string

func (e ErrInvalidIndirection) Error() string { return "Invalid indirection: " + string(e) }

func (expr Parameter) Evaluate(ctx *EvalContext) (Value, error) {
	return ctx.GetParameter(string(expr))
}

type ErrInvalidAssignment string

func (e ErrInvalidAssignment) Error() string { return "Invalid assignment to: " + string(e) }

func (expr StringListNullOperatorExpression) Evaluate(ctx *EvalContext) (Value, error) {
	val, err := expr.PropertyOrLabels.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	for _, part := range expr.Parts {
		val, err = part.evaluate(ctx, val)
		if err != nil {
			return nil, err
		}
	}
	return val, nil
}

func (expr StringListNullOperatorExpressionPart) evaluate(ctx *EvalContext, inputValue Value) (Value, error) {
	switch {
	case expr.IsNull != nil:
		if *expr.IsNull {
			return RValue{Value: inputValue.Get() == nil}, nil
		}
		return RValue{Value: inputValue.Get() != nil}, nil

	case expr.ListIndex != nil:
		listValue, ok := inputValue.Get().([]Value)
		if !ok {
			if inputValue.Get() != nil {
				return nil, ErrNotAList
			}
		}
		indexValue, err := expr.ListIndex.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		if indexValue.Get() == nil {
			return RValue{}, nil
		}
		intValue, ok := indexValue.Get().(int)
		if !ok {
			return nil, ErrInvalidListIndex
		}
		if listValue == nil {
			return RValue{}, nil
		}
		if intValue >= 0 {
			if intValue >= len(listValue) {
				return RValue{}, nil
			}
			return listValue[intValue], nil
		}
		index := len(listValue) + intValue
		if index < 0 {
			return RValue{}, nil
		}
		return listValue[index], nil

	case expr.ListIn != nil:
		listValue, err := expr.ListIn.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		list, ok := listValue.Get().([]Value)
		if ok {
			if listValue.Get() != nil {
				return nil, ErrNotAList
			}
		}
		if inputValue.Get() == nil {
			return RValue{}, nil
		}
		hasNull := false
		for _, elem := range list {
			if elem.Get() == nil {
				hasNull = true
			} else {
				v, err := comparePrimitiveValues(inputValue.Get(), elem.Get())
				if err != nil {
					return nil, err
				}
				if v == 0 {
					return RValue{Value: true}, nil
				}
			}
		}
		if hasNull {
			return RValue{}, nil
		}
		return RValue{Value: false}, nil

	case expr.ListRange != nil:
		constant := inputValue.IsConst()
		listValue, ok := inputValue.Get().([]Value)
		if !ok {
			if inputValue.Get() != nil {
				return nil, ErrNotAList
			}
		}
		from, err := expr.ListRange.First.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		if from.Get() == nil {
			return RValue{}, nil
		}
		if !from.IsConst() {
			constant = false
		}
		fromi, ok := from.Get().(int)
		if !ok {
			return nil, ErrInvalidListIndex
		}
		to, err := expr.ListRange.Second.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		if to.Get() == nil {
			return RValue{}, nil
		}
		if !to.IsConst() {
			constant = false
		}
		toi, ok := to.Get().(int)
		if !ok {
			return nil, ErrInvalidListIndex
		}
		if fromi < 0 || toi < 0 {
			return nil, ErrInvalidListIndex
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
			if !listValue[i].IsConst() {
				constant = false
			}
			arr = append(arr, listValue[i])
		}
		return RValue{Value: arr, Const: constant}, nil
	}
	return expr.String.evaluate(ctx, inputValue)
}

func (expr StringOperatorExpression) evaluate(ctx *EvalContext, inputValue Value) (Value, error) {
	inputStrValue, ok := inputValue.Get().(string)
	if !ok {
		return nil, ErrInvalidStringOperation
	}
	exprValue, err := expr.Expr.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	strValue, ok := exprValue.Get().(string)
	if !ok {
		return nil, ErrInvalidStringOperation
	}
	if expr.Operator == "STARTS" {
		return RValue{Value: strings.HasPrefix(inputStrValue, strValue)}, nil
	}
	if expr.Operator == "ENDS" {
		return RValue{Value: strings.HasSuffix(inputStrValue, strValue)}, nil
	}
	return RValue{Value: strings.Contains(inputStrValue, strValue)}, nil
}

func (pl PropertyOrLabelsExpression) Evaluate(ctx *EvalContext) (Value, error) {
	v, err := pl.Atom.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	val := RValue{Value: v.Get()}
	if pl.NodeLabels != nil {
		gobj, ok := val.Value.(graph.StringSet)
		if !ok {
			return nil, ErrNotAStringSet
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
			return RValue{}, nil
		}
		wp, ok := val.Value.(withProperty)
		if !ok {
			if edges, ed := val.Value.([]graph.Edge); ed {
				if len(edges) == 1 {
					wp = edges[0]
					ok = true
				}
			}
		}
		if ok {
			prop, ok := wp.GetProperty(property.String())
			if !ok {
				return RValue{}, nil
			}
			if n, ok := prop.(withNativeValue); ok {
				val = ValueOf(n.GetNativeValue()).(RValue)
			} else {
				val = ValueOf(prop).(RValue)
			}
		} else {
			return nil, ErrValueDoesNotHaveProperties
		}
	}
	return val, nil
}

func (f *FunctionInvocation) Evaluate(ctx *EvalContext) (Value, error) {
	if f.function == nil {
		fn, err := ctx.GetFunction(f.Name)
		if err != nil {
			return nil, err
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
				return nil, err
			}
			if a == 0 {
				isConst = v.IsConst()
			} else if !v.IsConst() {
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
			return nil, err
		}
		testValue = v
	}
	for _, alternative := range cs.Alternatives {
		when, err := alternative.When.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		if cs.Test != nil {
			result, err := comparePrimitiveValues(testValue, when)
			if err != nil {
				return nil, err
			}
			if result == 0 {
				return alternative.Then.Evaluate(ctx)
			}
		} else {
			boolValue, ok := when.Get().(bool)
			if !ok {
				return nil, ErrNotABooleanExpression
			}
			if boolValue {
				return alternative.Then.Evaluate(ctx)
			}
		}
	}
	if cs.Default != nil {
		return cs.Default.Evaluate(ctx)
	}
	return RValue{}, nil
}

func (v Variable) Evaluate(ctx *EvalContext) (Value, error) {
	val, err := ctx.GetVar(string(v))
	if err != nil {
		return nil, err
	}
	// A variable is always an lvalue
	return NewLValue(val), nil
}

// Evaluate a regular query, which is a single query with an optional
// union list
func (query RegularQuery) Evaluate(ctx *EvalContext) (Value, error) {
	result, err := query.SingleQuery.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	resultSet, ok := result.Get().(ResultSet)
	if !ok {
		return nil, ErrExpectingResultSet
	}
	for _, u := range query.Unions {
		newResult, err := u.SingleQuery.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		newResultSet, ok := newResult.Get().(ResultSet)
		if !ok {
			return nil, ErrExpectingResultSet
		}
		if err := resultSet.Union(newResultSet, u.All); err != nil {
			return nil, err
		}
	}
	return RValue{Value: resultSet}, nil
}

func (query singlePartQuery) Evaluate(ctx *EvalContext) (Value, error) {
	ret := ResultSet{}
	project := func(rows []map[string]Value) error {
		for _, item := range rows {
			val, err := query.Return.Projection.Items.Project(ctx, item)
			if err != nil {
				return err
			}
			ret.Rows = append(ret.Rows, val)
		}
		return nil
	}
	if len(query.Read) > 0 {
		results := ResultSet{}
		for _, r := range query.Read {
			rs, err := r.GetResults(ctx)
			if err != nil {
				return nil, err
			}
			results.Add(rs)
		}

		for _, upd := range query.Update {
			_, err := upd.Update(ctx, results)
			if err != nil {
				return nil, err
			}
		}
		if query.Return == nil {
			return RValue{Value: ResultSet{}}, nil
		}
		err := project(results.Rows)
		if err != nil {
			return nil, err
		}
		return RValue{Value: ret}, nil
	}

	for _, upd := range query.Update {
		if err := upd.TopLevelUpdate(ctx); err != nil {
			return nil, err
		}
	}
	if query.Return == nil {
		return RValue{Value: ResultSet{}}, nil
	}
	val, err := query.Return.Projection.Items.Project(ctx, nil)
	if err != nil {
		return nil, err
	}
	ret.Rows = append(ret.Rows, val)
	return RValue{Value: ret}, nil
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

func (pe propertyExpression) Evaluate(ctx *EvalContext) (Value, error) {
	val, err := pe.Atom.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	for i := range pe.Lookup {
		prop := pe.Lookup[i].String()
		// val must be a node, or map[string]Value
		if val == nil {
			return nil, nil
		}
		value := val.Get()
		switch parent := value.(type) {
		case graph.Node:
			val = LValue{
				getter: func() interface{} {
					v, _ := parent.GetProperty(prop)
					return v
				},
				setter: func(v interface{}) {
					if v == nil {
						parent.RemoveProperty(prop)
					} else {
						parent.SetProperty(prop, v)
					}
				},
			}
		case map[string]Value:
			val = LValue{
				getter: func() interface{} {
					v := parent[prop]
					if v == nil {
						return nil
					}
					return v.Get()
				},
				setter: func(v interface{}) {
					if v == nil {
						delete(parent, prop)
					} else {
						parent[prop] = ValueOf(v)
					}
				},
			}
		default:
			return nil, ErrInvalidIndirection(prop)
		}
	}
	return val, nil
}

func (unwind unwind) GetResults(ctx *EvalContext) (ResultSet, error)      { panic("Unimplemented") }
func (ls ListComprehension) Evaluate(ctx *EvalContext) (Value, error)     { panic("Unimplemented") }
func (p PatternComprehension) Evaluate(ctx *EvalContext) (Value, error)   { panic("Unimplemented") }
func (flt FilterAtom) Evaluate(ctx *EvalContext) (Value, error)           { panic("Unimplemented") }
func (rel RelationshipsPattern) Evaluate(ctx *EvalContext) (Value, error) { panic("Unimplemented") }
func (cnt CountAtom) Evaluate(ctx *EvalContext) (Value, error)            { panic("Unimplemented") }
func (mq multiPartQuery) Evaluate(ctx *EvalContext) (Value, error)        { panic("Unimplemented") }
