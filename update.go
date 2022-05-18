package opencypher

import (
	"fmt"

	"github.com/cloudprivacylabs/opencypher/graph"
)

func (s *Set) Update(ctx *EvalContext, result ResultSet) (Value, error) {
	// Work on the cartesian product of result columns
	var err error
	subctx := ctx.SubContext()
	result.CartesianProduct(func(data map[string]Value) bool {
		for k, v := range data {
			subctx.SetVar(k, v)
		}
		for i := range s.Items {
			if err = s.Items[i].update(subctx, data, result); err != nil {
				return false
			}
		}
		return true
	})
	if err != nil {
		return nil, err
	}
	return RValue{}, nil
}

func (s *SetItem) update(ctx *EvalContext, data map[string]Value, result ResultSet) (err error) {
	var exprResult Value

	if s.Expression != nil {
		exprResult, err = s.Expression.Evaluate(ctx)
		if err != nil {
			return err
		}
	}
	if s.Property != nil {
		value, err := s.Property.Evaluate(ctx)
		if err != nil {
			return err
		}
		lvalue, ok := value.(LValue)
		if !ok {
			return ErrNotAnLValue
		}
		lvalue.Set(exprResult.Get())
		return nil
	}

	value, err := s.Variable.Evaluate(ctx)
	lvalue, ok := value.(LValue)
	if !ok {
		return ErrNotAnLValue
	}

	getSourceProps := func() (map[string]interface{}, error) {
		sourceProps := make(map[string]interface{})
		exprValue := exprResult.Get()
		if node, ok := exprValue.(graph.Node); ok {
			node.ForEachProperty(func(key string, value interface{}) bool {
				sourceProps[key] = value
				return true
			})
		} else if mp, ok := exprValue.(map[string]Value); ok {
			for k, v := range mp {
				sourceProps[k] = v.Get()
			}
		} else {
			return nil, ErrInvalidAssignment(fmt.Sprint(exprValue))
		}
		return sourceProps, nil
	}
	switch s.Op {
	case "=":
		switch v := lvalue.Get().(type) {
		case graph.Node:
			sourceProps, err := getSourceProps()
			if err != nil {
				return err
			}
			props := make([]string, 0)
			v.ForEachProperty(func(key string, _ interface{}) bool {
				props = append(props, key)
				return true
			})
			for _, p := range props {
				v.RemoveProperty(p)
			}
			for key, val := range sourceProps {
				v.SetProperty(key, val)
			}
		default:
			return ErrInvalidAssignment(fmt.Sprintf("%T: %v", v, v))
		}
	case "+=":
		sourceProps, err := getSourceProps()
		if err != nil {
			return err
		}
		node, ok := lvalue.Get().(graph.Node)
		if !ok {
			return ErrInvalidAssignment(fmt.Sprintf("%T: %v", lvalue.Get(), lvalue.Get()))
		}
		for k, v := range sourceProps {
			if v == nil {
				node.RemoveProperty(k)
				continue
			}
			node.SetProperty(k, v)
		}
	default: // NodeLabels
		node, ok := lvalue.Get().(graph.Node)
		if !ok {
			return ErrInvalidAssignment("Not a node")
		}
		labels := node.GetLabels()
		for _, l := range s.NodeLabels {
			labels.Add(l.String())
		}
		node.SetLabels(labels)
	}
	return nil
}

func (d Delete) Update(ctx *EvalContext, result ResultSet) (Value, error) {
	subctx := ctx.SubContext()
	for _, row := range result.Rows {
		for k, v := range row {
			subctx.SetVar(k, v)
		}
		for _, expr := range d.Exprs {
			v, err := expr.Evaluate(subctx)
			if err != nil {
				return nil, err
			}
			if v.Get() == nil {
				continue
			}
			switch item := v.Get().(type) {
			case graph.Node:
				if item.GetEdges(graph.OutgoingEdge).Next() || item.GetEdges(graph.IncomingEdge).Next() {
					// Must have detach
					if !d.Detach {
						return nil, fmt.Errorf("Cannot delete attached node")
					}
				}
				item.DetachAndRemove()

			case []graph.Edge:
				for _, e := range item {
					e.Remove()
				}
			}
		}
	}
	return RValue{Value: ResultSet{}}, nil
}

func (r Remove) Update(ctx *EvalContext, result ResultSet) (Value, error) {
	subctx := ctx.SubContext()
	for _, row := range result.Rows {
		for k, v := range row {
			subctx.SetVar(k, v)
		}
		for _, item := range r.Items {
			if item.Property != nil {
				value, err := item.Property.Evaluate(subctx)
				if err != nil {
					return nil, err
				}
				lvalue, ok := value.(LValue)
				if !ok {
					return nil, ErrNotAnLValue
				}
				lvalue.Set(nil)
				continue
			}
			v, err := subctx.GetVar(string(*item.Variable))
			if err != nil {
				return nil, err
			}
			if v.Get() == nil {
				continue
			}
			node, ok := v.Get().(graph.Node)
			if !ok {
				return nil, fmt.Errorf("Expecting a node in remove statement")
			}
			labels := node.GetLabels()
			for _, l := range item.NodeLabels {
				labels.Remove(l.String())
			}
			node.SetLabels(labels)
		}
	}
	return RValue{Value: result}, nil
}
