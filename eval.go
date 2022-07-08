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

func (expr stringListNullOperatorExpression) Evaluate(ctx *EvalContext) (Value, error) {
	val, err := expr.propertyOrLabels.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	for _, part := range expr.parts {
		val, err = part.evaluate(ctx, val)
		if err != nil {
			return nil, err
		}
	}
	return val, nil
}

func (expr stringListNullOperatorExpressionPart) evaluate(ctx *EvalContext, inputValue Value) (Value, error) {
	switch {
	case expr.isNull != nil:
		if *expr.isNull {
			return RValue{Value: inputValue.Get() == nil}, nil
		}
		return RValue{Value: inputValue.Get() != nil}, nil

	case expr.listIndex != nil:
		listValue, ok := inputValue.Get().([]Value)
		if !ok {
			if inputValue.Get() != nil {
				return nil, ErrNotAList
			}
		}
		indexValue, err := expr.listIndex.Evaluate(ctx)
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

	case expr.listIn != nil:
		listValue, err := expr.listIn.Evaluate(ctx)
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

	case expr.listRange != nil:
		constant := inputValue.IsConst()
		listValue, ok := inputValue.Get().([]Value)
		if !ok {
			if inputValue.Get() != nil {
				return nil, ErrNotAList
			}
		}
		from, err := expr.listRange.first.Evaluate(ctx)
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
		to, err := expr.listRange.second.Evaluate(ctx)
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
	return expr.stringOp.evaluate(ctx, inputValue)
}

func (expr stringOperatorExpression) evaluate(ctx *EvalContext, inputValue Value) (Value, error) {
	inputStrValue, ok := inputValue.Get().(string)
	if !ok {
		return nil, ErrInvalidStringOperation
	}
	exprValue, err := expr.expr.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	strValue, ok := exprValue.Get().(string)
	if !ok {
		return nil, ErrInvalidStringOperation
	}
	if expr.operator == "STARTS" {
		return RValue{Value: strings.HasPrefix(inputStrValue, strValue)}, nil
	}
	if expr.operator == "ENDS" {
		return RValue{Value: strings.HasSuffix(inputStrValue, strValue)}, nil
	}
	return RValue{Value: strings.Contains(inputStrValue, strValue)}, nil
}

func (pl propertyOrLabelsExpression) Evaluate(ctx *EvalContext) (Value, error) {
	v, err := pl.atom.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	val := RValue{Value: v.Get()}
	if pl.nodeLabels != nil {
		gobj, ok := val.Value.(graph.StringSet)
		if !ok {
			return nil, ErrNotAStringSet
		}
		for _, label := range *pl.nodeLabels {
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
	for _, property := range pl.propertyLookup {
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

func (f *functionInvocation) Evaluate(ctx *EvalContext) (Value, error) {
	if f.function == nil {
		fname := make([]string, 0, len(f.name))
		for _, x := range f.name {
			fname = append(fname, string(x))
		}
		fn, err := ctx.GetFunction(fname)
		if err != nil {
			return nil, err
		}
		f.function = fn
	}
	args := f.parsedArgs
	if args == nil {
		args = make([]Evaluatable, 0, len(f.args))
		isConst := false

		for a := range f.args {
			v, err := f.args[a].Evaluate(ctx)
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
			f.parsedArgs = args
		}
	}
	return f.function(ctx, args)
}

func (cs caseClause) Evaluate(ctx *EvalContext) (Value, error) {
	var testValue Value
	if cs.test != nil {
		v, err := cs.test.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		testValue = v
	}
	for _, alternative := range cs.alternatives {
		when, err := alternative.when.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		if cs.test != nil {
			result, err := comparePrimitiveValues(testValue, when)
			if err != nil {
				return nil, err
			}
			if result == 0 {
				return alternative.then.Evaluate(ctx)
			}
		} else {
			boolValue, ok := when.Get().(bool)
			if !ok {
				return nil, ErrNotABooleanExpression
			}
			if boolValue {
				return alternative.then.Evaluate(ctx)
			}
		}
	}
	if cs.def != nil {
		return cs.def.Evaluate(ctx)
	}
	return RValue{}, nil
}

func (v variable) Evaluate(ctx *EvalContext) (Value, error) {
	val, err := ctx.GetVar(string(v))
	if err != nil {
		return nil, err
	}
	// A variable is always an lvalue
	return NewLValue(val), nil
}

// Evaluate a regular query, which is a single query with an optional
// union list
func (query regularQuery) Evaluate(ctx *EvalContext) (Value, error) {
	contexts := make([]*EvalContext, 0)

	lastContext := ctx.SubContext()
	contexts = append(contexts, lastContext)
	result, err := query.singleQuery.Evaluate(lastContext)
	if err != nil {
		return nil, err
	}
	resultSet, ok := result.Get().(ResultSet)
	if !ok {
		return nil, ErrExpectingResultSet
	}
	for _, u := range query.unions {
		lastContext = ctx.SubContext()
		contexts = append(contexts, lastContext)
		newResult, err := u.singleQuery.Evaluate(lastContext)
		if err != nil {
			return nil, err
		}
		newResultSet, ok := newResult.Get().(ResultSet)
		if !ok {
			return nil, ErrExpectingResultSet
		}
		if err := resultSet.Union(newResultSet, u.all); err != nil {
			return nil, err
		}
	}
	for _, c := range contexts {
		ctx.SetVars(c.GetVarsNearestScope())
	}
	return RValue{Value: resultSet}, nil
}

func (query singlePartQuery) Evaluate(ctx *EvalContext) (Value, error) {
	ret := ResultSet{}
	skip := -1
	limit := -1
	if query.ret != nil {
		var err error
		if query.ret.projection.skip != nil {
			skip, err = mustInt(query.ret.projection.skip.Evaluate(ctx))
			if err != nil {
				return nil, err
			}
		}
		if query.ret.projection.limit != nil {
			limit, err = mustInt(query.ret.projection.limit.Evaluate(ctx))
			if err != nil {
				return nil, err
			}
		}
	}

	project := func(rows []map[string]Value) error {
		for index, item := range rows {
			if skip != -1 && index < skip {
				continue
			}
			if limit != -1 && len(ret.Rows) >= limit {
				break
			}
			val, err := query.ret.projection.items.Project(ctx, item)
			if err != nil {
				return err
			}
			ret.Rows = append(ret.Rows, val)
		}
		return nil
	}
	results := ResultSet{}
	if len(query.read) > 0 {
		for _, r := range query.read {
			rs, err := r.GetResults(ctx)
			if err != nil {
				return nil, err
			}
			results.Add(rs)
		}

		for _, upd := range query.update {
			v, err := upd.Update(ctx, results)
			if err != nil {
				return nil, err
			}
			results = v.Get().(ResultSet)
		}
		if query.ret == nil {
			return RValue{Value: ResultSet{}}, nil
		}
		err := project(results.Rows)
		if err != nil {
			return nil, err
		}
		return RValue{Value: ret}, nil
	}

	for _, upd := range query.update {
		v, err := upd.TopLevelUpdate(ctx)
		if err != nil {
			return nil, err
		}
		if v != nil && v.Get() != nil {
			results = v.Get().(ResultSet)
		}
	}
	if query.ret == nil {
		return RValue{Value: ResultSet{}}, nil
	}

	if len(results.Rows) > 0 {
		for _, row := range results.Rows {
			val, err := query.ret.projection.items.Project(ctx, row)
			if err != nil {
				return nil, err
			}
			ret.Rows = append(ret.Rows, val)
		}
	} else {
		val, err := query.ret.projection.items.Project(ctx, nil)
		if err != nil {
			return nil, err
		}
		ret.Rows = append(ret.Rows, val)
	}
	return RValue{Value: ret}, nil
}

func (prj projectionItems) Project(ctx *EvalContext, values map[string]Value) (map[string]Value, error) {
	ret := make(map[string]Value)
	if prj.all {
		for k, v := range values {
			ret[k] = v
		}
		return ret, nil
	}

	for k, v := range values {
		ctx.SetVar(k, v)
	}
	for i, item := range prj.items {
		result, err := item.expr.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		var varName string
		if item.variable != nil {
			varName = string(*item.variable)
		} else {
			varName = strconv.Itoa(i + 1)
		}
		ret[varName] = result
	}
	return ret, nil
}

func (pe propertyExpression) Evaluate(ctx *EvalContext) (Value, error) {
	val, err := pe.atom.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	for i := range pe.lookup {
		prop := pe.lookup[i].String()
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
func (ls listComprehension) Evaluate(ctx *EvalContext) (Value, error)     { panic("Unimplemented") }
func (p patternComprehension) Evaluate(ctx *EvalContext) (Value, error)   { panic("Unimplemented") }
func (flt filterAtom) Evaluate(ctx *EvalContext) (Value, error)           { panic("Unimplemented") }
func (rel relationshipsPattern) Evaluate(ctx *EvalContext) (Value, error) { panic("Unimplemented") }
func (cnt countAtom) Evaluate(ctx *EvalContext) (Value, error)            { panic("Unimplemented") }
func (mq multiPartQuery) Evaluate(ctx *EvalContext) (Value, error)        { panic("Unimplemented") }
