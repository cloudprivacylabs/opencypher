package opencypher

import (
	"strings"
	"unicode"

	"github.com/cloudprivacylabs/lpg/v2"
)

type ErrUnknownParameter struct {
	Key string
}

func (e ErrUnknownParameter) Error() string { return "Unknown parameter: " + e.Key }

type EvalContext struct {
	parent     *EvalContext
	funcMap    map[string]Function
	variables  map[string]Value
	parameters map[string]Value
	graph      *lpg.Graph

	// If this function is non-nil, it will be called to filter property
	// values when setting properties of nodes or edges
	PropertyValueFromNativeFilter func(string, interface{}) interface{}
}

func NewEvalContext(graph *lpg.Graph) *EvalContext {
	return &EvalContext{
		funcMap:    globalFuncs,
		variables:  make(map[string]Value),
		parameters: make(map[string]Value),
		graph:      graph,
	}
}

// SubContext creates a new subcontext with a new variable set
func (ctx *EvalContext) SubContext() *EvalContext {
	return &EvalContext{
		parent:                        ctx,
		funcMap:                       ctx.funcMap,
		variables:                     make(map[string]Value),
		parameters:                    make(map[string]Value),
		graph:                         ctx.graph,
		PropertyValueFromNativeFilter: ctx.PropertyValueFromNativeFilter,
	}
}

// PropertyValueFromNative calls the ctx.PropertyValueFromNativeFilter
// to give the caller a chance to change property values
func (ctx *EvalContext) PropertyValueFromNative(key string, value interface{}) interface{} {
	if ctx.PropertyValueFromNativeFilter != nil {
		return ctx.PropertyValueFromNativeFilter(key, value)
	}
	return value
}

// SetParameter sets a parameter to be used in expressions
func (ctx *EvalContext) SetParameter(key string, value Value) *EvalContext {
	ctx.parameters[key] = value
	return ctx
}

func (ctx *EvalContext) GetParameter(key string) (Value, error) {
	value, ok := ctx.parameters[key]
	if ok {
		return value, nil
	}
	if ctx.parent == nil {
		return nil, ErrUnknownParameter{Key: key}
	}
	return ctx.parent.GetParameter(key)
}

type ErrUnknownFunction struct {
	Name string
}

func (e ErrUnknownFunction) Error() string { return "Unknown function: " + e.Name }

type ErrUnknownVariable struct {
	Name string
}

func (e ErrUnknownVariable) Error() string { return "Unknown variable:" + e.Name }

func (ctx *EvalContext) getFunction(name string) (Function, error) {
	f, ok := ctx.funcMap[name]
	if !ok {
		return Function{}, ErrUnknownFunction{name}
	}
	return f, nil
}

func (ctx *EvalContext) GetFunction(name []string) (Function, error) {
	bld := strings.Builder{}
	for i, x := range name {
		if i > 0 {
			bld.WriteRune('.')
		}
		bld.WriteString(string(x))
	}
	return ctx.getFunction(bld.String())
}

func (ctx *EvalContext) GetVar(name string) (Value, error) {
	val, ok := ctx.variables[name]
	if !ok {
		if ctx.parent == nil {
			return nil, ErrUnknownVariable{Name: name}
		}
		return ctx.parent.GetVar(name)
	}
	if rv, ok := val.(RValue); ok {
		rv.Const = false
		return rv, nil
	}
	return val, nil
}

func (ctx *EvalContext) GetVarsNearestScope() map[string]Value {
	return ctx.variables
}

func (ctx *EvalContext) SetVar(name string, value Value) {
	ctx.variables[name] = value
}

func (ctx *EvalContext) RemoveVar(name string) {
	delete(ctx.variables, name)
}

func (ctx *EvalContext) SetVars(m map[string]Value) {
	for k, v := range m {
		if !unicode.IsDigit(rune(k[0])) {
			ctx.SetVar(k, v)
		}
	}
}

func IsNamedVar(name string) bool {
	return !unicode.IsDigit(rune(name[0]))
}
