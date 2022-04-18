package opencypher

import (
	"strconv"
	"strings"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/cloudprivacylabs/opencypher/parser"
)

type Evaluatable interface {
	Evaluate(*EvalContext) (Value, error)
}

type RegularQuery struct {
	SingleQuery Evaluatable
	Unions      []Union
}

type Union struct {
	All         bool
	SingleQuery Evaluatable
}

type SinglePartQuery struct {
	Read   []ReadingClause
	Update []UpdatingClause
	Return *ReturnClause
}

type MultiPartQueryPart struct {
	Read   []ReadingClause
	Update []UpdatingClause
	With   WithClause
}

type WithClause struct {
	Projection ProjectionBody
	Where      Expression
}

type MultiPartQuery struct {
	Parts       []MultiPartQueryPart
	SingleQuery SinglePartQuery
}

type ReadingClause interface {
	GetResults(*EvalContext) (ResultSet, error)
}

type UpdatingClause interface {
	Evaluatable
}

type Expression interface {
	Evaluatable
}

type Unwind struct {
	Expr Expression
	As   Variable
}

type OrExpression struct {
	Parts []Evaluatable
}
type XorExpression struct {
	Parts []Evaluatable
}
type AndExpression struct {
	Parts []Evaluatable
}

type NotExpression struct {
	Part ComparisonExpression
}
type ComparisonExpression struct {
	First  Expression
	Second []PartialComparisonExpression
}

type PartialComparisonExpression struct {
	Op   string
	Expr AddOrSubtractExpression
}

type AddOrSubtractExpression struct {
	Add []MultiplyDivideModuloExpression
	Sub []MultiplyDivideModuloExpression

	constValue *Value
}

type MultiplyDivideModuloExpression struct {
	Parts []MultiplyDivideModuloExpressionPart

	constValue *Value
}

type MultiplyDivideModuloExpressionPart struct {
	// For the first element of parts, Op=0
	Op   rune
	Expr PowerOfExpression
}

type PowerOfExpression struct {
	Parts []UnaryAddOrSubtractExpression

	constValue *Value
}

type UnaryAddOrSubtractExpression struct {
	Neg  bool
	Expr StringListNullOperatorExpression

	constValue *Value
}

type StringListNullOperatorExpression struct {
	PropertyOrLabels PropertyOrLabelsExpression
	Parts            []StringListNullOperatorExpressionPart
}

type StringListNullOperatorExpressionPart struct {
	String    *StringOperatorExpression
	ListIn    Expression
	ListIndex Expression
	ListRange *ListRangeExpression
	IsNull    *bool
}

type ListRangeExpression struct {
	First  Expression
	Second Expression
}

type StringOperatorExpression struct {
	Operator string
	Expr     Expression
}

type PropertyOrLabelsExpression struct {
	Atom           Atom
	PropertyLookup []SchemaName
	NodeLabels     *NodeLabels
}

type RelationshipTypes struct {
	Rel []SchemaName
}

type ReturnClause struct {
	Projection ProjectionBody
}

type ProjectionBody struct {
	Distinct bool
	Items    ProjectionItems
	Order    *Order
	Skip     Expression
	Limit    Expression
}

type ProjectionItems struct {
	All   bool
	Items []ProjectionItem
}

type ProjectionItem struct {
	Var  *Variable
	Expr Expression
}

type Order struct {
	Items []SortItem
}

type SortItem struct {
	Asc  bool
	Expr Expression
}

type SchemaName struct {
	*SymbolicName
	*ReservedWord
}

func (s SchemaName) String() string {
	if s.SymbolicName != nil {
		return string(*s.SymbolicName)
	}
	if s.ReservedWord != nil {
		return string(*s.ReservedWord)
	}
	return ""
}

// oC_Match
//      :  ( OPTIONAL SP )? MATCH SP? oC_Pattern ( SP? oC_Where )? ;

type Match struct {
	Optional bool
	Pattern  Pattern
	Where    Expression
}
type NodeLabels []SchemaName

type SymbolicName string
type ReservedWord string

type Variable SymbolicName

type Pattern struct {
	Parts []PatternPart
}

type PatternPart struct {
	Var   *Variable
	Start NodePattern
	Path  []PatternChain
}

type NodePattern struct {
	Var        *Variable
	Labels     *NodeLabels
	Properties *Properties
}

type PatternChain struct {
	Rel  RelationshipPattern
	Node NodePattern
}

type RelationshipPattern struct {
	Backwards  bool
	Var        *Variable
	RelTypes   *RelationshipTypes
	Range      *RangeLiteral
	Properties *Properties
}

type Parameter string

type Properties struct {
	Map   *MapLiteral
	Param *Parameter
}

type CountAtom struct{}

type FilterAtom struct {
	Op     string
	Filter FilterExpression
}

type Atom interface {
	Evaluatable
}

type FilterExpression struct {
	Variable Variable
	InExpr   Expression
	Where    Expression
}

type Filter struct {
	// TODO: Derived from filter expression
}

type RelationshipsPattern struct {
	Start NodePattern
	Chain []PatternChain
}

type ListComprehension struct {
	Filter FilterExpression
	Expr   Expression
}

type FunctionInvocation struct {
	Name     []SymbolicName
	Distinct bool
	Args     []Expression

	function Function
	args     []Evaluatable
}

type PatternComprehension struct {
	Var   *Variable
	Rel   RelationshipsPattern
	Where Expression
	Expr  Expression
}

type Case struct {
	Test         Expression
	Alternatives []CaseAlternative
	Default      Expression
}

type CaseAlternative struct {
	When Expression
	Then Expression
}

type NullLiteral struct{}
type StringLiteral string
type DoubleLiteral float64
type IntLiteral int
type BooleanLiteral bool

type RangeLiteral struct {
	From, To *IntLiteral
}

type ListLiteral struct {
	Values []Expression

	constValue *Value
}

type MapLiteral struct {
	KeyValues []MapKeyValue

	constValue *Value
}

type MapKeyValue struct {
	Key   string
	Value Expression
}

func oC_Cypher(ctx *parser.OC_CypherContext) Evaluatable {
	return oC_Statement(ctx.OC_Statement().(*parser.OC_StatementContext))
}

func oC_Statement(ctx *parser.OC_StatementContext) Evaluatable {
	return oC_Query(ctx.OC_Query().(*parser.OC_QueryContext))
}

func oC_Query(ctx *parser.OC_QueryContext) Evaluatable {
	if x := ctx.OC_RegularQuery(); x != nil {
		return oC_RegularQuery(x.(*parser.OC_RegularQueryContext))
	}
	return oC_StandaloneCall(ctx.OC_StandaloneCall().(*parser.OC_StandaloneCallContext))
}

func oC_RegularQuery(ctx *parser.OC_RegularQueryContext) RegularQuery {
	ret := RegularQuery{
		SingleQuery: oC_SingleQuery(ctx.OC_SingleQuery().(*parser.OC_SingleQueryContext)),
	}
	for _, u := range ctx.AllOC_Union() {
		ret.Unions = append(ret.Unions, oC_Union(u.(*parser.OC_UnionContext)))
	}
	return ret
}

func oC_Union(ctx *parser.OC_UnionContext) Union {
	return Union{
		All:         ctx.ALL() != nil,
		SingleQuery: oC_SingleQuery(ctx.OC_SingleQuery().(*parser.OC_SingleQueryContext)),
	}
}

func oC_SingleQuery(ctx *parser.OC_SingleQueryContext) Evaluatable {
	if x := ctx.OC_SinglePartQuery(); x != nil {
		return oC_SinglePartQuery(x.(*parser.OC_SinglePartQueryContext))
	}
	return oC_MultiPartQuery(ctx.OC_MultiPartQuery().(*parser.OC_MultiPartQueryContext))
}

func oC_SinglePartQuery(ctx *parser.OC_SinglePartQueryContext) SinglePartQuery {
	ret := SinglePartQuery{}
	for _, r := range ctx.AllOC_ReadingClause() {
		ret.Read = append(ret.Read, oC_ReadingClause(r.(*parser.OC_ReadingClauseContext)))
	}
	for _, u := range ctx.AllOC_UpdatingClause() {
		ret.Update = append(ret.Update, oC_UpdatingClause(u.(*parser.OC_UpdatingClauseContext)))
	}
	if x := ctx.OC_Return(); x != nil {
		r := oC_Return(x.(*parser.OC_ReturnContext))
		ret.Return = &r
	}
	return ret
}

//oC_MultiPartQuery
//              :  ( ( oC_ReadingClause SP? )* ( oC_UpdatingClause SP? )* oC_With SP? )+ oC_SinglePartQuery ;
func oC_MultiPartQuery(ctx *parser.OC_MultiPartQueryContext) MultiPartQuery {
	ret := MultiPartQuery{Parts: []MultiPartQueryPart{}}
	count := ctx.GetChildCount()
	lastIsFull := true
	for child := 0; child < count; child++ {
		ch := ctx.GetChild(child)
		lastPart := &ret.Parts[len(ret.Parts)-1]
		switch expr := ch.(type) {
		case *parser.OC_ReadingClauseContext:
			lastPart.Read = append(lastPart.Read, oC_ReadingClause(expr))
			lastIsFull = true
		case *parser.OC_UpdatingClauseContext:
			lastPart.Update = append(lastPart.Update, oC_UpdatingClause(expr))
			lastIsFull = true
		case *parser.OC_WithContext:
			lastPart.With = oC_With(expr)
			ret.Parts = append(ret.Parts, MultiPartQueryPart{})
			lastIsFull = false
		case *parser.OC_SinglePartQueryContext:
			ret.SingleQuery = oC_SinglePartQuery(expr)
		}
	}
	if !lastIsFull {
		ret.Parts = ret.Parts[:len(ret.Parts)-1]
	}
	return ret
}

func oC_With(ctx *parser.OC_WithContext) WithClause {
	ret := WithClause{
		Projection: oC_ProjectionBody(ctx.OC_ProjectionBody().(*parser.OC_ProjectionBodyContext)),
	}
	if w := ctx.OC_Where(); w != nil {
		ret.Where = oC_Where(w.(*parser.OC_WhereContext))
	}
	return ret
}

func oC_ReadingClause(ctx *parser.OC_ReadingClauseContext) ReadingClause {
	if match := ctx.OC_Match(); match != nil {
		return oC_Match(match.(*parser.OC_MatchContext))
	}
	if unwind := ctx.OC_Unwind(); unwind != nil {
		return oC_Unwind(unwind.(*parser.OC_UnwindContext))
	}
	return oC_InQueryCall(ctx.OC_InQueryCall().(*parser.OC_InQueryCallContext))
}

func oC_UpdatingClause(ctx *parser.OC_UpdatingClauseContext) UpdatingClause {
	if create := ctx.OC_Create(); create != nil {
		return oC_Create(create.(*parser.OC_CreateContext))
	}
	if merge := ctx.OC_Merge(); merge != nil {
		return oC_Merge(merge.(*parser.OC_MergeContext))
	}
	if del := ctx.OC_Delete(); del != nil {
		return oC_Delete(del.(*parser.OC_DeleteContext))
	}
	if set := ctx.OC_Set(); set != nil {
		return oC_Set(set.(*parser.OC_SetContext))
	}
	return oC_Remove(ctx.OC_Remove().(*parser.OC_RemoveContext))
}

func oC_Return(ctx *parser.OC_ReturnContext) ReturnClause {
	return ReturnClause{
		Projection: oC_ProjectionBody(ctx.OC_ProjectionBody().(*parser.OC_ProjectionBodyContext)),
	}
}

func oC_ProjectionBody(ctx *parser.OC_ProjectionBodyContext) ProjectionBody {
	ret := ProjectionBody{
		Distinct: ctx.DISTINCT() != nil,
		Items:    oC_ProjectionItems(ctx.OC_ProjectionItems().(*parser.OC_ProjectionItemsContext)),
	}
	if o := ctx.OC_Order(); o != nil {
		x := oC_Order(o.(*parser.OC_OrderContext))
		ret.Order = &x
	}
	if s := ctx.OC_Skip(); s != nil {
		ret.Skip = oC_Skip(s.(*parser.OC_SkipContext))
	}
	if l := ctx.OC_Limit(); l != nil {
		ret.Limit = oC_Limit(l.(*parser.OC_LimitContext))
	}
	return ret
}

func oC_ProjectionItems(ctx *parser.OC_ProjectionItemsContext) ProjectionItems {
	ret := ProjectionItems{}
	if item, ok := ctx.GetChild(0).(antlr.TerminalNode); ok {
		if item.GetText() == "*" {
			ret.All = true
		}
	}
	for _, x := range ctx.AllOC_ProjectionItem() {
		ret.Items = append(ret.Items, oC_ProjectionItem(x.(*parser.OC_ProjectionItemContext)))
	}
	return ret
}

func oC_ProjectionItem(ctx *parser.OC_ProjectionItemContext) ProjectionItem {
	ret := ProjectionItem{
		Expr: oC_Expression(ctx.OC_Expression().(*parser.OC_ExpressionContext)),
	}
	if v := ctx.OC_Variable(); v != nil {
		x := oC_Variable(v.(*parser.OC_VariableContext))
		ret.Var = &x
	}
	return ret
}

func oC_Order(ctx *parser.OC_OrderContext) Order {
	ret := Order{}
	for _, x := range ctx.AllOC_SortItem() {
		ret.Items = append(ret.Items, oC_SortItem(x.(*parser.OC_SortItemContext)))
	}
	return ret
}

func oC_SortItem(ctx *parser.OC_SortItemContext) SortItem {
	ret := SortItem{
		Expr: oC_Expression(ctx.OC_Expression().(*parser.OC_ExpressionContext)),
		Asc:  ctx.ASCENDING() != nil || ctx.ASC() != nil,
	}
	return ret
}

func oC_Skip(ctx *parser.OC_SkipContext) Expression {
	return oC_Expression(ctx.OC_Expression().(*parser.OC_ExpressionContext))
}

func oC_Limit(ctx *parser.OC_LimitContext) Expression {
	return oC_Expression(ctx.OC_Expression().(*parser.OC_ExpressionContext))
}

func oC_Match(ctx *parser.OC_MatchContext) Match {
	ret := Match{
		Optional: ctx.OPTIONAL() != nil,
		Pattern:  oC_Pattern(ctx.OC_Pattern().(*parser.OC_PatternContext)),
	}
	if w := ctx.OC_Where(); w != nil {
		ret.Where = oC_Where(w.(*parser.OC_WhereContext))
	}
	return ret
}

func oC_Where(ctx *parser.OC_WhereContext) Expression {
	return oC_Expression(ctx.OC_Expression().(*parser.OC_ExpressionContext))
}

func oC_Expression(ctx *parser.OC_ExpressionContext) Expression {
	return oC_OrExpression(ctx.OC_OrExpression().(*parser.OC_OrExpressionContext))
}

func oC_OrExpression(ctx *parser.OC_OrExpressionContext) Expression {
	ret := OrExpression{}
	for _, x := range ctx.AllOC_XorExpression() {
		ret.Parts = append(ret.Parts, oC_XorExpression(x.(*parser.OC_XorExpressionContext)))
	}
	return ret
}

func oC_XorExpression(ctx *parser.OC_XorExpressionContext) Expression {
	ret := XorExpression{}
	for _, x := range ctx.AllOC_AndExpression() {
		ret.Parts = append(ret.Parts, oC_AndExpression(x.(*parser.OC_AndExpressionContext)))
	}
	return ret
}

func oC_AndExpression(ctx *parser.OC_AndExpressionContext) Expression {
	ret := AndExpression{}
	for _, x := range ctx.AllOC_NotExpression() {
		ret.Parts = append(ret.Parts, oC_NotExpression(x.(*parser.OC_NotExpressionContext)))
	}
	return ret
}

func oC_NotExpression(ctx *parser.OC_NotExpressionContext) Expression {
	if len(ctx.AllNOT())%2 == 1 {
		return NotExpression{
			Part: oC_ComparisonExpression(ctx.OC_ComparisonExpression().(*parser.OC_ComparisonExpressionContext)),
		}
	}
	return oC_ComparisonExpression(ctx.OC_ComparisonExpression().(*parser.OC_ComparisonExpressionContext))
}

func oC_ComparisonExpression(ctx *parser.OC_ComparisonExpressionContext) ComparisonExpression {
	ret := ComparisonExpression{
		First: oC_AddOrSubtractExpression(ctx.OC_AddOrSubtractExpression().(*parser.OC_AddOrSubtractExpressionContext)),
	}
	for _, x := range ctx.AllOC_PartialComparisonExpression() {
		ret.Second = append(ret.Second, oC_PartialComparisonExpression(x.(*parser.OC_PartialComparisonExpressionContext)))
	}
	return ret
}

// oC_AddOrSubtractExpression :
//      oC_MultiplyDivideModuloExpression (
//           ( SP? '+' SP? oC_MultiplyDivideModuloExpression ) |
//           ( SP? '-' SP? oC_MultiplyDivideModuloExpression )
//      )*
//
func oC_AddOrSubtractExpression(ctx *parser.OC_AddOrSubtractExpressionContext) *AddOrSubtractExpression {
	ret := &AddOrSubtractExpression{}
	target := &ret.Add
	count := ctx.GetChildCount()
	for child := 0; child < count; child++ {
		ch := ctx.GetChild(child)
		expr, ok := ch.(*parser.OC_MultiplyDivideModuloExpressionContext)
		if ok {
			(*target) = append(*target, oC_MultiplyDivideModuloExpression(expr))
		} else {
			term, ok := ch.(antlr.TerminalNode)
			if ok {
				if term.GetText() == "+" {
					target = &ret.Add
				} else if term.GetText() == "-" {
					target = &ret.Sub
				}
			}
		}
	}
	return ret
}

// oC_MultiplyDivideModuloExpression :
//      oC_PowerOfExpression (
//          ( SP? '*' SP? oC_PowerOfExpression ) |
//          ( SP? '/' SP? oC_PowerOfExpression ) |
//          ( SP? '%' SP? oC_PowerOfExpression ) )* ;
func oC_MultiplyDivideModuloExpression(ctx *parser.OC_MultiplyDivideModuloExpressionContext) MultiplyDivideModuloExpression {
	ret := MultiplyDivideModuloExpression{}
	count := ctx.GetChildCount()
	var lastOp rune
	for child := 0; child < count; child++ {
		ch := ctx.GetChild(child)
		expr, ok := ch.(*parser.OC_PowerOfExpressionContext)
		if ok {
			ret.Parts = append(ret.Parts, MultiplyDivideModuloExpressionPart{
				Op:   lastOp,
				Expr: oC_PowerOfExpression(expr),
			})
		} else {
			term, ok := ch.(antlr.TerminalNode)
			if ok {
				t := term.GetText()
				if t == "*" || t == "%" || t == "/" {
					lastOp = rune(t[0])
				}
			}
		}
	}
	return ret
}

// oC_PowerOfExpression :
//          oC_UnaryAddOrSubtractExpression ( SP? '^' SP? oC_UnaryAddOrSubtractExpression )* ;
func oC_PowerOfExpression(ctx *parser.OC_PowerOfExpressionContext) PowerOfExpression {
	ret := PowerOfExpression{}
	for _, x := range ctx.AllOC_UnaryAddOrSubtractExpression() {
		ret.Parts = append(ret.Parts, oC_UnaryAddOrSubtractExpression(x.(*parser.OC_UnaryAddOrSubtractExpressionContext)))
	}
	return ret
}

func oC_UnaryAddOrSubtractExpression(ctx *parser.OC_UnaryAddOrSubtractExpressionContext) UnaryAddOrSubtractExpression {
	ret := UnaryAddOrSubtractExpression{}
	for child := 0; child < ctx.GetChildCount(); child++ {
		ch := ctx.GetChild(child)
		if tok, ok := ch.(antlr.TerminalNode); ok {
			if tok.GetText() == "-" {
				ret.Neg = !ret.Neg
			}
		} else if expr, ok := ch.(*parser.OC_StringListNullOperatorExpressionContext); ok {
			ret.Expr = oC_StringListNullOperatorExpression(expr)
		}
	}
	return ret
}

func oC_StringListNullOperatorExpression(ctx *parser.OC_StringListNullOperatorExpressionContext) StringListNullOperatorExpression {
	var ret StringListNullOperatorExpression
	for child := 0; child < ctx.GetChildCount(); child++ {
		switch t := ctx.GetChild(child).(type) {
		case *parser.OC_PropertyOrLabelsExpressionContext:
			ret.PropertyOrLabels = oC_PropertyOrLabelsExpression(t)
		case *parser.OC_StringOperatorExpressionContext:
			ret.Parts = append(ret.Parts, oC_StringOperatorExpression(t))
		case *parser.OC_ListOperatorExpressionContext:
			ret.Parts = append(ret.Parts, oC_ListOperatorExpression(t))
		case *parser.OC_NullOperatorExpressionContext:
			ret.Parts = append(ret.Parts, oC_NullOperatorExpression(t))
		}
	}
	return ret
}

func oC_ListOperatorExpression(ctx *parser.OC_ListOperatorExpressionContext) StringListNullOperatorExpressionPart {
	if x := ctx.OC_PropertyOrLabelsExpression(); x != nil {
		return StringListNullOperatorExpressionPart{
			ListIn: oC_PropertyOrLabelsExpression(x.(*parser.OC_PropertyOrLabelsExpressionContext)),
		}
	}
	expr := ctx.AllOC_Expression()
	if len(expr) == 1 {
		return StringListNullOperatorExpressionPart{
			ListIndex: oC_Expression(expr[0].(*parser.OC_ExpressionContext)),
		}
	}
	return StringListNullOperatorExpressionPart{
		ListRange: &ListRangeExpression{
			First:  oC_Expression(expr[0].(*parser.OC_ExpressionContext)),
			Second: oC_Expression(expr[1].(*parser.OC_ExpressionContext)),
		}}
}

func oC_StringOperatorExpression(ctx *parser.OC_StringOperatorExpressionContext) StringListNullOperatorExpressionPart {
	ret := StringListNullOperatorExpressionPart{
		String: &StringOperatorExpression{
			Expr: oC_PropertyOrLabelsExpression(ctx.OC_PropertyOrLabelsExpression().(*parser.OC_PropertyOrLabelsExpressionContext)),
		},
	}
	if ctx.STARTS() != nil {
		ret.String.Operator = "STARTS"
	}
	if ctx.ENDS() != nil {
		ret.String.Operator = "ENDS"
	}
	if ctx.CONTAINS() != nil {
		ret.String.Operator = "CONTAINS"
	}
	return ret
}

func oC_NullOperatorExpression(ctx *parser.OC_NullOperatorExpressionContext) StringListNullOperatorExpressionPart {
	val := ctx.NOT() == nil
	return StringListNullOperatorExpressionPart{
		IsNull: &val,
	}
}

func oC_PropertyOrLabelsExpression(ctx *parser.OC_PropertyOrLabelsExpressionContext) PropertyOrLabelsExpression {
	ret := PropertyOrLabelsExpression{
		Atom: oC_Atom(ctx.OC_Atom().(*parser.OC_AtomContext)),
	}
	for _, x := range ctx.AllOC_PropertyLookup() {
		ret.PropertyLookup = append(ret.PropertyLookup, oC_PropertyLookup(x.(*parser.OC_PropertyLookupContext)))
	}
	if l := ctx.OC_NodeLabels(); l != nil {
		x := oC_NodeLabels(l.(*parser.OC_NodeLabelsContext))
		ret.NodeLabels = &x
	}
	return ret
}

func oC_PartialComparisonExpression(ctx *parser.OC_PartialComparisonExpressionContext) PartialComparisonExpression {
	ret := PartialComparisonExpression{Expr: *oC_AddOrSubtractExpression(ctx.OC_AddOrSubtractExpression().(*parser.OC_AddOrSubtractExpressionContext))}
	for i := 0; i < ctx.GetChildCount(); i++ {
		if tok, ok := ctx.GetChild(i).(antlr.TerminalNode); ok {
			t := tok.GetText()
			if t == "=" || t == "<>" || t == "<" || t == ">" || t == "<=" || t == "=>" {
				ret.Op = t
			}
		}
	}
	return ret
}

func oC_PropertyLookup(ctx *parser.OC_PropertyLookupContext) SchemaName {
	return oC_PropertyKeyName(ctx.OC_PropertyKeyName().(*parser.OC_PropertyKeyNameContext))
}

func oC_PropertyKeyName(ctx *parser.OC_PropertyKeyNameContext) SchemaName {
	return oC_SchemaName(ctx.OC_SchemaName().(*parser.OC_SchemaNameContext))
}

func oC_SchemaName(ctx *parser.OC_SchemaNameContext) SchemaName {
	if x := ctx.OC_SymbolicName(); x != nil {
		name := oC_SymbolicName(ctx.OC_SymbolicName().(*parser.OC_SymbolicNameContext))
		return SchemaName{SymbolicName: &name}
	}
	name := oC_ReservedWord(ctx.OC_ReservedWord().(*parser.OC_ReservedWordContext))
	return SchemaName{ReservedWord: &name}
}

func oC_SymbolicName(ctx *parser.OC_SymbolicNameContext) SymbolicName {
	if node := ctx.EscapedSymbolicName(); node != nil {
		text := node.GetText()
		return SymbolicName(text[1 : len(text)-1])
	}
	return SymbolicName(ctx.GetText())
}

func oC_ReservedWord(ctx *parser.OC_ReservedWordContext) ReservedWord {
	return ReservedWord(strings.ToUpper(ctx.GetText()))
}

func oC_NodeLabels(ctx *parser.OC_NodeLabelsContext) NodeLabels {
	children := ctx.AllOC_NodeLabel()
	ret := make([]SchemaName, 0, len(children))
	for _, x := range children {
		ret = append(ret, oC_SchemaName(x.(*parser.OC_NodeLabelContext).OC_LabelName().(*parser.OC_LabelNameContext).OC_SchemaName().(*parser.OC_SchemaNameContext)))
	}
	return ret
}

func oC_Pattern(ctx *parser.OC_PatternContext) Pattern {
	ret := Pattern{}
	for _, x := range ctx.AllOC_PatternPart() {
		ret.Parts = append(ret.Parts, oC_PatternPart(x.(*parser.OC_PatternPartContext)))
	}
	return ret
}

func oC_PatternPart(ctx *parser.OC_PatternPartContext) PatternPart {
	ret := PatternPart{}
	if v := ctx.OC_Variable(); v != nil {
		vr := oC_Variable(v.(*parser.OC_VariableContext))
		ret.Var = &vr
	}
	ret.Start, ret.Path = oC_AnonymousPatternPart(ctx.OC_AnonymousPatternPart().(*parser.OC_AnonymousPatternPartContext))
	return ret
}

func oC_AnonymousPatternPart(ctx *parser.OC_AnonymousPatternPartContext) (NodePattern, []PatternChain) {
	return oC_PatternElement(ctx.OC_PatternElement().(*parser.OC_PatternElementContext))
}

func oC_PatternElement(ctx *parser.OC_PatternElementContext) (NodePattern, []PatternChain) {
	if p := ctx.OC_PatternElement(); p != nil {
		return oC_PatternElement(p.(*parser.OC_PatternElementContext))
	}
	np := oC_NodePattern(ctx.OC_NodePattern().(*parser.OC_NodePatternContext))
	chain := make([]PatternChain, 0)
	for _, x := range ctx.AllOC_PatternElementChain() {
		chain = append(chain, oC_PatternElementChain(x.(*parser.OC_PatternElementChainContext)))
	}
	return np, chain
}

func oC_NodePattern(ctx *parser.OC_NodePatternContext) NodePattern {
	ret := NodePattern{}
	if x := ctx.OC_Variable(); x != nil {
		v := oC_Variable(x.(*parser.OC_VariableContext))
		ret.Var = &v
	}
	if x := ctx.OC_NodeLabels(); x != nil {
		l := oC_NodeLabels(x.(*parser.OC_NodeLabelsContext))
		ret.Labels = &l
	}
	if x := ctx.OC_Properties(); x != nil {
		o := oC_Properties(x.(*parser.OC_PropertiesContext))
		ret.Properties = &o
	}
	return ret
}

func oC_PatternElementChain(ctx *parser.OC_PatternElementChainContext) PatternChain {
	return PatternChain{
		Rel:  oC_RelationshipPattern(ctx.OC_RelationshipPattern().(*parser.OC_RelationshipPatternContext)),
		Node: oC_NodePattern(ctx.OC_NodePattern().(*parser.OC_NodePatternContext)),
	}
}

func oC_RelationshipPattern(ctx *parser.OC_RelationshipPatternContext) RelationshipPattern {
	ret := RelationshipPattern{
		Backwards: ctx.OC_LeftArrowHead() != nil,
	}
	if x := ctx.OC_RelationshipDetail(); x != nil {
		detctx := x.(*parser.OC_RelationshipDetailContext)
		if x := detctx.OC_Variable(); x != nil {
			v := oC_Variable(x.(*parser.OC_VariableContext))
			ret.Var = &v
		}
		if x := detctx.OC_RelationshipTypes(); x != nil {
			r := oC_RelationshipTypes(x.(*parser.OC_RelationshipTypesContext))
			ret.RelTypes = &r
		}
		if x := detctx.OC_RangeLiteral(); x != nil {
			r := oC_RangeLiteral(x.(*parser.OC_RangeLiteralContext))
			ret.Range = &r
		}
		if x := detctx.OC_Properties(); x != nil {
			p := oC_Properties(x.(*parser.OC_PropertiesContext))
			ret.Properties = &p
		}
	}
	return ret
}

func oC_Variable(ctx *parser.OC_VariableContext) Variable {
	return Variable(oC_SymbolicName(ctx.OC_SymbolicName().(*parser.OC_SymbolicNameContext)))
}

func oC_RelationshipTypes(ctx *parser.OC_RelationshipTypesContext) RelationshipTypes {
	ret := RelationshipTypes{}
	for _, x := range ctx.AllOC_RelTypeName() {
		ret.Rel = append(ret.Rel, oC_RelTypeName(x.(*parser.OC_RelTypeNameContext)))
	}
	return ret
}

func oC_RelTypeName(ctx *parser.OC_RelTypeNameContext) SchemaName {
	return oC_SchemaName(ctx.OC_SchemaName().(*parser.OC_SchemaNameContext))
}

func oC_Properties(ctx *parser.OC_PropertiesContext) Properties {
	ret := Properties{}
	if x := ctx.OC_MapLiteral(); x != nil {
		ret.Map = oC_MapLiteral(x.(*parser.OC_MapLiteralContext))
		return ret
	}
	c := oC_Parameter(ctx.OC_Parameter().(*parser.OC_ParameterContext))
	ret.Param = &c
	return ret
}

func oC_Parameter(ctx *parser.OC_ParameterContext) Parameter {
	var ret Parameter
	if x := ctx.OC_SymbolicName(); x != nil {
		c := oC_SymbolicName(x.(*parser.OC_SymbolicNameContext))
		ret = Parameter("$" + string(c))
		return ret
	}
	ret = Parameter("$" + ctx.DecimalInteger().GetText())
	return ret
}

func DecimalInteger(ctx antlr.TerminalNode) IntLiteral {
	i, err := strconv.Atoi(ctx.GetText())
	if err != nil {
		panic(err)
	}
	return IntLiteral(i)
}

func oC_RangeLiteral(ctx *parser.OC_RangeLiteralContext) RangeLiteral {
	ret := RangeLiteral{}
	values := make([]IntLiteral, 0, 2)
	rangeIndex := -1
	for i := 0; i < ctx.GetChildCount(); i++ {
		ch := ctx.GetChild(i)
		if tok, ok := ch.(antlr.TerminalNode); ok {
			if tok.GetText() == ".." {
				rangeIndex = len(values)
			}
		} else if i, ok := ch.(*parser.OC_IntegerLiteralContext); ok {
			values = append(values, oC_IntegerLiteral(i))
		}
	}

	switch len(values) {
	case 2:
		ret.From = &values[0]
		ret.To = &values[1]
	case 1:
		switch rangeIndex {
		case 0: // ..value
			ret.To = &values[0]
		case 1: // value..
			ret.From = &values[0]
		case -1: // value
			ret.From = &values[0]
			ret.To = &values[0]
		}
	}
	return ret
}

func oC_IntegerLiteral(ctx *parser.OC_IntegerLiteralContext) IntLiteral {
	if x := ctx.HexInteger(); x != nil {
		t := x.GetText()
		val, err := strconv.ParseInt(string(t[2:]), 16, 64)
		if err != nil {
			panic(err)
		}
		return IntLiteral(val)
	}
	if x := ctx.OctalInteger(); x != nil {
		t := x.GetText()
		val, err := strconv.ParseInt(t, 8, 64)
		if err != nil {
			panic(err)
		}
		return IntLiteral(val)
	}
	return DecimalInteger(ctx.DecimalInteger())
}

func oC_BooleanLiteral(ctx *parser.OC_BooleanLiteralContext) BooleanLiteral {
	if ctx.TRUE() != nil {
		return BooleanLiteral(true)
	}
	return BooleanLiteral(false)
}

func oC_ListLiteral(ctx *parser.OC_ListLiteralContext) *ListLiteral {
	ret := &ListLiteral{}
	for _, x := range ctx.AllOC_Expression() {
		ret.Values = append(ret.Values, oC_Expression(x.(*parser.OC_ExpressionContext)))
	}
	return ret
}

func oC_MapLiteral(ctx *parser.OC_MapLiteralContext) *MapLiteral {
	ret := &MapLiteral{}
	var keyName SchemaName
	for i := 0; i < ctx.GetChildCount(); i++ {
		switch c := ctx.GetChild(i).(type) {
		case *parser.OC_PropertyKeyNameContext:
			keyName = oC_SchemaName(c.OC_SchemaName().(*parser.OC_SchemaNameContext))
		case *parser.OC_ExpressionContext:
			value := MapKeyValue{
				Key:   keyName.String(),
				Value: oC_Expression(c),
			}
			ret.KeyValues = append(ret.KeyValues, value)
		}
	}
	return ret
}

func oC_Atom(ctx *parser.OC_AtomContext) Atom {
	if lit := ctx.OC_Literal(); lit != nil {
		return oC_Literal(lit.(*parser.OC_LiteralContext))
	}
	if param := ctx.OC_Parameter(); param != nil {
		return oC_Parameter(param.(*parser.OC_ParameterContext))
	}
	if casee := ctx.OC_CaseExpression(); casee != nil {
		return oC_CaseExpression(casee.(*parser.OC_CaseExpressionContext))
	}
	if ctx.COUNT() != nil {
		return CountAtom{}
	}

	if l := ctx.OC_ListComprehension(); l != nil {
		return oC_ListComprehension(l.(*parser.OC_ListComprehensionContext))
	}
	if p := ctx.OC_PatternComprehension(); p != nil {
		return oC_PatternComprehension(p.(*parser.OC_PatternComprehensionContext))
	}
	if ctx.ALL() != nil {
		return FilterAtom{
			Op:     "ALL",
			Filter: oC_FilterExpression(ctx.OC_FilterExpression().(*parser.OC_FilterExpressionContext)),
		}
	}
	if ctx.ANY() != nil {
		return FilterAtom{
			Op:     "ANY",
			Filter: oC_FilterExpression(ctx.OC_FilterExpression().(*parser.OC_FilterExpressionContext)),
		}
	}
	if ctx.NONE() != nil {
		return FilterAtom{
			Op:     "NONE",
			Filter: oC_FilterExpression(ctx.OC_FilterExpression().(*parser.OC_FilterExpressionContext)),
		}
	}
	if ctx.SINGLE() != nil {
		return FilterAtom{
			Op:     "SINGLE",
			Filter: oC_FilterExpression(ctx.OC_FilterExpression().(*parser.OC_FilterExpressionContext)),
		}
	}
	if r := ctx.OC_RelationshipsPattern(); r != nil {
		return oC_RelationshipsPattern(r.(*parser.OC_RelationshipsPatternContext))
	}
	if p := ctx.OC_ParenthesizedExpression(); p != nil {
		return oC_Expression(p.(*parser.OC_ParenthesizedExpressionContext).OC_Expression().(*parser.OC_ExpressionContext))
	}
	if f := ctx.OC_FunctionInvocation(); f != nil {
		return oC_FunctionInvocation(f.(*parser.OC_FunctionInvocationContext))
	}
	return oC_Variable(ctx.OC_Variable().(*parser.OC_VariableContext))
}

func oC_FilterExpression(ctx *parser.OC_FilterExpressionContext) FilterExpression {
	idInColl := ctx.OC_IdInColl().(*parser.OC_IdInCollContext)
	ret := FilterExpression{
		Variable: oC_Variable(idInColl.OC_Variable().(*parser.OC_VariableContext)),
		InExpr:   oC_Expression(idInColl.OC_Expression().(*parser.OC_ExpressionContext)),
	}
	if w := ctx.OC_Where(); w != nil {
		ret.Where = oC_Where(w.(*parser.OC_WhereContext))
	}

	return ret
}

//oC_RelationshipsPattern   :  oC_NodePattern ( SP? oC_PatternElementChain )+ ;
// oC_PatternElementChain :  oC_RelationshipPattern SP? oC_NodePattern ;
func oC_RelationshipsPattern(ctx *parser.OC_RelationshipsPatternContext) RelationshipsPattern {
	ret := RelationshipsPattern{
		Start: oC_NodePattern(ctx.OC_NodePattern().(*parser.OC_NodePatternContext)),
	}
	for _, element := range ctx.AllOC_PatternElementChain() {
		ret.Chain = append(ret.Chain, oC_PatternElementChain(element.(*parser.OC_PatternElementChainContext)))
	}
	return ret
}

func oC_ListComprehension(ctx *parser.OC_ListComprehensionContext) ListComprehension {
	ret := ListComprehension{
		Filter: oC_FilterExpression(ctx.OC_FilterExpression().(*parser.OC_FilterExpressionContext)),
	}
	if x := ctx.OC_Expression(); x != nil {
		ret.Expr = oC_Expression(x.(*parser.OC_ExpressionContext))
	}
	return ret
}

func oC_FunctionInvocation(ctx *parser.OC_FunctionInvocationContext) *FunctionInvocation {
	ret := &FunctionInvocation{
		Name:     oC_FunctionName(ctx.OC_FunctionName().(*parser.OC_FunctionNameContext)),
		Distinct: ctx.DISTINCT() != nil,
	}
	for _, x := range ctx.AllOC_Expression() {
		ret.Args = append(ret.Args, oC_Expression(x.(*parser.OC_ExpressionContext)))
	}
	return ret
}

func oC_FunctionName(ctx *parser.OC_FunctionNameContext) []SymbolicName {
	if ctx.EXISTS() != nil {
		return []SymbolicName{"EXISTS"}
	}
	ns := oC_Namespace(ctx.OC_Namespace().(*parser.OC_NamespaceContext))
	return append(ns, oC_SymbolicName(ctx.OC_SymbolicName().(*parser.OC_SymbolicNameContext)))
}

func oC_Namespace(ctx *parser.OC_NamespaceContext) []SymbolicName {
	ret := make([]SymbolicName, 0)
	for _, x := range ctx.AllOC_SymbolicName() {
		ret = append(ret, oC_SymbolicName(x.(*parser.OC_SymbolicNameContext)))
	}
	return ret
}

func oC_PatternComprehension(ctx *parser.OC_PatternComprehensionContext) PatternComprehension {
	ret := PatternComprehension{
		Rel: oC_RelationshipsPattern(ctx.OC_RelationshipsPattern().(*parser.OC_RelationshipsPatternContext)),
	}
	if v := ctx.OC_Variable(); v != nil {
		x := oC_Variable(v.(*parser.OC_VariableContext))
		ret.Var = &x
	}
	expr := ctx.AllOC_Expression()
	if len(expr) == 1 {
		ret.Expr = oC_Expression(expr[0].(*parser.OC_ExpressionContext))
	} else {
		ret.Where = oC_Expression(expr[0].(*parser.OC_ExpressionContext))
		ret.Expr = oC_Expression(expr[1].(*parser.OC_ExpressionContext))
	}
	return ret
}

func oC_CaseExpression(ctx *parser.OC_CaseExpressionContext) Case {
	ret := Case{}
	for _, alt := range ctx.AllOC_CaseAlternatives() {
		ret.Alternatives = append(ret.Alternatives, oC_CaseAlternatives(alt.(*parser.OC_CaseAlternativesContext)))
	}
	elseSeen := false
	for i := 0; i < ctx.GetChildCount(); i++ {
		ch := ctx.GetChild(i)
		if term, ok := ch.(antlr.TerminalNode); ok {
			if strings.ToUpper(term.GetText()) == "ELSE" {
				elseSeen = true
			}
		} else if expr, ok := ch.(*parser.OC_ExpressionContext); ok {
			if elseSeen {
				ret.Default = oC_Expression(expr)
			} else {
				ret.Test = oC_Expression(expr)
			}
		}
	}
	return ret
}

func oC_CaseAlternatives(ctx *parser.OC_CaseAlternativesContext) CaseAlternative {
	expr := ctx.AllOC_Expression()
	return CaseAlternative{
		When: oC_Expression(expr[0].(*parser.OC_ExpressionContext)),
		Then: oC_Expression(expr[1].(*parser.OC_ExpressionContext)),
	}
}

func oC_Literal(ctx *parser.OC_LiteralContext) Evaluatable {
	if ctx.NULL() != nil {
		return NullLiteral{}
	}
	if n := ctx.OC_NumberLiteral(); n != nil {
		return oC_NumberLiteral(n.(*parser.OC_NumberLiteralContext))
	}
	if s := ctx.StringLiteral(); s != nil {
		text := s.GetText()
		return StringLiteral(text[1 : len(text)-1])
	}
	if b := ctx.OC_BooleanLiteral(); b != nil {
		return oC_BooleanLiteral(b.(*parser.OC_BooleanLiteralContext))
	}
	if m := ctx.OC_MapLiteral(); m != nil {
		return oC_MapLiteral(m.(*parser.OC_MapLiteralContext))
	}
	return oC_ListLiteral(ctx.OC_ListLiteral().(*parser.OC_ListLiteralContext))
}

func oC_NumberLiteral(ctx *parser.OC_NumberLiteralContext) Evaluatable {
	if d := ctx.OC_DoubleLiteral(); d != nil {
		return oC_DoubleLiteral(d.(*parser.OC_DoubleLiteralContext))
	}
	return oC_IntegerLiteral(ctx.OC_IntegerLiteral().(*parser.OC_IntegerLiteralContext))
}

func oC_DoubleLiteral(ctx *parser.OC_DoubleLiteralContext) DoubleLiteral {
	v, err := strconv.ParseFloat(ctx.GetText(), 64)
	if err != nil {
		panic(err)
	}
	return DoubleLiteral(v)
}

func oC_Unwind(ctx *parser.OC_UnwindContext) Unwind {
	return Unwind{
		Expr: oC_Expression(ctx.OC_Expression().(*parser.OC_ExpressionContext)),
		As:   oC_Variable(ctx.OC_Variable().(*parser.OC_VariableContext)),
	}
}

func oC_InQueryCall(ctx *parser.OC_InQueryCallContext) ReadingClause {
	panic("Unsupported: inQueryCall")
}

func oC_Create(ctx *parser.OC_CreateContext) Evaluatable {
	panic("Unsupported: create")
}

func oC_Merge(ctx *parser.OC_MergeContext) Evaluatable {
	panic("Unsupported: merge")
}

func oC_Delete(ctx *parser.OC_DeleteContext) Evaluatable {
	panic("Unsupported: delete")
}

func oC_Set(ctx *parser.OC_SetContext) Evaluatable {
	panic("Unsupported: set")
}

func oC_Remove(ctx *parser.OC_RemoveContext) Evaluatable {
	panic("Unsupported: remove")
}

func oC_StandaloneCall(ctx *parser.OC_StandaloneCallContext) Evaluatable {
	panic("Unsupported: standaloneCall")
}
