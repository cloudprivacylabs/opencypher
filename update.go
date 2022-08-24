package opencypher

import (
	"fmt"

	"github.com/cloudprivacylabs/lpg"
)

func (s *set) Update(ctx *EvalContext, result ResultSet) (Value, error) {
	// Work on the cartesian product of result columns
	var err error
	subctx := ctx.SubContext()
	result.CartesianProduct(func(data map[string]Value) bool {
		subctx.SetVars(data)
		for i := range s.items {
			if err = s.items[i].update(subctx, data, result); err != nil {
				return false
			}
		}
		return true
	})
	if err != nil {
		return nil, err
	}
	return RValue{Value: result}, nil
}

func (s set) TopLevelUpdate(ctx *EvalContext) (Value, error) {
	return nil, fmt.Errorf("Cannot use SET at top level")
}

func (s *setItem) update(ctx *EvalContext, data map[string]Value, result ResultSet) (err error) {
	var exprResult Value

	if s.expression != nil {
		exprResult, err = s.expression.Evaluate(ctx)
		if err != nil {
			return err
		}
	}
	if s.property != nil {
		value, err := s.property.Evaluate(ctx)
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

	value, err := s.variable.Evaluate(ctx)
	lvalue, ok := value.(LValue)
	if !ok {
		return ErrNotAnLValue
	}

	getSourceProps := func() (map[string]interface{}, error) {
		sourceProps := make(map[string]interface{})
		exprValue := exprResult.Get()
		if node, ok := exprValue.(*lpg.Node); ok {
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
	switch s.op {
	case "=":
		switch v := lvalue.Get().(type) {
		case *lpg.Node:
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
		node, ok := lvalue.Get().(*lpg.Node)
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
		node, ok := lvalue.Get().(*lpg.Node)
		if !ok {
			return ErrInvalidAssignment("Not a node")
		}
		labels := node.GetLabels()
		for _, l := range s.nodeLabels {
			labels.Add(l.String())
		}
		node.SetLabels(labels)
	}
	return nil
}

func (deleteClause) TopLevelUpdate(ctx *EvalContext) (Value, error) {
	return nil, fmt.Errorf("Cannot use DELETE at top level")
}

func (d deleteClause) Update(ctx *EvalContext, result ResultSet) (Value, error) {
	subctx := ctx.SubContext()
	for _, row := range result.Rows {
		subctx.SetVars(row)
		for _, expr := range d.exprs {
			v, err := expr.Evaluate(subctx)
			if err != nil {
				return nil, err
			}
			if v.Get() == nil {
				continue
			}
			switch item := v.Get().(type) {
			case *lpg.Node:
				if item.GetEdges(lpg.OutgoingEdge).Next() || item.GetEdges(lpg.IncomingEdge).Next() {
					// Must have detach
					if !d.detach {
						return nil, fmt.Errorf("Cannot delete attached node")
					}
				}
				item.DetachAndRemove()

			case []*lpg.Edge:
				for _, e := range item {
					e.Remove()
				}
			}
		}
	}
	return RValue{Value: result}, nil
}

func (remove) TopLevelUpdate(ctx *EvalContext) (Value, error) {
	return nil, fmt.Errorf("Cannot use REMOVE at top level")
}

func (r remove) Update(ctx *EvalContext, result ResultSet) (Value, error) {
	subctx := ctx.SubContext()
	for _, row := range result.Rows {
		subctx.SetVars(row)
		for _, item := range r.items {
			if item.property != nil {
				value, err := item.property.Evaluate(subctx)
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
			v, err := subctx.GetVar(string(*item.variable))
			if err != nil {
				return nil, err
			}
			if v.Get() == nil {
				continue
			}
			node, ok := v.Get().(*lpg.Node)
			if !ok {
				return nil, fmt.Errorf("Expecting a node in remove statement")
			}
			labels := node.GetLabels()
			for _, l := range item.nodeLabels {
				labels.Remove(l.String())
			}
			node.SetLabels(labels)
		}
	}
	return RValue{Value: result}, nil
}

func (c create) TopLevelUpdate(ctx *EvalContext) (Value, error) {
	for _, part := range c.pattern.Parts {
		if _, _, err := part.Create(ctx); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (c create) Update(ctx *EvalContext, result ResultSet) (Value, error) {
	for _, row := range result.Rows {
		ctx.SetVars(row)
		if _, err := c.TopLevelUpdate(ctx); err != nil {
			return nil, err
		}
	}
	return RValue{Value: result}, nil
}

func (np nodePattern) Create(ctx *EvalContext) (string, *lpg.Node, error) {
	// Is there a variable
	var varName string
	if np.variable != nil {
		varName = string(*np.variable)
	}
	// Is the var defined already
	existingNode, err := ctx.GetVar(varName)
	if err == nil {
		// Var is defined already. Cannot have labels or properties
		if np.labels != nil || np.properties != nil {
			return "", nil, fmt.Errorf("Cannot specify labels or properties in bound node of a CREATE statement")
		}
		node, ok := existingNode.Get().(*lpg.Node)
		if !ok {
			return "", nil, fmt.Errorf("Not a node: %s", varName)
		}
		return varName, node, nil
	}
	node, err := np.createNode(ctx)
	if err != nil {
		return "", nil, err
	}
	if len(varName) > 0 {
		ctx.SetVar(varName, ValueOf(node))
	}
	return varName, node, nil
}

func (np nodePattern) createNode(ctx *EvalContext) (*lpg.Node, error) {
	labels := lpg.NewStringSet()
	if np.labels != nil {
		for _, n := range *np.labels {
			labels.Add(n.String())
		}
	}
	properties, err := np.properties.getPropertiesMap(ctx)
	if err != nil {
		return nil, err
	}
	node := ctx.graph.NewNode(labels.Slice(), properties)
	return node, nil
}

func (part PatternPart) Create(ctx *EvalContext) (*lpg.Node, []*lpg.Edge, error) {
	_, lastNode, err := part.start.Create(ctx)
	if err != nil {
		return nil, nil, err
	}
	firstNode := lastNode
	edges := make([]*lpg.Edge, 0)
	for _, pathPart := range part.path {
		_, targetNode, err := pathPart.node.Create(ctx)
		if err != nil {
			return nil, nil, err
		}
		edge, err := pathPart.rel.Create(ctx, lastNode, targetNode)
		if err != nil {
			return nil, nil, err
		}
		edges = append(edges, edge)
		lastNode = targetNode
	}
	if part.variable != nil {
		if len(edges) == 0 {
			ctx.SetVar(string(*part.variable), ValueOf(lastNode))
		} else {
			ctx.SetVar(string(*part.variable), ValueOf(edges))
		}
	}
	return firstNode, edges, nil
}

func (rel relationshipPattern) Create(ctx *EvalContext, from, to *lpg.Node) (*lpg.Edge, error) {
	if rel.rng != nil {
		return nil, fmt.Errorf("Cannot specify range in CREATE")
	}
	if rel.relTypes != nil && len(rel.relTypes.rel) > 1 {
		return nil, fmt.Errorf("Multiple labels for an edge")
	}
	var varName string
	if rel.variable != nil {
		varName = string(*rel.variable)
		// Is the var defined already
		_, err := ctx.GetVar(varName)
		if err == nil {
			// Var is defined already.
			return nil, fmt.Errorf("Cannot refer to an edge in CREATE")
		}
	}
	var label string
	if rel.relTypes != nil && len(rel.relTypes.rel) == 1 {
		label = rel.relTypes.rel[0].String()
	}
	properties, err := rel.properties.getPropertiesMap(ctx)
	if err != nil {
		return nil, err
	}
	var edge *lpg.Edge
	if rel.toLeft && !rel.toRight {
		edge = ctx.graph.NewEdge(to, from, label, properties)
	} else {
		edge = ctx.graph.NewEdge(from, to, label, properties)
	}
	if len(varName) > 0 {
		ctx.SetVar(varName, ValueOf([]*lpg.Edge{edge}))
	}
	return edge, nil
}

func (m merge) getResults(ctx *EvalContext) (map[string]struct{}, ResultSet, error) {
	pattern, err := m.pattern.getPattern(ctx)
	if err != nil {
		return nil, ResultSet{}, err
	}

	unbound := make(map[string]struct{})
	// Get unbound symbols of pattern.
	for symbol := range pattern.GetSymbolNames().M {
		_, err := ctx.GetVar(symbol)
		if err == nil {
			continue
		}
		// symbol is not bound
		unbound[symbol] = struct{}{}
	}

	results := matchResultAccumulator{
		evalCtx: ctx,
	}
	symbols, err := BuildPatternSymbols(ctx, pattern)
	if err != nil {
		return nil, ResultSet{}, err
	}
	err = pattern.Run(ctx.graph, symbols, &results)
	if err != nil {
		return nil, ResultSet{}, err
	}
	return unbound, results.result, nil
}

func (m merge) resultsToCtx(ctx *EvalContext, results ResultSet) {
	if len(results.Rows) > 0 {
		for k := range results.Rows[0] {
			if IsNamedVar(k) {
				col := results.Column(k)
				if len(col) == 1 {
					ctx.SetVar(k, col[0])
				} else {
					ctx.SetVar(k, RValue{Value: col})
				}
			}
		}
	}
}

func (m merge) doMerge(ctx *EvalContext) (created bool, matched bool, result ResultSet, err error) {
	_, result, err = m.getResults(ctx)
	if err != nil {
		return
	}
	if len(result.Rows) == 0 {
		// Nothing found
		subctx := ctx.SubContext()
		_, _, err = m.pattern.Create(subctx)
		if err != nil {
			return
		}
		vars := subctx.GetVarsNearestScope()
		row := make(map[string]Value)
		for k, v := range vars {
			row[k] = v
		}
		result = ResultSet{}
		result.Append(row)
		created = true
		return
	}
	// Things found
	matched = true
	return
}

func (m merge) processActions(ctx *EvalContext, created, matched bool, rs ResultSet) error {
	for _, action := range m.actions {
		if (created && action.on == mergeActionOnCreate) ||
			(matched && action.on == mergeActionOnMatch) {
			_, err := action.set.Update(ctx, rs)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m merge) Update(ctx *EvalContext, rs ResultSet) (Value, error) {
	results := ResultSet{}
	for _, row := range rs.Rows {
		subctx := ctx.SubContext()
		subctx.SetVars(row)
		created, matched, rs, err := m.doMerge(subctx)
		if err != nil {
			return nil, err
		}
		if err := m.processActions(subctx, created, matched, rs); err != nil {
			return nil, err
		}
		if len(rs.Rows) == 0 {
			results.Append(row)
		} else {
			for _, r := range rs.Rows {
				newRow := make(map[string]Value)
				for k, v := range row {
					newRow[k] = v
				}
				for k, v := range r {
					newRow[k] = v
				}
				results.Append(newRow)
			}
		}
	}
	return RValue{Value: results}, nil
}

func (m merge) TopLevelUpdate(ctx *EvalContext) (Value, error) {
	created, matched, rs, err := m.doMerge(ctx)
	if err != nil {
		return nil, err
	}
	if err := m.processActions(ctx, created, matched, rs); err != nil {
		return nil, err
	}
	results := ResultSet{}
	for _, r := range rs.Rows {
		newRow := make(map[string]Value)
		for k, v := range r {
			newRow[k] = v
		}
		results.Append(newRow)
	}
	return RValue{Value: results}, nil
}
