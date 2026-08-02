package trigger

import (
	"math"
)

func TriggerOnset(charfct []float64, thres1, thres2 float64, maxLen int, maxLenDelete bool) [][]int {
	ind1 := numpyWhereFloat(charfct, func(item float64) bool { return item >= thres1 })
	if len(ind1) == 0 {
		return [][]int{}
	}
	ind2 := numpyWhereFloat(charfct, func(item float64) bool { return item >= thres2 })

	on := deque{items: []int{ind1[0]}}
	of := deque{items: []int{-1}}

	// determine the indices where charfct falls below off-threshold
	ind2_ := numpyEmptyLikeBool(ind2)
	for i, diff := range numpyDiffInt(ind2) {
		ind2_[i] = diff > 1
	}

	// last occurence is missed by the diff, add it manually
	ind2_[len(ind2_)-1] = true
	var filteredInd2List []int
	for i, v := range ind2_ {
		if v {
			filteredInd2List = append(filteredInd2List, ind2[i])
		}
	}
	of.extend(filteredInd2List)
	var filteredInd1List []int
	for _, v := range numpyWhereInt(numpyDiffInt(ind1), func(f int) bool { return f > 1 }) {
		filteredInd1List = append(filteredInd1List, ind1[v+1])
	}
	on.extend(filteredInd1List)
	if maxLenDelete {
		// drop it
		of.extend([]int{math.MaxInt})
		on.extend([]int{on.items[len(on.items)-1]})
	} else {
		// include it
		of.extend([]int{ind2[len(ind2)-1]})
	}

	// add last sample to ensure trigger gets switched off if ctf does not fall
	// below off-threshold before hitting the end
	of.append(len(charfct))

	var pick [][]int
	for on.items[len(on.items)-1] > of.items[0] {
		for on.items[0] <= of.items[0] {
			on.popLeft()
		}
		for of.items[0] < on.items[0] {
			of.popLeft()
		}
		if of.items[0]-on.items[0] > maxLen {
			if maxLenDelete {
				on.popLeft()
				continue
			}
			of.appendLeft(on.items[0] + maxLen)
		}
		pick = append(pick, []int{on.items[0], of.items[0]})
	}

	return pick
}
