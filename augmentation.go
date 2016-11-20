package humancube

import (
	"math/rand"
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
	// Crossover specifies how many scrambles to generate by
	// performing "genetic" crossover on reconstructions.
	// The Crossover stage is performed first, before the
	// other augmentation stages.
	Crossover int

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
	s.Samples = append(s.Samples, augmentCrossover(s, params.Crossover)...)
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

type crossoverTransition struct {
	newState string
	moves    []string
}

func augmentCrossover(s *SampleSet, count int) []Sample {
	if len(s.Samples) == 0 {
		return nil
	}

	transitions := map[string][]crossoverTransition{}
	for _, sample := range s.Samples {
		cube := *sample.Start
		var mostSolved int
		var lastState string
		var lastMoves int
		sampleMoves := strings.Fields(sample.Moves)
		for j, move := range sampleMoves {
			Move(&cube, move)
			state := describeF2LPairs(&cube)
			if numSolvedPairs(state) <= mostSolved {
				continue
			}
			mostSolved = numSolvedPairs(state)
			trans := crossoverTransition{
				newState: state,
				moves:    sampleMoves[lastMoves : j+1],
			}
			transitions[lastState] = append(transitions[lastState], trans)
			lastMoves = j + 1
			lastState = state
		}
		finalTrans := crossoverTransition{
			newState: "done",
			moves:    sampleMoves[lastMoves:],
		}
		transitions[lastState] = append(transitions[lastState], finalTrans)
	}

	var res []Sample
	for i := 0; i < count; i++ {
		var moves []string
		var state string
		for state != "done" {
			transOptions := transitions[state]
			t := transOptions[rand.Intn(len(transOptions))]
			state = t.newState
			moves = append(moves, t.moves...)
		}
		cube := gocube.SolvedCubieCube()
		for i := len(moves) - 1; i >= 0; i-- {
			MoveInverse(&cube, moves[i])
		}
		res = append(res, Sample{
			Start: &cube,
			Moves: strings.Join(moves, " "),
		})
	}
	return res
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

// describeF2LPairs returns a string of four 1s or 0s,
// each of which indicates if an F2L pair is solved.
// If the cross is not solved, it returns "".
func describeF2LPairs(c *gocube.CubieCube) string {
	if !hasCrossSolved(c) {
		return ""
	}
	var desc string
	edges := []int{1, 3, 7, 9}
	corners := []int{5, 4, 1, 0}
	for i, edge := range edges {
		corner := corners[i]
		if c.Corners[corner].Orientation != 1 || c.Corners[corner].Piece != corner ||
			c.Edges[edge].Flip || c.Edges[edge].Piece != edge {
			desc += "0"
		} else {
			desc += "1"
		}
	}
	return desc
}

func numSolvedPairs(desc string) int {
	var res int
	for _, x := range desc {
		if x == '1' {
			res++
		}
	}
	return res
}
