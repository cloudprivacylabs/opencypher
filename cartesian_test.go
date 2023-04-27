package opencypher

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/cloudprivacylabs/lpg/v2"
)

func TestCartesianProductPaths(t *testing.T) {
	f, err := os.Open("testdata/g1.json")
	if err != nil {
		t.Error(err)
		return
	}
	target := lpg.NewGraph()
	err = lpg.JSON{}.Decode(target, json.NewDecoder(f))
	if err != nil {
		t.Error(err)
		return
	}
	path := &lpg.Path{}
	pe := make([]lpg.PathElement, 0)
	for itr := target.GetEdges(); itr.Next(); {
		// path.Append(lpg.PathElement{
		// 	Edge: itr.Edge(),
		// })
		pe = append(pe, lpg.PathElement{
			Edge: itr.Edge(),
		})
	}
	path.Append(pe...)
	{
		x := [][]ResultPath{
			{
				ResultPath{
					Result: path,
				},
				ResultPath{
					Result: &lpg.Path{},
				},
				ResultPath{
					Result: &lpg.Path{},
				},
			},
			{
				ResultPath{
					Result: &lpg.Path{},
				},
			},
			{
				ResultPath{
					Result: &lpg.Path{},
				},
			},
		}
		// j := [][]ResultPath{3, 2, 1}
		// k := [][]ResultPath{2, 1, 3}

		prod := CartesianProductPaths(NewEvalContext(target), len(x), func(i int, ec *EvalContext) []ResultPath {
			return x[i]
		}, func([]ResultPath) bool {
			return true
		})
		n := len(prod)
		e := len(x)
		fmt.Println(prod)
		if n != e {
			t.Fatalf("there should be %d but there were %d groups\n", e, n)
		}

	}
}

func TestCartesianProductPaths2(t *testing.T) {
	f, err := os.Open("testdata/g1.json")
	if err != nil {
		t.Error(err)
		return
	}
	target := lpg.NewGraph()
	err = lpg.JSON{}.Decode(target, json.NewDecoder(f))
	if err != nil {
		t.Error(err)
		return
	}
	path := &lpg.Path{}
	pe := make([]lpg.PathElement, 0)
	for itr := target.GetEdges(); itr.Next(); {
		// path.Append(lpg.PathElement{
		// 	Edge: itr.Edge(),
		// })
		pe = append(pe, lpg.PathElement{
			Edge: itr.Edge(),
		})
	}
	path.Append(pe...)
	{
		x := [][]ResultPath{
			{
				ResultPath{
					Result: path,
				},
				ResultPath{
					Result: &lpg.Path{},
				},
				ResultPath{
					Result: &lpg.Path{},
				},
			},
			{
				ResultPath{
					Result: &lpg.Path{},
				},
				ResultPath{
					Result: &lpg.Path{},
				},
			},
			{
				ResultPath{
					Result: &lpg.Path{},
				},
				ResultPath{
					Result: &lpg.Path{},
				},
			},
		}
		// j := [][]ResultPath{3, 2, 1}
		// k := [][]ResultPath{2, 1, 3}

		prod := CartesianProductPaths(NewEvalContext(target), len(x), func(i int, ec *EvalContext) []ResultPath {
			return x[i]
		}, func([]ResultPath) bool {
			return true
		})
		if len(prod) != 12 {
			t.Errorf("Got %d", len(prod))
			for _, x := range prod {
				fmt.Println("test2:", x)
			}
		}
	}
}
