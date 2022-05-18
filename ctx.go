package opencypher

import (
	"strings"

	"github.com/cloudprivacylabs/opencypher/graph"
)

type ErrUnknownParameter struct {
	Key string
}

func (e ErrUnknownParameter) Error() string { return "Unknown parameter: " + e.Key }

type Function func(*EvalContext, []Evaluatable) (Value, error)

type EvalContext struct {
	parent     *EvalContext
	funcMap    map[string]Function
	variables  map[string]Value
	parameters map[string]Value
	graph      graph.Graph
}

func NewEvalContext(graph graph.Graph) *EvalContext {
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
		parent:     ctx,
		funcMap:    ctx.funcMap,
		variables:  make(map[string]Value),
		parameters: make(map[string]Value),
		graph:      ctx.graph,
	}
}

// SetParameter sets a parameter to be used in expressions
func (ctx *EvalContext) SetParameter(key string, value Value) *EvalContext {
	ctx.parameters[key] = value
	return ctx
}

func (ctx *EvalContext) GetParameter(key string) (Value, error) {
	value, ok := ctx.parameters[key]
	if !ok {
		return nil, ErrUnknownParameter{Key: key}
	}
	return value, nil
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
	f := ctx.funcMap[name]
	if f == nil {
		return nil, ErrUnknownFunction{name}
	}
	return f, nil
}

func (ctx *EvalContext) GetFunction(name []SymbolicName) (Function, error) {
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

func (ctx *EvalContext) SetVar(name string, value Value) {
	ctx.variables[name] = value
}
