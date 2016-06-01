package main

import "github.com/unixpickle/num-analysis/linalg"

func MaxValueIndex(v linalg.Vector) int {
	var maxVal float64
	var maxIdx int
	for i, x := range v {
		if x > maxVal {
			maxVal = x
			maxIdx = i
		}
	}
	return maxIdx
}
