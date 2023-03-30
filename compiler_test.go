package opencypher

import (
	"testing"

	"github.com/cloudprivacylabs/lpg/v2"
	"github.com/cloudprivacylabs/opencypher/parser"
)

func TestExpr(t *testing.T) {
	c := GetParser(`5  +  7+1`).OC_Expression()
	out := oC_Expression(c.(*parser.OC_ExpressionContext))
	result, err := out.Evaluate(NewEvalContext(lpg.NewGraph()))
	if err != nil {
		t.Error(err)
	}
	if result.Get() != 13 {
		t.Errorf("Wrong result: %+v %T", result, result.Get())
	}
}

func runTestMatch(t *testing.T, expr string, g *lpg.Graph) ResultSet {
	ctx := NewEvalContext(g)
	ev, err := Parse(expr)
	if err != nil {
		t.Errorf("%s: %s", expr, err)
		return *NewResultSet()
	}
	value, err := ev.Evaluate(ctx)
	if err != nil {
		t.Errorf("%s: %s", expr, err)
		return *NewResultSet()
	}

	return value.Get().(ResultSet)
}

func TestBasicMatch(t *testing.T) {

	g := lpg.NewGraph()
	// (:t1:t2)
	n1 := g.NewNode([]string{"t1", "t2"}, nil)
	rs := runTestMatch(t, "match (n) return n", g)
	if !rs.Rows[0]["1"].Get().(*lpg.Node).GetLabels().HasAll("t1", "t2") {
		t.Errorf("Expecting to see one row with t1, t2")
	}

	// (:t1:t2) (:t1:t3)
	n2 := g.NewNode([]string{"t1", "t3"}, nil)
	rs = runTestMatch(t, "match (n:t1) return n", g)
	if !(len(rs.Rows) == 2 && rs.Rows[0]["1"].Get().(*lpg.Node).GetLabels().Has("t1") &&
		rs.Rows[1]["1"].Get().(*lpg.Node).GetLabels().Has("t1")) {
		t.Errorf("Expecting to see two rows with t1, t2")
	}
	rs = runTestMatch(t, "match (n:t2) return n", g)
	if !(len(rs.Rows) == 1 && rs.Rows[0]["1"].Get().(*lpg.Node).GetLabels().Has("t1")) {
		t.Errorf("Expecting to see one row with t1: %v", rs)
	}

	// (:t1:t2) --[:e1]-->(:t1:t3)
	g.NewEdge(n1, n2, "e1", nil)
	rs = runTestMatch(t, "match (m:t1)-[:e1]->(n:t3) return n", g)
	if !(len(rs.Rows) == 1 && rs.Rows[0]["1"].Get().(*lpg.Node) == n2) {
		t.Errorf("Expecting to see one row n2: %v", rs)
	}
	rs = runTestMatch(t, "match (m:t1)<-[:e1]-(n:t1) return n", g)
	if !(len(rs.Rows) == 1 && rs.Rows[0]["1"].Get().(*lpg.Node) == n1) {
		t.Errorf("Expecting to see one row n1: %v", rs)
	}
	rs = runTestMatch(t, "match (m:t1)<-[:e1]-(n:t1) return n,m", g)
	if !(len(rs.Rows) == 1 && rs.Rows[0]["1"].Get().(*lpg.Node) == n1 && rs.Rows[0]["2"].Get().(*lpg.Node) == n2) {
		t.Errorf("Expecting to see one row n1 n2: %v", rs)
	}

	// (:t1:t2) --[:e1]-->(:t1:t3) --[:e2]-->(:t4)
	n3 := g.NewNode([]string{"t4"}, nil)
	g.NewEdge(n2, n3, "e2", nil)
	rs = runTestMatch(t, "match (m:t1)-[*]->(n:t4) return n", g)
	if !(len(rs.Rows) == 2 && rs.Rows[0]["1"].Get().(*lpg.Node) == n3 && rs.Rows[1]["1"].Get().(*lpg.Node) == n3) {
		t.Errorf("Expecting to see two rows n3: %v", rs)
	}

	{
		ctx := NewEvalContext(g)
		ctx.SetVar("m", RValue{})
		ev, err := Parse("match (m)-[*]->(n:t4) return n")
		if err != nil {
			t.Error(err)
		}
		value, err := ev.Evaluate(ctx)
		if err != nil {
			t.Error(err)
		}
		if len(value.Get().(ResultSet).Rows) != 0 {
			t.Errorf("Empty result set expected")
		}
	}

}

func TestCaseExpr(t *testing.T) {
	ctx := NewEvalContext(lpg.NewGraph())
	ctx.SetVar("x", RValue{Value: "a"})
	ctx.SetVar("y", RValue{Value: "test"})
	ctx.SetVar("spl", RValue{Value: []Value{RValue{Value: "text"}}})
	v, err := ParseAndEvaluate(`return case x
   when null then y
   else trim(substring(y,0,size(y)-size(spl[size(spl)-1])))
   end as z`, ctx)
	if err != nil {
		t.Error(err)
	}
	t.Log(v)
}

func TestQueryChain(t *testing.T) {
	g := lpg.NewGraph()
	ctx := NewEvalContext(g)
	// (:l1:1) -[]->(:l2:1)
	// (:l1:1) -[]->(:l3:1)
	// (:l1:2) -[]->(:l2:2)
	// (:l1:2) -[]->(:l3:2)
	l1_1 := g.NewNode([]string{"l1", "1"}, nil)
	l2_1 := g.NewNode([]string{"l2", "1"}, nil)
	l3_1 := g.NewNode([]string{"l3", "1"}, nil)
	l1_2 := g.NewNode([]string{"l1", "2"}, nil)
	l2_2 := g.NewNode([]string{"l2", "2"}, nil)
	l3_2 := g.NewNode([]string{"l3", "2"}, nil)
	g.NewEdge(l1_1, l2_1, "edge", nil)
	g.NewEdge(l1_1, l3_1, "edge", nil)
	g.NewEdge(l1_2, l2_2, "edge", nil)
	g.NewEdge(l1_2, l3_2, "edge", nil)

	v, err := ParseAndEvaluate(`match (root:l1)-[]->(a:l2), (root)-[]->(b:l3) return a as a, b as b`, ctx)
	if err != nil {
		t.Error(err)
	}
	rs := v.Get().(ResultSet)
	if len(rs.Rows) != 2 {
		t.Errorf("Expecting 2 rows got %d", len(rs.Rows))
	}
	if rs.Rows[0]["a"].Get() != l2_1 {
		t.Errorf("Wrong a")
	}
	if rs.Rows[0]["b"].Get() != l3_1 {
		t.Errorf("Wrong b")
	}
}
