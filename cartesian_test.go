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
		tests := []struct {
			rp     [][]ResultPath
			expLen int
		}{
			{
				rp: [][]ResultPath{
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
				},
				expLen: 12,
			},
			{
				rp: [][]ResultPath{
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
				},
				expLen: 3,
			},
			{
				rp: [][]ResultPath{
					{
						ResultPath{
							Result: path,
						},
					},
					{
						ResultPath{
							Result: &lpg.Path{},
						},
					},
				},
				expLen: 1,
			},
		}

		for _, test := range tests {
			prod := CartesianProductPaths(NewEvalContext(target), len(test.rp), func(i int, ec *EvalContext) ([]ResultPath, error) {
				return test.rp[i], nil
			}, func([]ResultPath) bool {
				return true
			})
			if len(prod) != test.expLen {
				t.Errorf("Got %d", len(prod))
				for _, x := range prod {
					fmt.Println("test2:", x)
				}
			}
		}
	}
}
