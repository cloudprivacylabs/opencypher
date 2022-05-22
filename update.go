package opencypher

import (
	"fmt"

	"github.com/cloudprivacylabs/opencypher/graph"
)

func (s *set) Update(ctx *EvalContext, result ResultSet) (Value, error) {
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

func (s set) TopLevelUpdate(ctx *EvalContext) error {
	return fmt.Errorf("Cannot use SET at top level")
}

func (s *setItem) update(ctx *EvalContext, data map[string]Value, result ResultSet) (err error) {
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

func (deleteClause) TopLevelUpdate(ctx *EvalContext) error {
	return fmt.Errorf("Cannot use DELETE at top level")
}

func (d deleteClause) Update(ctx *EvalContext, result ResultSet) (Value, error) {
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

func (remove) TopLevelUpdate(ctx *EvalContext) error {
	return fmt.Errorf("Cannot use REMOVE at top level")
}

func (r remove) Update(ctx *EvalContext, result ResultSet) (Value, error) {
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

func (c create) TopLevelUpdate(ctx *EvalContext) error {
	for _, part := range c.Pattern.Parts {
		if err := part.Create(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (c create) Update(ctx *EvalContext, result ResultSet) (Value, error) {
	for _, row := range result.Rows {
		for k, v := range row {
			ctx.SetVar(k, v)
		}
		if err := c.TopLevelUpdate(ctx); err != nil {
			return nil, err
		}
	}
	return RValue{Value: result}, nil
}

func (np NodePattern) Create(ctx *EvalContext) (string, graph.Node, error) {
	// Is there a variable
	var varName string
	if np.Var != nil {
		varName = string(*np.Var)
		// Is the var defined already
		existingNode, err := ctx.GetVar(varName)
		if err == nil {
			// Var is defined already. Cannot have labels or properties
			if np.Labels != nil || np.Properties != nil {
				return "", nil, fmt.Errorf("Cannot specify labels or properties in bound node of a CREATE statement")
			}
			node, ok := existingNode.Get().(graph.Node)
			if !ok {
				return "", nil, fmt.Errorf("Not a node: %s", varName)
			}
			return varName, node, nil
		}
	}
	labels := graph.NewStringSet()
	if np.Labels != nil {
		for _, n := range *np.Labels {
			labels.Add(n.String())
		}
	}
	properties, err := np.Properties.getPropertiesMap(ctx)
	if err != nil {
		return "", nil, err
	}
	node := ctx.graph.NewNode(labels.Slice(), properties)
	if len(varName) > 0 {
		ctx.SetVar(varName, ValueOf(node))
	}
	return varName, node, nil
}

func (part PatternPart) Create(ctx *EvalContext) error {
	_, lastNode, err := part.Start.Create(ctx)
	if err != nil {
		return err
	}
	edges := make([]graph.Edge, 0)
	for _, pathPart := range part.Path {
		_, targetNode, err := pathPart.Node.Create(ctx)
		if err != nil {
			return err
		}
		edge, err := pathPart.Rel.Create(ctx, lastNode, targetNode)
		if err != nil {
			return err
		}
		edges = append(edges, edge)
		lastNode = targetNode
	}
	if part.Var != nil {
		if len(edges) == 0 {
			ctx.SetVar(string(*part.Var), ValueOf(lastNode))
		} else {
			ctx.SetVar(string(*part.Var), ValueOf(edges))
		}
	}
	return nil
}

func (rel RelationshipPattern) Create(ctx *EvalContext, from, to graph.Node) (graph.Edge, error) {
	if rel.Range != nil {
		return nil, fmt.Errorf("Cannot specify range in CREATE")
	}
	if rel.RelTypes != nil && len(rel.RelTypes.Rel) > 1 {
		return nil, fmt.Errorf("Multiple labels for an edge")
	}
	var varName string
	if rel.Var != nil {
		varName = string(*rel.Var)
		// Is the var defined already
		_, err := ctx.GetVar(varName)
		if err == nil {
			// Var is defined already.
			return nil, fmt.Errorf("Cannot refer to an edge in CREATE")
		}
	}
	var label string
	if rel.RelTypes != nil && len(rel.RelTypes.Rel) == 1 {
		label = rel.RelTypes.Rel[0].String()
	}
	properties, err := rel.Properties.getPropertiesMap(ctx)
	if err != nil {
		return nil, err
	}
	var edge graph.Edge
	if rel.ToLeft && !rel.ToRight {
		edge = ctx.graph.NewEdge(to, from, label, properties)
	} else if !rel.ToLeft && rel.ToRight {
		edge = ctx.graph.NewEdge(from, to, label, properties)
	} else {
		return nil, fmt.Errorf("Ambiguous edge direction")
	}
	if len(varName) > 0 {
		ctx.SetVar(varName, ValueOf([]graph.Edge{edge}))
	}
	return edge, nil
}
