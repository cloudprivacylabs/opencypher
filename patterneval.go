package opencypher

import (
	"github.com/cloudprivacylabs/opencypher/graph"
)

type ErrInvalidValueReferenceInPattern struct {
	Symbol string
}

func (e ErrInvalidValueReferenceInPattern) Error() string {
	return "Invalid value reference in pattern: " + e.Symbol
}

func (properties Properties) AsLiteral(ctx *EvalContext) ([]mapKeyValue, error) {
	if properties.Param != nil {
		param, err := ctx.GetParameter(string(*properties.Param))
		if err != nil {
			return nil, err
		}
		lit, ok := param.Get().(map[string]Value)
		if !ok {
			return nil, ErrPropertiesParameterExpected
		}
		kv := make([]mapKeyValue, 0, len(lit))
		for k, v := range lit {
			kv = append(kv, mapKeyValue{key: k, value: v})
		}
		return kv, nil
	}
	if properties.Map != nil {
		return properties.Map.keyValues, nil
	}
	return nil, nil
}

type matchResultAccumulator struct {
	evalCtx *EvalContext
	result  ResultSet
	err     error
}

func (acc *matchResultAccumulator) StoreResult(ctx *graph.MatchContext, path interface{}, symbols map[string]interface{}) {
	if acc.err != nil {
		return
	}
	// Record results in the context
	result := make(map[string]Value)
	for k, v := range symbols {
		result[k] = RValue{Value: v}
		acc.evalCtx.SetVar(k, RValue{Value: v})
	}
	acc.result.Append(result)
}

func (match Match) GetResults(ctx *EvalContext) (ResultSet, error) {
	patterns := make([]graph.Pattern, 0, len(match.Pattern.Parts))
	for i := range match.Pattern.Parts {
		p, err := match.Pattern.Parts[i].getPattern(ctx)
		if err != nil {
			return ResultSet{}, err
		}
		patterns = append(patterns, p)
	}
	newContext := ctx.SubContext()

	results := make([]matchResultAccumulator, len(patterns))
	for i := range patterns {
		results[i].evalCtx = newContext
		symbols, err := BuildPatternSymbols(ctx, patterns[i])
		if err != nil {
			return ResultSet{}, err
		}

		err = patterns[i].Run(ctx.graph, symbols, &results[i])
		if err != nil {
			return ResultSet{}, err
		}
	}

	resultSets := make([]ResultSet, 0, len(patterns))
	for _, r := range results {
		resultSets = append(resultSets, r.result)
	}
	var err error
	// Build resultset from results
	rs := CartesianProduct(resultSets, match.Optional, func(row map[string]Value) bool {
		if match.Where == nil {
			return true
		}
		if err != nil {
			return false
		}
		for k, v := range row {
			newContext.SetVar(k, v)
		}
		rs, e := match.Where.Evaluate(newContext)
		if e != nil {
			err = e
			return false
		}
		if b, _ := ValueAsBool(rs); !b {
			return false
		}
		return true
	})

	return rs, err
}

// BuildPatternSymbols copies all the symbols referenced in the
// pattern from the context, and puts them in a map
func BuildPatternSymbols(ctx *EvalContext, pattern graph.Pattern) (map[string]*graph.PatternSymbol, error) {
	symbols := make(map[string]*graph.PatternSymbol)
	for symbol := range pattern.GetSymbolNames() {
		// If a symbol is in the context, then get its value. Otherwise, it is a local symbol. Add to context
		value, err := ctx.GetVar(symbol)
		if err != nil {
			continue
		}
		ps := &graph.PatternSymbol{}
		// A variable with the same name exists
		// Must be a Node, or []Edge
		switch val := value.Get().(type) {
		case graph.Node:
			ps.AddNode(val)
		case []graph.Edge:
			ps.AddPath(val)
		default:
			return nil, ErrInvalidValueReferenceInPattern{Symbol: symbol}
		}
		symbols[symbol] = ps
	}
	return symbols, nil
}

func (part PatternPart) getPattern(ctx *EvalContext) (graph.Pattern, error) {
	pattern := make([]graph.PatternItem, 0, len(part.path)*2+1)
	np, err := part.start.getPattern(ctx)
	if err != nil {
		return nil, err
	}
	pattern = append(pattern, np)
	for _, pathItem := range part.path {
		pi, err := pathItem.rel.getPattern(ctx)
		if err != nil {
			return nil, err
		}
		pattern = append(pattern, pi)

		pi, err = pathItem.node.getPattern(ctx)
		if err != nil {
			return nil, err
		}
		pattern = append(pattern, pi)
	}
	return pattern, nil
}

func (np nodePattern) getPattern(ctx *EvalContext) (graph.PatternItem, error) {
	ret := graph.PatternItem{}
	if np.variable != nil {
		ret.Name = string(*np.variable)
	}
	ret.Labels = np.labels.getPattern()
	var err error
	props, err := np.properties.getPattern(ctx)
	if err != nil {
		return graph.PatternItem{}, err
	}
	if len(props) > 0 {
		ret.Properties = make(map[string]interface{})
		for k, v := range props {
			ret.Properties[k] = v.Get()
		}
	}
	return ret, nil
}

func (rp relationshipPattern) getPattern(ctx *EvalContext) (graph.PatternItem, error) {
	ret := graph.PatternItem{}
	if rp.variable != nil {
		ret.Name = string(*rp.variable)
	}
	ret.Labels = rp.relTypes.getPattern()
	if rp.rng != nil {
		from, to, err := rp.rng.Evaluate(ctx)
		if err != nil {
			return graph.PatternItem{}, err
		}
		if from != nil {
			ret.Min = *from
		} else {
			ret.Min = -1
		}
		if to != nil {
			ret.Max = *to
		} else {
			ret.Max = -1
		}
	} else {
		ret.Min, ret.Max = 1, 1
	}
	if !rp.toRight && !rp.toLeft {
		ret.Undirected = true
	} else if rp.toLeft {
		ret.ToLeft = rp.toLeft
	}

	var err error
	props, err := rp.properties.getPattern(ctx)
	if err != nil {
		return graph.PatternItem{}, err
	}
	if len(props) > 0 {
		ret.Properties = make(map[string]interface{})
		for k, v := range props {
			ret.Properties[k] = v.Get()
		}
	}
	return ret, nil
}

func (rt *relationshipTypes) getPattern() graph.StringSet {
	if rt == nil {
		return nil
	}
	if len(rt.rel) == 0 {
		return nil
	}
	ret := graph.NewStringSet()
	for _, r := range rt.rel {
		s := r.String()
		if len(s) > 0 {
			ret.Add(s)
		}
	}
	if len(ret) == 0 {
		return nil
	}
	return ret
}

func (nl *NodeLabels) getPattern() graph.StringSet {
	if nl == nil {
		return nil
	}
	ret := graph.NewStringSet()
	for _, l := range *nl {
		s := l.String()
		if len(s) > 0 {
			ret.Add(s)
		}
	}
	if len(ret) == 0 {
		return nil
	}
	return ret
}

func (p *Properties) getPattern(ctx *EvalContext) (map[string]Value, error) {
	if p == nil {
		return nil, nil
	}
	if p.Param != nil {
		value, err := ctx.GetParameter(string(*p.Param))
		if err != nil {
			return nil, err
		}
		m, ok := value.Get().(map[string]Value)
		if !ok {
			return nil, ErrPropertiesParameterExpected
		}
		return m, nil
	}
	if p.Map != nil {
		value, err := p.Map.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		m, ok := value.Get().(map[string]Value)
		if !ok {
			return nil, ErrPropertiesExpected
		}
		return m, nil
	}

	return nil, nil
}

func (p *Properties) getPropertiesMap(ctx *EvalContext) (map[string]interface{}, error) {
	var properties map[string]interface{}
	if p != nil {
		propertiesAsValue, err := p.getPattern(ctx)
		if err != nil {
			return nil, err
		}
		properties = make(map[string]interface{})
		for k, v := range propertiesAsValue {
			properties[k] = v.Get()
		}
	}
	return properties, nil
}
