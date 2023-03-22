package opencypher

import (
	"fmt"
	"time"

	"github.com/cloudprivacylabs/lpg"
)

// Function describes a function
type Function struct {
	Name      string
	MinArgs   int
	MaxArgs   int
	Func      func(*EvalContext, []Evaluatable) (Value, error)
	ValueFunc func(*EvalContext, []Value) (Value, error)
}

type ErrInvalidFunctionCall struct {
	Msg string
}

func (e ErrInvalidFunctionCall) Error() string {
	return "Invalid function call: " + e.Msg
}

func RegisterGlobalFunc(fn ...Function) {
	for _, f := range fn {
		globalFuncs[f.Name] = f
	}
}

var globalFuncs = map[string]Function{
	"range": Function{
		Name:      "range",
		MinArgs:   2,
		MaxArgs:   3,
		ValueFunc: rangeFunc,
	},
	"labels": Function{
		Name:      "labels",
		MinArgs:   1,
		MaxArgs:   1,
		ValueFunc: labelsFunc,
	},
	"timestamp": Function{
		Name:      "timestamp",
		MinArgs:   0,
		MaxArgs:   0,
		ValueFunc: timestampFunc,
	},
	"type": Function{
		Name:      "type",
		MinArgs:   1,
		MaxArgs:   1,
		ValueFunc: typeFunc,
	},
	"size": Function{
		Name:      "size",
		MinArgs:   1,
		MaxArgs:   1,
		ValueFunc: sizeFunc,
	},
}

func rangeFunc(ctx *EvalContext, args []Value) (Value, error) {
	start, err := ValueAsInt(args[0])
	if err != nil {
		return nil, err
	}
	end, err := ValueAsInt(args[1])
	if err != nil {
		return nil, err
	}
	skip := 1
	if len(args) == 3 {
		skip, err = ValueAsInt(args[2])
		if err != nil {
			return nil, err
		}
	}
	if (end <= start && skip > 0) || (end >= start && skip < 0) || skip == 0 {
		return RValue{Value: []Value{}}, nil
	}
	arr := make([]Value, 0)
	if end > start {
		for at := start; at < end; at += skip {
			arr = append(arr, RValue{Value: at})
		}
	} else {
		for at := start; at > end; at += skip {
			arr = append(arr, RValue{Value: at})
		}
	}
	return RValue{Value: arr}, nil
}

func labelsFunc(ctx *EvalContext, args []Value) (Value, error) {
	if args[0].Get() == nil {
		return RValue{}, nil
	}
	node, ok := args[0].Get().(*lpg.Node)
	if !ok {
		return nil, fmt.Errorf("Not a node")
	}
	return RValue{Value: node.GetLabels().Slice()}, nil
}

func timestampFunc(ctx *EvalContext, args []Value) (Value, error) {
	return RValue{Value: int(time.Now().Unix())}, nil
}

func typeFunc(ctx *EvalContext, args []Value) (Value, error) {
	edge, ok := args[0].Get().(*lpg.Edge)
	if ok {
		return RValue{Value: edge.GetLabel()}, nil
	}
	path, ok := args[0].Get().(*lpg.Path)
	if !ok || path.NumEdges() != 1 {
		return nil, fmt.Errorf("Cannot determine type of %T", args[0].Get())
	}
	return RValue{Value: path.GetEdge(0).GetLabel()}, nil
}

func sizeFunc(ctx *EvalContext, args []Value) (Value, error) {
	val := args[0].Get()
	if val == nil {
		return RValue{}, nil
	}
	if arr, ok := val.([]Value); ok {
		return RValue{Value: len(arr)}, nil
	}
	if str, ok := val.(string); ok {
		return RValue{Value: len(str)}, nil
	}
	return RValue{}, nil
}
