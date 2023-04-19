package opencypher

import (
	"strconv"
	"strings"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/cloudprivacylabs/lpg/v2"
	"github.com/cloudprivacylabs/opencypher/parser"
)

type Evaluatable interface {
	Evaluate(*EvalContext) (Value, error)
}

type ResultPath struct {
	Result  *lpg.Path
	Symbols map[string]Value
}

type regularQuery struct {
	singleQuery Evaluatable
	unions      []union
}

type union struct {
	all         bool
	singleQuery Evaluatable
}

type singlePartQuery struct {
	read   []ReadingClause
	update []UpdatingClause
	ret    *returnClause
}

type create struct {
	pattern Pattern
}

type deleteClause struct {
	detach bool
	exprs  []Expression
}

type remove struct {
	items []removeItem
}

type removeItem struct {
	variable   *variable
	nodeLabels NodeLabels
	property   *propertyExpression
}

type set struct {
	items []setItem
}

type setItem struct {
	property   *propertyExpression
	variable   *variable
	op         string
	expression Expression
	nodeLabels NodeLabels
}

type propertyExpression struct {
	atom   atom
	lookup []schemaName
}

type multiPartQueryPart struct {
	read   []ReadingClause
	update []UpdatingClause
	with   withClause
}

type withClause struct {
	projection projectionBody
	where      Expression
}

type multiPartQuery struct {
	parts       []multiPartQueryPart
	singleQuery singlePartQuery
}

type ReadingClause interface {
	GetResults(*EvalContext) ([]ResultPath, error)
}

type UpdatingClause interface {
	Update(*EvalContext, []ResultPath) (Value, error)
	TopLevelUpdate(*EvalContext) (Value, error)
}

type Expression interface {
	Evaluatable
}

type unwind struct {
	expr Expression
	as   variable
}

type orExpression struct {
	parts []Evaluatable
}
type xorExpression struct {
	parts []Evaluatable
}
type andExpression struct {
	parts []Evaluatable
}

type notExpression struct {
	part Expression
}
type comparisonExpression struct {
	first  Expression
	second []partialComparisonExpression
}

type partialComparisonExpression struct {
	op   string
	expr Expression
}

type addOrSubtractExpression struct {
	add []Expression
	sub []Expression

	constValue Value
}

type multiplyDivideModuloExpression struct {
	parts []multiplyDivideModuloExpressionPart

	constValue Value
}

type multiplyDivideModuloExpressionPart struct {
	// For the first element of parts, Op=0
	op   rune
	expr Evaluatable
}

type powerOfExpression struct {
	parts []Evaluatable

	constValue Value
}

type unaryAddOrSubtractExpression struct {
	neg  bool
	expr stringListNullOperatorExpression

	constValue Value
}

type stringListNullOperatorExpression struct {
	propertyOrLabels propertyOrLabelsExpression
	parts            []stringListNullOperatorExpressionPart
}

type stringListNullOperatorExpressionPart struct {
	stringOp  *stringOperatorExpression
	listIn    Expression
	listIndex Expression
	listRange *listRangeExpression
	isNull    *bool
}

type listRangeExpression struct {
	first  Expression
	second Expression
}

type stringOperatorExpression struct {
	operator string
	expr     Expression
}

type propertyOrLabelsExpression struct {
	atom           atom
	propertyLookup []schemaName
	nodeLabels     *NodeLabels
}

type relationshipTypes struct {
	rel []schemaName
}

type returnClause struct {
	projection projectionBody
}

type projectionBody struct {
	distinct bool
	items    projectionItems
	order    *order
	skip     Expression
	limit    Expression
}

type projectionItems struct {
	all   bool
	items []projectionItem
}

func (p projectionItems) getProjectedNames() []string {
	ret := make([]string, 0, len(p.items))
	for i, item := range p.items {
		if item.variable != nil {
			ret = append(ret, string(*item.variable))
		} else {
			ret = append(ret, strconv.Itoa(i+1))
		}
	}
	return ret
}

type projectionItem struct {
	variable *variable
	expr     Expression
}

type order struct {
	items []sortItem
}

type sortItem struct {
	asc  bool
	expr Expression
}

type schemaName struct {
	*symbolicName
	*reservedWord
}

func (s schemaName) String() string {
	if s.symbolicName != nil {
		return string(*s.symbolicName)
	}
	if s.reservedWord != nil {
		return string(*s.reservedWord)
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
type NodeLabels []schemaName

type symbolicName string
type reservedWord string

type variable symbolicName

type merge struct {
	pattern PatternPart
	actions []mergeAction
}

const mergeActionOnMatch = 0
const mergeActionOnCreate = 1

type mergeAction struct {
	on  int
	set UpdatingClause
}

type Pattern struct {
	Parts []PatternPart
}

type PatternPart struct {
	variable *variable
	start    nodePattern
	path     []patternChain
}

type nodePattern struct {
	variable   *variable
	labels     *NodeLabels
	properties *Properties
}

type patternChain struct {
	rel  relationshipPattern
	node nodePattern
}

type relationshipPattern struct {
	toLeft     bool
	toRight    bool
	variable   *variable
	relTypes   *relationshipTypes
	rng        *rangeLiteral
	properties *Properties
}

type Parameter string

type Properties struct {
	Map   *mapLiteral
	Param *Parameter
}

type countAtom struct{}

type filterAtom struct {
	op     string
	filter filterExpression
}

type atom interface {
	Evaluatable
}

type filterExpression struct {
	variable variable
	inExpr   Expression
	where    Expression
}

type relationshipsPattern struct {
	start nodePattern
	chain []patternChain
}

type listComprehension struct {
	filter filterExpression
	expr   Expression
}

type functionInvocation struct {
	name     []symbolicName
	distinct bool
	args     []Expression

	function       *Function
	constArgs      []Evaluatable
	constArgValues []Value
}

type patternComprehension struct {
	variable *variable
	rel      relationshipsPattern
	where    Expression
	expr     Expression
}

type caseClause struct {
	test         Expression
	alternatives []caseAlternative
	def          Expression
}

type caseAlternative struct {
	when Expression
	then Expression
}

type nullLiteral struct{}
type stringLiteral string
type doubleLiteral float64
type intLiteral int
type booleanLiteral bool

type rangeLiteral struct {
	from, to *intLiteral
}

type listLiteral struct {
	values []Expression

	constValue Value
}

type mapLiteral struct {
	keyValues []mapKeyValue

	constValue Value
}

type mapKeyValue struct {
	key   string
	value Expression
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

func oC_RegularQuery(ctx *parser.OC_RegularQueryContext) regularQuery {
	ret := regularQuery{
		singleQuery: oC_SingleQuery(ctx.OC_SingleQuery().(*parser.OC_SingleQueryContext)),
	}
	for _, u := range ctx.AllOC_Union() {
		ret.unions = append(ret.unions, oC_Union(u.(*parser.OC_UnionContext)))
	}
	return ret
}

func oC_Union(ctx *parser.OC_UnionContext) union {
	return union{
		all:         ctx.ALL() != nil,
		singleQuery: oC_SingleQuery(ctx.OC_SingleQuery().(*parser.OC_SingleQueryContext)),
	}
}

func oC_SingleQuery(ctx *parser.OC_SingleQueryContext) Evaluatable {
	if x := ctx.OC_SinglePartQuery(); x != nil {
		return oC_SinglePartQuery(x.(*parser.OC_SinglePartQueryContext))
	}
	return oC_MultiPartQuery(ctx.OC_MultiPartQuery().(*parser.OC_MultiPartQueryContext))
}

func oC_SinglePartQuery(ctx *parser.OC_SinglePartQueryContext) singlePartQuery {
	ret := singlePartQuery{}
	for _, r := range ctx.AllOC_ReadingClause() {
		ret.read = append(ret.read, oC_ReadingClause(r.(*parser.OC_ReadingClauseContext)))
	}
	for _, u := range ctx.AllOC_UpdatingClause() {
		ret.update = append(ret.update, oC_UpdatingClause(u.(*parser.OC_UpdatingClauseContext)))
	}
	if x := ctx.OC_Return(); x != nil {
		r := oC_Return(x.(*parser.OC_ReturnContext))
		ret.ret = &r
	}
	return ret
}

// oC_MultiPartQuery
//
//	:  ( ( oC_ReadingClause SP? )* ( oC_UpdatingClause SP? )* oC_With SP? )+ oC_SinglePartQuery ;
func oC_MultiPartQuery(ctx *parser.OC_MultiPartQueryContext) multiPartQuery {
	ret := multiPartQuery{parts: []multiPartQueryPart{}}
	count := ctx.GetChildCount()
	lastPartIx := -1
	for child := 0; child < count; child++ {
		ch := ctx.GetChild(child)
		switch expr := ch.(type) {
		case *parser.OC_ReadingClauseContext:
			if lastPartIx == -1 {
				ret.parts = append(ret.parts, multiPartQueryPart{})
				lastPartIx = len(ret.parts) - 1
			}
			ret.parts[lastPartIx].read = append(ret.parts[lastPartIx].read, oC_ReadingClause(expr))
		case *parser.OC_UpdatingClauseContext:
			if lastPartIx == -1 {
				ret.parts = append(ret.parts, multiPartQueryPart{})
				lastPartIx = len(ret.parts) - 1
			}
			ret.parts[lastPartIx].update = append(ret.parts[lastPartIx].update, oC_UpdatingClause(expr))
		case *parser.OC_WithContext:
			if lastPartIx == -1 {
				ret.parts = append(ret.parts, multiPartQueryPart{})
				lastPartIx = len(ret.parts) - 1
			}
			ret.parts[lastPartIx].with = oC_With(expr)
			lastPartIx = -1
		case *parser.OC_SinglePartQueryContext:
			ret.singleQuery = oC_SinglePartQuery(expr)
		}
	}
	return ret
}

func oC_With(ctx *parser.OC_WithContext) withClause {
	ret := withClause{
		projection: oC_ProjectionBody(ctx.OC_ProjectionBody().(*parser.OC_ProjectionBodyContext)),
	}
	if w := ctx.OC_Where(); w != nil {
		ret.where = oC_Where(w.(*parser.OC_WhereContext))
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

func oC_Return(ctx *parser.OC_ReturnContext) returnClause {
	return returnClause{
		projection: oC_ProjectionBody(ctx.OC_ProjectionBody().(*parser.OC_ProjectionBodyContext)),
	}
}

func oC_ProjectionBody(ctx *parser.OC_ProjectionBodyContext) projectionBody {
	ret := projectionBody{
		distinct: ctx.DISTINCT() != nil,
		items:    oC_ProjectionItems(ctx.OC_ProjectionItems().(*parser.OC_ProjectionItemsContext)),
	}
	if o := ctx.OC_Order(); o != nil {
		x := oC_Order(o.(*parser.OC_OrderContext))
		ret.order = &x
	}
	if s := ctx.OC_Skip(); s != nil {
		ret.skip = oC_Skip(s.(*parser.OC_SkipContext))
	}
	if l := ctx.OC_Limit(); l != nil {
		ret.limit = oC_Limit(l.(*parser.OC_LimitContext))
	}
	return ret
}

func oC_ProjectionItems(ctx *parser.OC_ProjectionItemsContext) projectionItems {
	ret := projectionItems{}
	if item, ok := ctx.GetChild(0).(antlr.TerminalNode); ok {
		if item.GetText() == "*" {
			ret.all = true
		}
	}
	for _, x := range ctx.AllOC_ProjectionItem() {
		ret.items = append(ret.items, oC_ProjectionItem(x.(*parser.OC_ProjectionItemContext)))
	}
	return ret
}

func oC_ProjectionItem(ctx *parser.OC_ProjectionItemContext) projectionItem {
	ret := projectionItem{
		expr: oC_Expression(ctx.OC_Expression().(*parser.OC_ExpressionContext)),
	}
	if v := ctx.OC_Variable(); v != nil {
		x := oC_Variable(v.(*parser.OC_VariableContext))
		ret.variable = &x
	}
	return ret
}

func oC_Order(ctx *parser.OC_OrderContext) order {
	ret := order{}
	for _, x := range ctx.AllOC_SortItem() {
		ret.items = append(ret.items, oC_SortItem(x.(*parser.OC_SortItemContext)))
	}
	return ret
}

func oC_SortItem(ctx *parser.OC_SortItemContext) sortItem {
	ret := sortItem{
		expr: oC_Expression(ctx.OC_Expression().(*parser.OC_ExpressionContext)),
		asc:  ctx.ASCENDING() != nil || ctx.ASC() != nil,
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
	ret := orExpression{}
	for _, x := range ctx.AllOC_XorExpression() {
		ret.parts = append(ret.parts, oC_XorExpression(x.(*parser.OC_XorExpressionContext)))
	}
	if len(ret.parts) == 1 {
		return ret.parts[0]
	}
	return ret
}

func oC_XorExpression(ctx *parser.OC_XorExpressionContext) Expression {
	ret := xorExpression{}
	for _, x := range ctx.AllOC_AndExpression() {
		ret.parts = append(ret.parts, oC_AndExpression(x.(*parser.OC_AndExpressionContext)))
	}
	if len(ret.parts) == 1 {
		return ret.parts[0]
	}
	return ret
}

func oC_AndExpression(ctx *parser.OC_AndExpressionContext) Expression {
	ret := andExpression{}
	for _, x := range ctx.AllOC_NotExpression() {
		ret.parts = append(ret.parts, oC_NotExpression(x.(*parser.OC_NotExpressionContext)))
	}
	if len(ret.parts) == 1 {
		return ret.parts[0]
	}
	return ret
}

func oC_NotExpression(ctx *parser.OC_NotExpressionContext) Expression {
	if len(ctx.AllNOT())%2 == 1 {
		return notExpression{
			part: oC_ComparisonExpression(ctx.OC_ComparisonExpression().(*parser.OC_ComparisonExpressionContext)),
		}
	}
	return oC_ComparisonExpression(ctx.OC_ComparisonExpression().(*parser.OC_ComparisonExpressionContext))
}

func oC_ComparisonExpression(ctx *parser.OC_ComparisonExpressionContext) Expression {
	ret := comparisonExpression{
		first: oC_AddOrSubtractExpression(ctx.OC_AddOrSubtractExpression().(*parser.OC_AddOrSubtractExpressionContext)),
	}
	for _, x := range ctx.AllOC_PartialComparisonExpression() {
		ret.second = append(ret.second, oC_PartialComparisonExpression(x.(*parser.OC_PartialComparisonExpressionContext)))
	}
	if len(ret.second) == 0 {
		return ret.first
	}
	return ret
}

// oC_AddOrSubtractExpression :
//
//	oC_MultiplyDivideModuloExpression (
//	     ( SP? '+' SP? oC_MultiplyDivideModuloExpression ) |
//	     ( SP? '-' SP? oC_MultiplyDivideModuloExpression )
//	)*
func oC_AddOrSubtractExpression(ctx *parser.OC_AddOrSubtractExpressionContext) Expression {
	ret := &addOrSubtractExpression{}
	target := &ret.add
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
					target = &ret.add
				} else if term.GetText() == "-" {
					target = &ret.sub
				}
			}
		}
	}
	if len(ret.add) == 1 && len(ret.sub) == 0 {
		return ret.add[0]
	}
	return ret
}

// oC_MultiplyDivideModuloExpression :
//
//	oC_PowerOfExpression (
//	    ( SP? '*' SP? oC_PowerOfExpression ) |
//	    ( SP? '/' SP? oC_PowerOfExpression ) |
//	    ( SP? '%' SP? oC_PowerOfExpression ) )* ;
func oC_MultiplyDivideModuloExpression(ctx *parser.OC_MultiplyDivideModuloExpressionContext) Expression {
	ret := &multiplyDivideModuloExpression{}
	count := ctx.GetChildCount()
	var lastOp rune
	for child := 0; child < count; child++ {
		ch := ctx.GetChild(child)
		expr, ok := ch.(*parser.OC_PowerOfExpressionContext)
		if ok {
			ret.parts = append(ret.parts, multiplyDivideModuloExpressionPart{
				op:   lastOp,
				expr: oC_PowerOfExpression(expr),
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
	if len(ret.parts) == 1 {
		return ret.parts[0].expr
	}
	return ret
}

// oC_PowerOfExpression :
//
//	oC_UnaryAddOrSubtractExpression ( SP? '^' SP? oC_UnaryAddOrSubtractExpression )* ;
func oC_PowerOfExpression(ctx *parser.OC_PowerOfExpressionContext) Evaluatable {
	ret := powerOfExpression{}
	for _, x := range ctx.AllOC_UnaryAddOrSubtractExpression() {
		ret.parts = append(ret.parts, oC_UnaryAddOrSubtractExpression(x.(*parser.OC_UnaryAddOrSubtractExpressionContext)))
	}
	if len(ret.parts) == 1 {
		return ret.parts[0]
	}
	return &ret
}

func oC_UnaryAddOrSubtractExpression(ctx *parser.OC_UnaryAddOrSubtractExpressionContext) Evaluatable {
	ret := unaryAddOrSubtractExpression{}
	for child := 0; child < ctx.GetChildCount(); child++ {
		ch := ctx.GetChild(child)
		if tok, ok := ch.(antlr.TerminalNode); ok {
			if tok.GetText() == "-" {
				ret.neg = !ret.neg
			}
		} else if expr, ok := ch.(*parser.OC_StringListNullOperatorExpressionContext); ok {
			ret.expr = oC_StringListNullOperatorExpression(expr)
		}
	}
	if !ret.neg {
		return ret.expr
	}
	return &ret
}

func oC_StringListNullOperatorExpression(ctx *parser.OC_StringListNullOperatorExpressionContext) stringListNullOperatorExpression {
	var ret stringListNullOperatorExpression
	for child := 0; child < ctx.GetChildCount(); child++ {
		switch t := ctx.GetChild(child).(type) {
		case *parser.OC_PropertyOrLabelsExpressionContext:
			ret.propertyOrLabels = oC_PropertyOrLabelsExpression(t)
		case *parser.OC_StringOperatorExpressionContext:
			ret.parts = append(ret.parts, oC_StringOperatorExpression(t))
		case *parser.OC_ListOperatorExpressionContext:
			ret.parts = append(ret.parts, oC_ListOperatorExpression(t))
		case *parser.OC_NullOperatorExpressionContext:
			ret.parts = append(ret.parts, oC_NullOperatorExpression(t))
		}
	}
	return ret
}

func oC_ListOperatorExpression(ctx *parser.OC_ListOperatorExpressionContext) stringListNullOperatorExpressionPart {
	if x := ctx.OC_PropertyOrLabelsExpression(); x != nil {
		return stringListNullOperatorExpressionPart{
			listIn: oC_PropertyOrLabelsExpression(x.(*parser.OC_PropertyOrLabelsExpressionContext)),
		}
	}
	expr := ctx.AllOC_Expression()
	if len(expr) == 1 {
		return stringListNullOperatorExpressionPart{
			listIndex: oC_Expression(expr[0].(*parser.OC_ExpressionContext)),
		}
	}
	return stringListNullOperatorExpressionPart{
		listRange: &listRangeExpression{
			first:  oC_Expression(expr[0].(*parser.OC_ExpressionContext)),
			second: oC_Expression(expr[1].(*parser.OC_ExpressionContext)),
		}}
}

func oC_StringOperatorExpression(ctx *parser.OC_StringOperatorExpressionContext) stringListNullOperatorExpressionPart {
	ret := stringListNullOperatorExpressionPart{
		stringOp: &stringOperatorExpression{
			expr: oC_PropertyOrLabelsExpression(ctx.OC_PropertyOrLabelsExpression().(*parser.OC_PropertyOrLabelsExpressionContext)),
		},
	}
	if ctx.STARTS() != nil {
		ret.stringOp.operator = "STARTS"
	}
	if ctx.ENDS() != nil {
		ret.stringOp.operator = "ENDS"
	}
	if ctx.CONTAINS() != nil {
		ret.stringOp.operator = "CONTAINS"
	}
	return ret
}

func oC_NullOperatorExpression(ctx *parser.OC_NullOperatorExpressionContext) stringListNullOperatorExpressionPart {
	val := ctx.NOT() == nil
	return stringListNullOperatorExpressionPart{
		isNull: &val,
	}
}

func oC_PropertyOrLabelsExpression(ctx *parser.OC_PropertyOrLabelsExpressionContext) propertyOrLabelsExpression {
	ret := propertyOrLabelsExpression{
		atom: oC_Atom(ctx.OC_Atom().(*parser.OC_AtomContext)),
	}
	for _, x := range ctx.AllOC_PropertyLookup() {
		ret.propertyLookup = append(ret.propertyLookup, oC_PropertyLookup(x.(*parser.OC_PropertyLookupContext)))
	}
	if l := ctx.OC_NodeLabels(); l != nil {
		x := oC_NodeLabels(l.(*parser.OC_NodeLabelsContext))
		ret.nodeLabels = &x
	}
	return ret
}

func oC_PartialComparisonExpression(ctx *parser.OC_PartialComparisonExpressionContext) partialComparisonExpression {
	ret := partialComparisonExpression{expr: oC_AddOrSubtractExpression(ctx.OC_AddOrSubtractExpression().(*parser.OC_AddOrSubtractExpressionContext))}
	for i := 0; i < ctx.GetChildCount(); i++ {
		if tok, ok := ctx.GetChild(i).(antlr.TerminalNode); ok {
			t := tok.GetText()
			if t == "=" || t == "<>" || t == "<" || t == ">" || t == "<=" || t == "=>" {
				ret.op = t
			}
		}
	}
	return ret
}

func oC_PropertyLookup(ctx *parser.OC_PropertyLookupContext) schemaName {
	return oC_PropertyKeyName(ctx.OC_PropertyKeyName().(*parser.OC_PropertyKeyNameContext))
}

func oC_PropertyKeyName(ctx *parser.OC_PropertyKeyNameContext) schemaName {
	return oC_SchemaName(ctx.OC_SchemaName().(*parser.OC_SchemaNameContext))
}

func oC_SchemaName(ctx *parser.OC_SchemaNameContext) schemaName {
	if x := ctx.OC_SymbolicName(); x != nil {
		name := oC_SymbolicName(ctx.OC_SymbolicName().(*parser.OC_SymbolicNameContext))
		return schemaName{symbolicName: &name}
	}
	name := oC_ReservedWord(ctx.OC_ReservedWord().(*parser.OC_ReservedWordContext))
	return schemaName{reservedWord: &name}
}

func oC_SymbolicName(ctx *parser.OC_SymbolicNameContext) symbolicName {
	if node := ctx.EscapedSymbolicName(); node != nil {
		text := node.GetText()
		return symbolicName(text[1 : len(text)-1])
	}
	return symbolicName(ctx.GetText())
}

func oC_ReservedWord(ctx *parser.OC_ReservedWordContext) reservedWord {
	return reservedWord(strings.ToUpper(ctx.GetText()))
}

func oC_NodeLabels(ctx *parser.OC_NodeLabelsContext) NodeLabels {
	children := ctx.AllOC_NodeLabel()
	ret := make([]schemaName, 0, len(children))
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
		ret.variable = &vr
	}
	ret.start, ret.path = oC_AnonymousPatternPart(ctx.OC_AnonymousPatternPart().(*parser.OC_AnonymousPatternPartContext))
	return ret
}

func oC_AnonymousPatternPart(ctx *parser.OC_AnonymousPatternPartContext) (nodePattern, []patternChain) {
	return oC_PatternElement(ctx.OC_PatternElement().(*parser.OC_PatternElementContext))
}

func oC_PatternElement(ctx *parser.OC_PatternElementContext) (nodePattern, []patternChain) {
	if p := ctx.OC_PatternElement(); p != nil {
		return oC_PatternElement(p.(*parser.OC_PatternElementContext))
	}
	np := oC_NodePattern(ctx.OC_NodePattern().(*parser.OC_NodePatternContext))
	chain := make([]patternChain, 0)
	for _, x := range ctx.AllOC_PatternElementChain() {
		chain = append(chain, oC_PatternElementChain(x.(*parser.OC_PatternElementChainContext)))
	}
	return np, chain
}

func oC_NodePattern(ctx *parser.OC_NodePatternContext) nodePattern {
	ret := nodePattern{}
	if x := ctx.OC_Variable(); x != nil {
		v := oC_Variable(x.(*parser.OC_VariableContext))
		ret.variable = &v
	}
	if x := ctx.OC_NodeLabels(); x != nil {
		l := oC_NodeLabels(x.(*parser.OC_NodeLabelsContext))
		ret.labels = &l
	}
	if x := ctx.OC_Properties(); x != nil {
		o := oC_Properties(x.(*parser.OC_PropertiesContext))
		ret.properties = &o
	}
	return ret
}

func oC_PatternElementChain(ctx *parser.OC_PatternElementChainContext) patternChain {
	return patternChain{
		rel:  oC_RelationshipPattern(ctx.OC_RelationshipPattern().(*parser.OC_RelationshipPatternContext)),
		node: oC_NodePattern(ctx.OC_NodePattern().(*parser.OC_NodePatternContext)),
	}
}

func oC_RelationshipPattern(ctx *parser.OC_RelationshipPatternContext) relationshipPattern {
	ret := relationshipPattern{
		toLeft:  ctx.OC_LeftArrowHead() != nil,
		toRight: ctx.OC_RightArrowHead() != nil,
	}
	if x := ctx.OC_RelationshipDetail(); x != nil {
		detctx := x.(*parser.OC_RelationshipDetailContext)
		if x := detctx.OC_Variable(); x != nil {
			v := oC_Variable(x.(*parser.OC_VariableContext))
			ret.variable = &v
		}
		if x := detctx.OC_RelationshipTypes(); x != nil {
			r := oC_RelationshipTypes(x.(*parser.OC_RelationshipTypesContext))
			ret.relTypes = &r
		}
		if x := detctx.OC_RangeLiteral(); x != nil {
			r := oC_RangeLiteral(x.(*parser.OC_RangeLiteralContext))
			ret.rng = &r
		}
		if x := detctx.OC_Properties(); x != nil {
			p := oC_Properties(x.(*parser.OC_PropertiesContext))
			ret.properties = &p
		}
	}
	return ret
}

func oC_Variable(ctx *parser.OC_VariableContext) variable {
	return variable(oC_SymbolicName(ctx.OC_SymbolicName().(*parser.OC_SymbolicNameContext)))
}

func oC_RelationshipTypes(ctx *parser.OC_RelationshipTypesContext) relationshipTypes {
	ret := relationshipTypes{}
	for _, x := range ctx.AllOC_RelTypeName() {
		ret.rel = append(ret.rel, oC_RelTypeName(x.(*parser.OC_RelTypeNameContext)))
	}
	return ret
}

func oC_RelTypeName(ctx *parser.OC_RelTypeNameContext) schemaName {
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

func decimalInteger(ctx antlr.TerminalNode) intLiteral {
	i, err := strconv.Atoi(ctx.GetText())
	if err != nil {
		panic(err)
	}
	return intLiteral(i)
}

func oC_RangeLiteral(ctx *parser.OC_RangeLiteralContext) rangeLiteral {
	ret := rangeLiteral{}
	values := make([]intLiteral, 0, 2)
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
		ret.from = &values[0]
		ret.to = &values[1]
	case 1:
		switch rangeIndex {
		case 0: // ..value
			ret.to = &values[0]
		case 1: // value..
			ret.from = &values[0]
		case -1: // value
			ret.from = &values[0]
			ret.to = &values[0]
		}
	}
	return ret
}

func oC_IntegerLiteral(ctx *parser.OC_IntegerLiteralContext) intLiteral {
	if x := ctx.HexInteger(); x != nil {
		t := x.GetText()
		val, err := strconv.ParseInt(string(t[2:]), 16, 64)
		if err != nil {
			panic(err)
		}
		return intLiteral(val)
	}
	if x := ctx.OctalInteger(); x != nil {
		t := x.GetText()
		val, err := strconv.ParseInt(t, 8, 64)
		if err != nil {
			panic(err)
		}
		return intLiteral(val)
	}
	return decimalInteger(ctx.DecimalInteger())
}

func oC_BooleanLiteral(ctx *parser.OC_BooleanLiteralContext) booleanLiteral {
	if ctx.TRUE() != nil {
		return booleanLiteral(true)
	}
	return booleanLiteral(false)
}

func oC_ListLiteral(ctx *parser.OC_ListLiteralContext) *listLiteral {
	ret := &listLiteral{}
	for _, x := range ctx.AllOC_Expression() {
		ret.values = append(ret.values, oC_Expression(x.(*parser.OC_ExpressionContext)))
	}
	return ret
}

func oC_MapLiteral(ctx *parser.OC_MapLiteralContext) *mapLiteral {
	ret := &mapLiteral{}
	var keyName schemaName
	for i := 0; i < ctx.GetChildCount(); i++ {
		switch c := ctx.GetChild(i).(type) {
		case *parser.OC_PropertyKeyNameContext:
			keyName = oC_SchemaName(c.OC_SchemaName().(*parser.OC_SchemaNameContext))
		case *parser.OC_ExpressionContext:
			value := mapKeyValue{
				key:   keyName.String(),
				value: oC_Expression(c),
			}
			ret.keyValues = append(ret.keyValues, value)
		}
	}
	return ret
}

func oC_Atom(ctx *parser.OC_AtomContext) atom {
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
		return countAtom{}
	}

	if l := ctx.OC_ListComprehension(); l != nil {
		return oC_ListComprehension(l.(*parser.OC_ListComprehensionContext))
	}
	if p := ctx.OC_PatternComprehension(); p != nil {
		return oC_PatternComprehension(p.(*parser.OC_PatternComprehensionContext))
	}
	if ctx.ALL() != nil {
		return filterAtom{
			op:     "ALL",
			filter: oC_FilterExpression(ctx.OC_FilterExpression().(*parser.OC_FilterExpressionContext)),
		}
	}
	if ctx.ANY() != nil {
		return filterAtom{
			op:     "ANY",
			filter: oC_FilterExpression(ctx.OC_FilterExpression().(*parser.OC_FilterExpressionContext)),
		}
	}
	if ctx.NONE() != nil {
		return filterAtom{
			op:     "NONE",
			filter: oC_FilterExpression(ctx.OC_FilterExpression().(*parser.OC_FilterExpressionContext)),
		}
	}
	if ctx.SINGLE() != nil {
		return filterAtom{
			op:     "SINGLE",
			filter: oC_FilterExpression(ctx.OC_FilterExpression().(*parser.OC_FilterExpressionContext)),
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

func oC_FilterExpression(ctx *parser.OC_FilterExpressionContext) filterExpression {
	idInColl := ctx.OC_IdInColl().(*parser.OC_IdInCollContext)
	ret := filterExpression{
		variable: oC_Variable(idInColl.OC_Variable().(*parser.OC_VariableContext)),
		inExpr:   oC_Expression(idInColl.OC_Expression().(*parser.OC_ExpressionContext)),
	}
	if w := ctx.OC_Where(); w != nil {
		ret.where = oC_Where(w.(*parser.OC_WhereContext))
	}

	return ret
}

// oC_RelationshipsPattern   :  oC_NodePattern ( SP? oC_PatternElementChain )+ ;
// oC_PatternElementChain :  oC_RelationshipPattern SP? oC_NodePattern ;
func oC_RelationshipsPattern(ctx *parser.OC_RelationshipsPatternContext) relationshipsPattern {
	ret := relationshipsPattern{
		start: oC_NodePattern(ctx.OC_NodePattern().(*parser.OC_NodePatternContext)),
	}
	for _, element := range ctx.AllOC_PatternElementChain() {
		ret.chain = append(ret.chain, oC_PatternElementChain(element.(*parser.OC_PatternElementChainContext)))
	}
	return ret
}

func oC_ListComprehension(ctx *parser.OC_ListComprehensionContext) listComprehension {
	ret := listComprehension{
		filter: oC_FilterExpression(ctx.OC_FilterExpression().(*parser.OC_FilterExpressionContext)),
	}
	if x := ctx.OC_Expression(); x != nil {
		ret.expr = oC_Expression(x.(*parser.OC_ExpressionContext))
	}
	return ret
}

func oC_FunctionInvocation(ctx *parser.OC_FunctionInvocationContext) *functionInvocation {
	ret := &functionInvocation{
		name:     oC_FunctionName(ctx.OC_FunctionName().(*parser.OC_FunctionNameContext)),
		distinct: ctx.DISTINCT() != nil,
	}
	for _, x := range ctx.AllOC_Expression() {
		ret.args = append(ret.args, oC_Expression(x.(*parser.OC_ExpressionContext)))
	}
	return ret
}

func oC_FunctionName(ctx *parser.OC_FunctionNameContext) []symbolicName {
	if ctx.EXISTS() != nil {
		return []symbolicName{"EXISTS"}
	}
	ns := oC_Namespace(ctx.OC_Namespace().(*parser.OC_NamespaceContext))
	return append(ns, oC_SymbolicName(ctx.OC_SymbolicName().(*parser.OC_SymbolicNameContext)))
}

func oC_Namespace(ctx *parser.OC_NamespaceContext) []symbolicName {
	ret := make([]symbolicName, 0)
	for _, x := range ctx.AllOC_SymbolicName() {
		ret = append(ret, oC_SymbolicName(x.(*parser.OC_SymbolicNameContext)))
	}
	return ret
}

func oC_PatternComprehension(ctx *parser.OC_PatternComprehensionContext) patternComprehension {
	ret := patternComprehension{
		rel: oC_RelationshipsPattern(ctx.OC_RelationshipsPattern().(*parser.OC_RelationshipsPatternContext)),
	}
	if v := ctx.OC_Variable(); v != nil {
		x := oC_Variable(v.(*parser.OC_VariableContext))
		ret.variable = &x
	}
	expr := ctx.AllOC_Expression()
	if len(expr) == 1 {
		ret.expr = oC_Expression(expr[0].(*parser.OC_ExpressionContext))
	} else {
		ret.where = oC_Expression(expr[0].(*parser.OC_ExpressionContext))
		ret.expr = oC_Expression(expr[1].(*parser.OC_ExpressionContext))
	}
	return ret
}

func oC_CaseExpression(ctx *parser.OC_CaseExpressionContext) caseClause {
	ret := caseClause{}
	for _, alt := range ctx.AllOC_CaseAlternatives() {
		ret.alternatives = append(ret.alternatives, oC_CaseAlternatives(alt.(*parser.OC_CaseAlternativesContext)))
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
				ret.def = oC_Expression(expr)
			} else {
				ret.test = oC_Expression(expr)
			}
		}
	}
	return ret
}

func oC_CaseAlternatives(ctx *parser.OC_CaseAlternativesContext) caseAlternative {
	expr := ctx.AllOC_Expression()
	return caseAlternative{
		when: oC_Expression(expr[0].(*parser.OC_ExpressionContext)),
		then: oC_Expression(expr[1].(*parser.OC_ExpressionContext)),
	}
}

func oC_Literal(ctx *parser.OC_LiteralContext) Evaluatable {
	if ctx.NULL() != nil {
		return nullLiteral{}
	}
	if n := ctx.OC_NumberLiteral(); n != nil {
		return oC_NumberLiteral(n.(*parser.OC_NumberLiteralContext))
	}
	if s := ctx.StringLiteral(); s != nil {
		text := s.GetText()
		return stringLiteral(text[1 : len(text)-1])
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

func oC_DoubleLiteral(ctx *parser.OC_DoubleLiteralContext) doubleLiteral {
	v, err := strconv.ParseFloat(ctx.GetText(), 64)
	if err != nil {
		panic(err)
	}
	return doubleLiteral(v)
}

func oC_Unwind(ctx *parser.OC_UnwindContext) unwind {
	return unwind{
		expr: oC_Expression(ctx.OC_Expression().(*parser.OC_ExpressionContext)),
		as:   oC_Variable(ctx.OC_Variable().(*parser.OC_VariableContext)),
	}
}

func oC_Set(ctx *parser.OC_SetContext) UpdatingClause {
	ret := &set{}
	for _, item := range ctx.AllOC_SetItem() {
		ret.items = append(ret.items, oC_SetItem(item.(*parser.OC_SetItemContext)))
	}
	return ret
}

func oC_SetItem(ctx *parser.OC_SetItemContext) setItem {
	if p := ctx.OC_PropertyExpression(); p != nil {
		pe := oC_PropertyExpression(p.(*parser.OC_PropertyExpressionContext))
		return setItem{
			property:   &pe,
			expression: oC_Expression(ctx.OC_Expression().(*parser.OC_ExpressionContext)),
		}
	}
	v := oC_Variable(ctx.OC_Variable().(*parser.OC_VariableContext))
	if expr := ctx.OC_Expression(); expr != nil {
		ret := setItem{
			variable:   &v,
			expression: oC_Expression(expr.(*parser.OC_ExpressionContext)),
		}
		for i := 0; i < ctx.GetChildCount(); i++ {
			ch := ctx.GetChild(i)
			if term, ok := ch.(antlr.TerminalNode); ok {
				text := term.GetText()
				if text == "=" || text == "+=" {
					ret.op = text
					break
				}
			}
		}
		return ret
	}
	ret := setItem{
		variable:   &v,
		nodeLabels: oC_NodeLabels(ctx.OC_NodeLabels().(*parser.OC_NodeLabelsContext)),
	}
	return ret
}

func oC_PropertyExpression(ctx *parser.OC_PropertyExpressionContext) propertyExpression {
	ret := propertyExpression{
		atom: oC_Atom(ctx.OC_Atom().(*parser.OC_AtomContext)),
	}
	for _, l := range ctx.AllOC_PropertyLookup() {
		ret.lookup = append(ret.lookup, oC_PropertyLookup(l.(*parser.OC_PropertyLookupContext)))
	}
	return ret
}

func oC_Delete(ctx *parser.OC_DeleteContext) UpdatingClause {
	ret := deleteClause{
		detach: ctx.DETACH() != nil,
	}
	for _, e := range ctx.AllOC_Expression() {
		ret.exprs = append(ret.exprs, oC_Expression(e.(*parser.OC_ExpressionContext)))
	}
	return ret
}

func oC_Remove(ctx *parser.OC_RemoveContext) UpdatingClause {
	ret := remove{}
	for _, item := range ctx.AllOC_RemoveItem() {
		ret.items = append(ret.items, oC_RemoveItem(item.(*parser.OC_RemoveItemContext)))
	}
	return ret
}

func oC_RemoveItem(ctx *parser.OC_RemoveItemContext) removeItem {
	ret := removeItem{}
	if v := ctx.OC_Variable(); v != nil {
		variable := oC_Variable(v.(*parser.OC_VariableContext))
		ret.variable = &variable
	}
	if v := ctx.OC_NodeLabels(); v != nil {
		ret.nodeLabels = oC_NodeLabels(v.(*parser.OC_NodeLabelsContext))
	}
	if v := ctx.OC_PropertyExpression(); v != nil {
		expr := oC_PropertyExpression(v.(*parser.OC_PropertyExpressionContext))
		ret.property = &expr
	}
	return ret
}

func oC_Create(ctx *parser.OC_CreateContext) UpdatingClause {
	ret := create{
		pattern: oC_Pattern(ctx.OC_Pattern().(*parser.OC_PatternContext)),
	}
	return ret
}

func oC_Merge(ctx *parser.OC_MergeContext) UpdatingClause {
	ret := merge{
		pattern: oC_PatternPart(ctx.OC_PatternPart().(*parser.OC_PatternPartContext)),
	}
	for _, action := range ctx.AllOC_MergeAction() {
		actx := action.(*parser.OC_MergeActionContext)
		action := mergeAction{
			set: oC_Set(actx.OC_Set().(*parser.OC_SetContext)),
		}
		if actx.MATCH() == nil {
			action.on = mergeActionOnCreate
		} else {
			action.on = mergeActionOnMatch
		}
		ret.actions = append(ret.actions, action)
	}
	return ret
}

func oC_InQueryCall(ctx *parser.OC_InQueryCallContext) ReadingClause {
	panic("Unsupported: inQueryCall")
}

func oC_StandaloneCall(ctx *parser.OC_StandaloneCallContext) Evaluatable {
	panic("Unsupported: standaloneCall")
}
