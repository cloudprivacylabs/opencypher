package opencypher

import (
	"fmt"
	"time"

	"github.com/cloudprivacylabs/lpg"
)

func mustInt(v Value, err error) (int, error) {
	if err != nil {
		return 0, err
	}
	i, ok := v.Get().(int)
	if !ok {
		return 0, ErrIntValueRequired
	}
	return i, nil
}

type ErrInvalidFunctionCall struct {
	Msg string
}

func (e ErrInvalidFunctionCall) Error() string {
	return "Invalid function call: " + e.Msg
}

var globalFuncs = map[string]Function{
	"range":     rangeFunc,
	"labels":    labelsFunc,
	"timestamp": timestampFunc,
	"type":      typeFunc,
}

func rangeFunc(ctx *EvalContext, args []Evaluatable) (Value, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, ErrInvalidFunctionCall{"range(start,stop,[step]) needs 3 args"}
	}
	start, err := mustInt(args[0].Evaluate(ctx))
	if err != nil {
		return nil, err
	}
	end, err := mustInt(args[1].Evaluate(ctx))
	if err != nil {
		return nil, err
	}
	skip := 1
	if len(args) == 3 {
		skip, err = mustInt(args[2].Evaluate(ctx))
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

func labelsFunc(ctx *EvalContext, args []Evaluatable) (Value, error) {
	if len(args) != 1 {
		return nil, ErrInvalidFunctionCall{"labels(node) needs 1 arg"}
	}
	v, err := args[0].Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	if v.Get() == nil {
		return RValue{}, nil
	}
	node, ok := v.Get().(*lpg.Node)
	if !ok {
		return nil, fmt.Errorf("Not a node")
	}
	return RValue{Value: node.GetLabels().Slice()}, nil
}

func timestampFunc(ctx *EvalContext, args []Evaluatable) (Value, error) {
	if len(args) != 0 {
		return nil, ErrInvalidFunctionCall{"timestamp() does not take args"}
	}
	return RValue{Value: int(time.Now().Unix())}, nil
}

func typeFunc(ctx *EvalContext, args []Evaluatable) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("type() needs one arg")
	}
	v, err := args[0].Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	edge, ok := v.Get().(*lpg.Edge)
	if ok {
		return RValue{Value: edge.GetLabel()}, nil
	}
	edges, ok := v.Get().([]*lpg.Edge)
	if !ok || len(edges) != 1 {
		return nil, fmt.Errorf("Cannot determine type of %T", v.Get())
	}
	return RValue{Value: edges[0].GetLabel()}, nil
}
