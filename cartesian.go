package opencypher

// CartesianProuductPaths builds the product of all the resultpaths
func CartesianProductPaths(ctx *EvalContext, numItems int, getItem func(int, *EvalContext) []ResultPath, filter func([]ResultPath) bool) [][]ResultPath {
	// nextIndex retrieves the next index for the next set
	nextIndex := func(ix []int) {
		for j := len(ix) - 1; j >= 0; j-- {
			ix[j]++
			if j == 0 || ix[j] < len(getItem(j, ctx)) {
				return
			}
			ix[j] = 0
		}
	}
	if numItems == 1 {

	}
	product := make([][]ResultPath, 0, numItems)
	basePath := getItem(0, ctx)
	for ix := make([]int, numItems); ix[0] < len(basePath); nextIndex(ix) {
		r := make([]ResultPath, 0, len(ix))
		for j := range ix {
			r = append(r, getItem(j, ctx)...)
		}
		if filter(r) {
			product = append(product, r)
		}
	}
	return product
}

func AllPathsToResultSets(paths [][]ResultPath) []ResultSet {
	res := make([]ResultSet, len(paths))
	for i, path := range paths {
		for j := range path {
			res[i].Rows[i] = paths[i][j].Symbols
		}
	}
	return res
}

func ResultPathToResultSet(resultPath []ResultPath) ResultSet {
	rs := ResultSet{Rows: make([]map[string]Value, 0, len(resultPath))}
	for _, rp := range resultPath {
		rs.Rows = append(rs.Rows, rp.Symbols)
	}
	return rs
}
