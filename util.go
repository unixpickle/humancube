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

func MoveForOutput(n *Network, out linalg.Vector) string {
	moveIdx := MaxValueIndex(out)

	for name, idx := range n.MoveMap {
		if idx == moveIdx {
			return name
		}
	}

	panic("code should be unreachable")
}
