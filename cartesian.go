package opencypher

// CartesianProuductPaths builds the product of all the resultpaths
func CartesianProductPaths(ctx *EvalContext, numItems int, all bool, getItem func(int, *EvalContext) []ResultPath, filter func([]ResultPath) bool) [][]ResultPath {
	product := make([][]ResultPath, numItems)
	product[0] = getItem(0, ctx)
	if numItems == 1 {
		return product
	}

	indexes := make([]int, numItems)
	for {
		var path []ResultPath
		for i := range indexes {
			path = product[i]
			if path == nil {
				path = getItem(i, ctx)
			}
			if indexes[i] >= len(path) {
				if !all {
					return product
				}
				if filter(path) {
					product = append(product, path)
				}
				product[i] = nil
			}
		}
		carry := false
		for i := range indexes {
			indexes[i]++
			if indexes[i] >= len(path) {
				indexes[i] = 0
				carry = true
				continue
			}
			carry = false
			break
		}
		if carry {
			break
		}
	}
	return product
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
