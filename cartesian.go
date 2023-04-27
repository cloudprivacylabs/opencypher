package opencypher

// CartesianProuductPaths builds the product of all the resultpaths
func CartesianProductPaths(ctx *EvalContext, numItems int, getItem func(int, *EvalContext) []ResultPath, filter func([]ResultPath) bool) [][]ResultPath {
	product := make([][]ResultPath, numItems)
	indexes := make([]int, numItems)
	columnProcessor := func(next func(int)) func(int) {
		return func(column int) {
			product[column] = getItem(column, ctx)
			for i := range product[column] {
				indexes[column] = i
				next(column + 1)
			}
		}
	}

	result := make([][]ResultPath, 0)

	capture := func(int) {
		row := make([]ResultPath, 0, numItems)
		for i, x := range indexes {
			row = append(row, product[i][x])
		}
		result = append(result, row)
	}

	next := columnProcessor(capture)
	for column := numItems - 2; column >= 0; column-- {
		next = columnProcessor(next)
	}
	next(0)
	return result
}

// nextIndex retrieves the next index for the next set
// 	nextIndex := func(ix []int) {
// 		for j := len(ix) - 1; j >= 0; j-- {
// 			ix[j]++
// 			if j == 0 || ix[j] < len(getItem(j, ctx)) {
// 				return
// 			}
// 			ix[j] = 0
// 		}
// 	}
// 	result := make([][]ResultPath, 0)
// 	indexes := make([]int, numItems)
// 	level := 1

// 	// Match Q1, Q2, Q3
// 	//
// 	// There are 3 levels of queries
// 	//
// 	//
// 	//   Q1    Q2      Q3
// 	//   -----------------
// 	//   p11  p11_21  p11_21_31
// 	//        p11_22  p11_21_32
// 	//
// 	//   p12
// 	//   ...

// 	// getItems(0)
// 	// for each items[0] {
// 	// 	getItems(1)
// 	// 	for each items[1] {
// 	// 		getItems(2)
// 	// 		for each items [2]
// 	// 		capture

// 	for {
// 		product[level] = getItem(level, ctx)
// 		for indexes[level], path = range product[level] {
// 			if level+1 == numItems {
// 				if filter(something) {
// 					captureResult()
// 				}
// 			} else {
// 				level++
// 			}
// 		}
// 		if level+1 == numItems {
// 			level--
// 			if indexes[level]+1>=len(product[level]) {
// 		}
// 	}
// 	return product
// }

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
