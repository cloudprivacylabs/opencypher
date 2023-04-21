package opencypher

// CartesianProuductPaths builds the product of all the resultpaths
func CartesianProductPaths(paths [][]ResultPath, filter func() []ResultPath) [][]ResultPath {
	// nextIndex retrieves the next index for the next set
	nextIndex := func(ix []int, lengthsFunc func(i int) int) {
		for j := len(ix) - 1; j >= 0; j-- {
			ix[j]++
			if j == 0 || ix[j] < lengthsFunc(j) {
				return
			}
			ix[j] = 0
		}
	}
	lengths := func(i int) int { return len(paths[i]) }
	product := make([][]ResultPath, 0)
	for ix := make([]int, len(paths)); ix[0] < lengths(0); nextIndex(ix, lengths) {
		var r []ResultPath
		for j, k := range ix {
			r = append(r, paths[j][k])
		}
		product = append(product, r)
	}
	return product
}
