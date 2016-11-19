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
	// LLCases specifies the number of different last layer
	// cases to provide for a single F2L solve.
	LLCases int

	// CrossSkips enables samples which start with a pre-made
	// cross.
	CrossSkips bool

	// FirstSkips enables samples which have the first move
	// already made.
	FirstSkips bool
}

// Augment generates new samples based on a SampleSet of
// complete solves.
func Augment(s *SampleSet, params *AugmentParams) {
	s.Samples = append(s.Samples, augmentLastLayer(s, params.LLCases)...)
	var extra []Sample
	if params.CrossSkips {
		extra = append(extra, augmentCrossSkips(s)...)
	}
	if params.FirstSkips {
		extra = append(extra, augmentFirstSkips(s)...)
	}
	s.Samples = append(s.Samples, extra...)
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

func augmentCrossSkips(orig *SampleSet) []Sample {
	var res []Sample
	for _, sample := range orig.Samples {
		if hasCrossSolved(sample.Start) {
			continue
		}
		cube := *sample.Start
		moves := strings.Fields(sample.Moves)
		for i, move := range moves {
			Move(&cube, move)
			if hasCrossSolved(&cube) {
				res = append(res, Sample{
					Moves: strings.Join(moves[i+1:], " "),
					Start: &cube,
				})
				break
			}
		}
	}
	return res
}

func augmentFirstSkips(orig *SampleSet) []Sample {
	var res []Sample
	for _, sample := range orig.Samples {
		cube := *sample.Start
		if hasCrossSolved(&cube) {
			continue
		}
		Move(&cube, strings.Fields(sample.Moves)[0])
		if hasCrossSolved(&cube) {
			continue
		}
		res = append(res, Sample{
			Moves: strings.Join(strings.Fields(sample.Moves)[1:], " "),
			Start: &cube,
		})
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

func hasCrossSolved(c *gocube.CubieCube) bool {
	for _, edgeIdx := range []int{2, 8, 10, 11} {
		edge := c.Edges[edgeIdx]
		if edge.Flip || edge.Piece != edgeIdx {
			return false
		}
	}
	return true
}
