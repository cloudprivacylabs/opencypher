package opencypher

func mustInt(v Value, err error) (int, error) {
	if err != nil {
		return 0, err
	}
	i, ok := v.Value.(int)
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
	"range": rangeFunc,
}

func rangeFunc(ctx *EvalContext, args []Evaluatable) (Value, error) {
	if len(args) < 2 || len(args) > 3 {
		return Value{}, ErrInvalidFunctionCall{"range(start,stop,[step]) needs 3 args"}
	}
	start, err := mustInt(args[0].Evaluate(ctx))
	if err != nil {
		return Value{}, err
	}
	end, err := mustInt(args[1].Evaluate(ctx))
	if err != nil {
		return Value{}, err
	}
	skip := 1
	if len(args) == 3 {
		skip, err = mustInt(args[2].Evaluate(ctx))
		if err != nil {
			return Value{}, err
		}
	}
	if (end <= start && skip > 0) || (end >= start && skip < 0) || skip == 0 {
		return Value{Value: []Value{}}, nil
	}
	arr := make([]Value, 0)
	if end > start {
		for at := start; at < end; at += skip {
			arr = append(arr, Value{Value: at})
		}
	} else {
		for at := start; at > end; at += skip {
			arr = append(arr, Value{Value: at})
		}
	}
	return Value{Value: arr}, nil
}
