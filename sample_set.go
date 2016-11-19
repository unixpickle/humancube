package humancube

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/unixpickle/gocube"
	"github.com/unixpickle/num-analysis/linalg"
	"github.com/unixpickle/sgd"
	"github.com/unixpickle/weakai/rnn/seqtoseq"
)

// A Sample details a solve or part of a solve.
type Sample struct {
	Start *gocube.CubieCube
	Moves string
}

// A SampleSet is an sgd.SampleSet of Samples.
type SampleSet struct {
	Samples []Sample
	MoveMap map[string]int
}

// NewSampleSet creates a SampleSet with all of the valid
// solves in a list of reconstructions.
// It does not apply any form of data augmentation.
func NewSampleSet(r []ReconstructedSolve) *SampleSet {
	solves := usableSolves(r)
	moveWords := map[string]int{}
	for _, solve := range solves {
		for _, move := range strings.Fields(solve.Scramble + " " + solve.Reconstruction) {
			if _, ok := moveWords[move]; !ok {
				moveWords[move] = len(moveWords)
			}
		}
	}
	res := &SampleSet{
		MoveMap: moveWords,
		Samples: make([]Sample, 0, len(solves)),
	}
	for _, solve := range solves {
		cube, _ := CubeForMoves(solve.Scramble)
		res.Samples = append(res.Samples, Sample{
			Start: cube,
			Moves: solve.Reconstruction,
		})
	}
	return res
}

// LoadSampleSet is like NewSampleSet, but it loads the
// reconstructions from a file.
func LoadSampleSet(f string) (*SampleSet, error) {
	contents, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}
	var r []ReconstructedSolve
	if err := json.Unmarshal(contents, &r); err != nil {
		return nil, err
	}
	return NewSampleSet(r), nil
}

// Len returns the number of samples.
func (s *SampleSet) Len() int {
	return len(s.Samples)
}

// Swap swaps two samples.
func (s *SampleSet) Swap(i, j int) {
	s.Samples[i], s.Samples[j] = s.Samples[j], s.Samples[i]
}

// GetSample generates a seqtoseq.Sample for the given
// sample index.
func (s *SampleSet) GetSample(idx int) interface{} {
	sample := s.Samples[idx]
	cube := *sample.Start

	var ins, outs []linalg.Vector
	for _, move := range strings.Fields(sample.Moves) {
		outputVec := make(linalg.Vector, len(s.MoveMap))
		outputVec[s.MoveMap[move]] = 1
		inputVec := CubeVector(&cube)
		Move(&cube, move)
		ins = append(ins, inputVec)
		outs = append(outs, outputVec)
	}

	return seqtoseq.Sample{Inputs: ins, Outputs: outs}
}

// Hash generates a hash for a sample.
func (s *SampleSet) Hash(idx int) []byte {
	return s.GetSample(idx).(seqtoseq.Sample).Hash()
}

// Subset returns a subset of the sample set.
func (s *SampleSet) Subset(i, j int) sgd.SampleSet {
	return &SampleSet{
		Samples: s.Samples[i:j],
		MoveMap: s.MoveMap,
	}
}

// Copy returns a copy of the sample set.
func (s *SampleSet) Copy() sgd.SampleSet {
	return &SampleSet{
		Samples: append([]Sample{}, s.Samples...),
		MoveMap: s.MoveMap,
	}
}

func usableSolves(solves []ReconstructedSolve) []ReconstructedSolve {
	var res []ReconstructedSolve

SolveLoop:
	for _, solve := range solves {
		cube, err := CubeForMoves(solve.Scramble)
		if err != nil {
			continue
		}
		for _, move := range strings.Fields(solve.Reconstruction) {
			if err := Move(cube, move); err != nil {
				continue SolveLoop
			}
		}
		if cube.Solved() {
			res = append(res, solve)
		}
	}

	return res
}
