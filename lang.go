package opencypher

import (
	"fmt"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/cloudprivacylabs/lpg"
	"github.com/cloudprivacylabs/opencypher/parser"
)

//go:generate antlr4 -Dlanguage=Go Cypher.g4 -o parser

type errorListener struct {
	antlr.DefaultErrorListener
	err error
}

type ErrSyntax string
type ErrInvalidExpression string

func (e ErrSyntax) Error() string            { return "Syntax error: " + string(e) }
func (e ErrInvalidExpression) Error() string { return "Invalid expression: " + string(e) }

func (lst *errorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	if lst.err == nil {
		lst.err = ErrSyntax(fmt.Sprintf("line %d:%d %s ", line, column, msg))
	}
}

// GetParser returns a parser that will parse the input string
func GetParser(input string) *parser.CypherParser {
	lexer := parser.NewCypherLexer(antlr.NewInputStream(input))
	stream := antlr.NewCommonTokenStream(lexer, 0)
	p := parser.NewCypherParser(stream)
	p.BuildParseTrees = true
	return p
}

// GetEvaluatable returns an evaluatable object
func Parse(input string) (Evaluatable, error) {
	pr := GetParser(input)
	errListener := errorListener{}
	pr.AddErrorListener(&errListener)
	c := pr.OC_Cypher()
	if errListener.err != nil {
		return nil, errListener.err
	}
	out := oC_Cypher(c.(*parser.OC_CypherContext))
	return out, nil
}

func ParseAndEvaluate(input string, ctx *EvalContext) (Value, error) {
	e, err := Parse(input)
	if err != nil {
		return nil, err
	}
	return e.Evaluate(ctx)
}

// ParsePatternExpr parses the pattern expression that starts at the
// current node named 'this', and describes a path reaching one or
// more nodes named 'target'. For instance:
//
//  (this)-[]->(target)
//
// will return all nodes reachable from the current node by one step.
//
// This expression:
//
//  (this)<[a]-()-[]->(target :x)
//
// will start from the current node, go back one nore following an
// edge with label `a`, and then move to a node with label `x`
func ParsePatternExpr(expr string) (PatternPart, error) {
	pr := GetParser(expr)
	errListener := errorListener{}
	pr.AddErrorListener(&errListener)
	c := pr.OC_PatternPart()
	if errListener.err != nil {
		return PatternPart{}, errListener.err
	}
	out := oC_PatternPart(c.(*parser.OC_PatternPartContext))
	return out, nil
}

// FindRelative evaluates a pattern expression starting at the given
// node. It may return zero or more nodes reached from the node
func (p PatternPart) FindRelative(this *lpg.Node) ([]*lpg.Node, error) {
	ctx := NewEvalContext(this.GetGraph())
	ctx.SetVar("this", RValue{Value: this})
	pattern, err := p.getPattern(ctx)
	if err != nil {
		return nil, err
	}

	symbols, err := BuildPatternSymbols(ctx, pattern)
	if err != nil {
		return nil, err
	}

	resultAccumulator := matchResultAccumulator{
		evalCtx: ctx,
	}
	err = pattern.Run(ctx.graph, symbols, &resultAccumulator)
	if err != nil {
		return nil, err
	}

	ret := make([]*lpg.Node, 0, len(resultAccumulator.result.Rows))
	for _, row := range resultAccumulator.result.Rows {
		t, ok := row["target"]
		if !ok {
			continue
		}
		if n, ok := t.Get().(*lpg.Node); ok {
			ret = append(ret, n)
		}
	}
	return ret, nil
}
