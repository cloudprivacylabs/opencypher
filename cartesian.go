package opencypher

// CartesianProuductPaths builds the product of all the resultpaths
func CartesianProductPaths(ctx *EvalContext, numItems int, getItem func(int, *EvalContext) ([]ResultPath, error), filter func([]ResultPath) bool) [][]ResultPath {
	result := make([][]ResultPath, 0)
	product := make([][]ResultPath, numItems)
	indexes := make([]int, numItems)

	columnProcessor := func(next func(int)) func(int) {
		return func(column int) {
			product[column], _ = getItem(column, ctx)
			for i := range product[column] {
				indexes[column] = i
				next(column + 1)
			}
		}
	}

	capture := func(int) {
		row := make([]ResultPath, 0, numItems)
		for i, x := range indexes {
			row = append(row, product[i][x])
		}
		if filter(row) {
			result = append(result, row)
		}
	}

	next := columnProcessor(capture)
	for column := numItems - 2; column >= 0; column-- {
		next = columnProcessor(next)
	}
	next(0)
	return result
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
