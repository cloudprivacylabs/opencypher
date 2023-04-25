package opencypher

import (
	"github.com/cloudprivacylabs/lpg/v2"
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
	result  []ResultPath
	err     error
}

func (acc *matchResultAccumulator) StoreResult(ctx *lpg.MatchContext, path *lpg.Path, symbols map[string]interface{}) {
	if acc.err != nil {
		return
	}
	// Record results in the context
	acc.result = append(acc.result, ResultPath{Result: path, Symbols: make(map[string]Value)})
	for k, v := range symbols {
		acc.result[len(acc.result)-1].Symbols[k] = ValueOf(v)
	}
}

func (match Match) GetResults(ctx *EvalContext) ([]ResultPath, error) {
	patterns := make([]lpg.Pattern, 0, len(match.Pattern.Parts))
	for i := range match.Pattern.Parts {
		p, err := match.Pattern.Parts[i].getPattern(ctx)
		if err != nil {
			return []ResultPath{}, err
		}
		patterns = append(patterns, p)
	}

	var nextPattern func(*EvalContext, []lpg.Pattern, int) error

	returnResults := []ResultPath{}
	// currentRow := make([]map[string]Value, len(patterns))

	// addRow := func() {
	// 	newRow := make(map[string]Value)
	// 	for _, x := range currentRow {
	// 		for k, v := range x {
	// 			newRow[k] = v
	// 		}
	// 	}
	// 	results.Rows = append(results.Rows, newRow)
	// }

	nextPattern = func(prevContext *EvalContext, pat []lpg.Pattern, index int) error {
		newContext := prevContext.SubContext()
		symbols, err := BuildPatternSymbols(newContext, pat[0])
		if err != nil {
			return err
		}
		results := matchResultAccumulator{
			evalCtx: newContext,
			result:  []ResultPath{},
		}
		err = pat[0].Run(newContext.graph, symbols, &results)
		if err != nil {
			return err
		}
		for _, row := range results.result {
			for k, v := range row.Symbols {
				newContext.SetVar(k, RValue{Value: v})
			}
			// currentRow[index] = row
			if len(pat) > 1 {
				if err := nextPattern(newContext, pat[1:], index+1); err != nil {
					return err
				}
			} else {
				if match.Where != nil {
					rs, err := match.Where.Evaluate(newContext)
					if err != nil {
						return err
					}
					if b, _ := ValueAsBool(rs); b {
						// addRow()
						returnResults = append(returnResults, row)
					}
				} else {
					returnResults = append(returnResults, row)
				}
			}
		}
		return nil
	}
	if err := nextPattern(ctx, patterns, 0); err != nil {
		return []ResultPath{}, err
	}
	return returnResults, nil
}

// BuildPatternSymbols copies all the symbols referenced in the
// pattern from the context, and puts them in a map.
func BuildPatternSymbols(ctx *EvalContext, pattern lpg.Pattern) (map[string]*lpg.PatternSymbol, error) {
	symbols := make(map[string]*lpg.PatternSymbol)
	for symbol := range pattern.GetSymbolNames().M {
		// If a symbol is in the context, then get its value. Otherwise, it is a local symbol. Add to context
		value, err := ctx.GetVar(symbol)
		if err != nil {
			continue
		}
		ps := &lpg.PatternSymbol{}
		// A variable with the same name exists
		// Must be a Node, or []Edge
		v := value.Get()
		if v == nil {
			symbols[symbol] = ps
		} else {
			switch val := v.(type) {
			case *lpg.Node:
				ps.AddNode(val)
			case *lpg.Path:
				ps.AddPath(val)
			// case RValue:
			// 	switch rv := val.Value.(type) {
			// 	case *lpg.Node:
			// 		ps.AddNode(rv)
			// 	}
			default:
				// v RValue{Value: *lpg.Node}
				return nil, ErrInvalidValueReferenceInPattern{Symbol: symbol}
			}
			symbols[symbol] = ps
		}
	}
	return symbols, nil
}

func (part PatternPart) getPattern(ctx *EvalContext) (lpg.Pattern, error) {
	pattern := make([]lpg.PatternItem, 0, len(part.path)*2+1)
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

func (np nodePattern) getPattern(ctx *EvalContext) (lpg.PatternItem, error) {
	ret := lpg.PatternItem{}
	if np.variable != nil {
		ret.Name = string(*np.variable)
	}
	ret.Labels = np.labels.getPattern()
	var err error
	props, err := np.properties.getPattern(ctx)
	if err != nil {
		return lpg.PatternItem{}, err
	}
	if len(props) > 0 {
		ret.Properties = make(map[string]interface{})
		for k, v := range props {
			ret.Properties[k] = v.Get()
		}
	}
	return ret, nil
}

func (rp relationshipPattern) getPattern(ctx *EvalContext) (lpg.PatternItem, error) {
	ret := lpg.PatternItem{}
	if rp.variable != nil {
		ret.Name = string(*rp.variable)
	}
	ret.Labels = rp.relTypes.getPattern()
	if rp.rng != nil {
		from, to, err := rp.rng.Evaluate(ctx)
		if err != nil {
			return lpg.PatternItem{}, err
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
		return lpg.PatternItem{}, err
	}
	if len(props) > 0 {
		ret.Properties = make(map[string]interface{})
		for k, v := range props {
			ret.Properties[k] = v.Get()
		}
	}
	return ret, nil
}

func (rt *relationshipTypes) getPattern() lpg.StringSet {
	if rt == nil {
		return lpg.StringSet{}
	}
	if len(rt.rel) == 0 {
		return lpg.StringSet{}
	}
	ret := lpg.NewStringSet()
	for _, r := range rt.rel {
		s := r.String()
		if len(s) > 0 {
			ret.Add(s)
		}
	}
	if ret.Len() == 0 {
		return lpg.StringSet{}
	}
	return ret
}

func (nl *NodeLabels) getPattern() lpg.StringSet {
	if nl == nil {
		return lpg.StringSet{}
	}
	ret := lpg.NewStringSet()
	for _, l := range *nl {
		s := l.String()
		if len(s) > 0 {
			ret.Add(s)
		}
	}
	if ret.Len() == 0 {
		return lpg.StringSet{}
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
