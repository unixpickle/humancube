package humancube

import (
	"strings"

	"github.com/unixpickle/gocube"
)

var (
	f2lEdges   = []int{1, 2, 3, 7, 8, 9, 10, 11}
	f2lCorners = []int{0, 1, 4, 5}

	lastLayerEdges   = []int{0, 4, 5, 6}
	lastLayerCorners = []int{2, 3, 6, 7}
)

type AugmentParams struct {
	// LLCases specifies the number of different last
	// layer cases to provide for a single F2L solve.
	LLCases int
}

// Augment generates new samples based on a SampleSet of
// complete solves.
func Augment(s *SampleSet, params *AugmentParams) {
	orig := s.Copy().(*SampleSet)
	s.Samples = append(s.Samples, augmentLastLayer(orig, params.LLCases)...)
}

func augmentLastLayer(orig *SampleSet, numCases int) []Sample {
	var res []Sample
	for _, sample := range orig.Samples {
		prefix, ok := sampleF2LPrefix(sample)
		if ok {
			for i := 0; i < numCases; i++ {
				res = append(res, randomizeLastLayer(sample, prefix))
			}
		}
	}
	return res
}

func randomizeLastLayer(s Sample, f2lSolve string) Sample {
	cube := *s.Start
	moves := strings.Fields(f2lSolve)
	for _, move := range moves {
		Move(&cube, move)
	}

	scramble := gocube.RandomZBLL()
	for _, idx := range lastLayerEdges {
		cube.Edges[idx] = scramble.Edges[idx]
	}
	for _, idx := range lastLayerCorners {
		cube.Corners[idx] = scramble.Corners[idx]
	}

	for i := len(moves) - 1; i >= 0; i-- {
		MoveInverse(&cube, moves[i])
	}
	return Sample{Start: &cube, Moves: f2lSolve}
}

func sampleF2LPrefix(s Sample) (string, bool) {
	var moves []string
	cube := *s.Start
	for _, move := range strings.Fields(s.Moves) {
		Move(&cube, move)
		moves = append(moves, move)
		if hasF2LSolved(&cube) {
			return strings.Join(moves, " "), true
		}
	}
	return "", false
}

func hasF2LSolved(c *gocube.CubieCube) bool {
	for _, edgeIdx := range f2lEdges {
		if c.Edges[edgeIdx].Flip || c.Edges[edgeIdx].Piece != edgeIdx {
			return false
		}
	}
	for _, cornerIdx := range f2lCorners {
		if c.Corners[cornerIdx].Piece != cornerIdx || c.Corners[cornerIdx].Orientation != 1 {
			return false
		}
	}
	return true
}
