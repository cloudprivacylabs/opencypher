package opencypher

// CartesianProuductPaths builds the product of all the resultpaths
func CartesianProductPaths(paths [][]ResultPath, filter func(rp []ResultPath) bool) [][]ResultPath {
	// nextIndex retrieves the next index for the next set
	nextIndex := func(ix []int) {
		for j := len(ix) - 1; j >= 0; j-- {
			ix[j]++
			if j == 0 || ix[j] < len(paths[j]) {
				return
			}
			ix[j] = 0
		}
	}
	if len(paths) == 1 {
		return paths
	}
	product := make([][]ResultPath, 0, len(paths))
	for ix := make([]int, len(paths)); ix[0] < len(paths[0]); nextIndex(ix) {
		r := make([]ResultPath, 0, len(ix))
		for j, k := range ix {
			r = append(r, paths[j][k])
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
