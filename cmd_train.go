package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/unixpickle/num-analysis/kahan"
	"github.com/unixpickle/num-analysis/linalg"
	"github.com/unixpickle/serializer"
	"github.com/unixpickle/weakai/neuralnet"
	"github.com/unixpickle/weakai/rnn"
)

const (
	testingSamples    = 200
	evaluateBatchSize = 20
)

type SolveData struct {
	InSeqs  [][]linalg.Vector
	OutSeqs [][]linalg.Vector
	MoveMap map[string]int
}

func TrainCmd(solveFile, outFile string, stepSize float64, trainingCount, batchSize int) error {
	data, err := generateSolveData(solveFile)
	if err != nil {
		return err
	}

	crossIn := data.InSeqs[trainingCount : trainingCount+testingSamples]
	crossOut := data.OutSeqs[trainingCount : trainingCount+testingSamples]
	trainingIn := data.InSeqs[:trainingCount]
	trainingOut := data.OutSeqs[:trainingCount]
	sampleSet := makeSampleSet(trainingIn, trainingOut)

	net, err := ReadNetwork(outFile)
	if err != nil {
		log.Println("Could not load network. Creating new one.")
		rand.Seed(time.Now().UnixNano())
		net = NewNetwork(len(data.InSeqs[0][0]), data.MoveMap)
	} else {
		log.Println("Loaded existing network from file.")
	}

	gradienter := &neuralnet.Equilibration{
		RGradienter: &rnn.FullRGradienter{
			Learner:  net.Block,
			CostFunc: neuralnet.MeanSquaredCost{},
		},
		Learner:        net.Block,
		Memory:         0.5,
		NumSamples:     10,
		UpdateInterval: 50,
	}

	log.Println("Training (Ctrl+C to finish)...")

	epochIdx := 0
	neuralnet.SGDInteractive(gradienter, sampleSet, stepSize, batchSize, func() bool {
		log.Printf("Epoch %d: error=%f, correct=%f, cross=%f", epochIdx,
			totalError(trainingIn, trainingOut, net),
			totalCorrect(trainingIn, trainingOut, net),
			totalCorrect(crossIn, crossOut, net))
		epochIdx++
		return true
	})

	log.Println("Saving...")
	saved, err := serializer.SerializeWithType(net)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(outFile, saved, 0755)
}

func generateSolveData(solveFile string) (*SolveData, error) {
	data, err := ioutil.ReadFile(solveFile)
	if err != nil {
		return nil, err
	}

	var solves []ReconstructedSolve
	if err := json.Unmarshal(data, &solves); err != nil {
		return nil, err
	}

	shuffled := make([]ReconstructedSolve, len(solves))
	for i, j := range rand.Perm(len(solves)) {
		shuffled[i] = solves[j]
	}
	solves = usableSolves(shuffled)

	moveWords := map[string]int{}
	for _, solve := range solves {
		for _, move := range strings.Fields(solve.Scramble + " " + solve.Reconstruction) {
			if _, ok := moveWords[move]; !ok {
				moveWords[move] = len(moveWords)
			}
		}
	}

	var ins, outs [][]linalg.Vector
	for _, solve := range solves {
		i, o := generateSolveVec(solve, moveWords)
		ins = append(ins, i)
		outs = append(outs, o)
	}

	return &SolveData{
		InSeqs:  ins,
		OutSeqs: outs,
		MoveMap: moveWords,
	}, nil
}

func generateSolveVec(r ReconstructedSolve, moveMap map[string]int) (ins, outs []linalg.Vector) {
	cube, _ := CubeForMoves(r.Scramble)

	for _, move := range strings.Fields(r.Reconstruction) {
		outputVec := make(linalg.Vector, len(moveMap))
		outputVec[moveMap[move]] = 1
		inputVec := CubeVector(cube)
		Move(cube, move)
		ins = append(ins, inputVec)
		outs = append(outs, outputVec)
	}

	return
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

func makeSampleSet(ins, outs [][]linalg.Vector) neuralnet.SampleSet {
	res := make(neuralnet.SampleSet, len(ins))
	for i, in := range ins {
		res[i] = rnn.Sequence{Inputs: in, Outputs: outs[i]}
	}
	return res
}

func totalError(inSeqs, outSeqs [][]linalg.Vector, n *Network) float64 {
	var res kahan.Summer64
	evaluateAll(inSeqs, outSeqs, n, func(actual, expected linalg.Vector) {
		for k, a := range actual {
			res.Add(math.Pow(a-expected[k], 2))
		}
	})
	return res.Sum()
}

func totalCorrect(inSeqs, outSeqs [][]linalg.Vector, n *Network) float64 {
	var numCorrect int
	var numTotal int
	evaluateAll(inSeqs, outSeqs, n, func(actual, expected linalg.Vector) {
		if MaxValueIndex(actual) == MaxValueIndex(expected) {
			numCorrect++
		}
		numTotal++
	})
	return float64(numCorrect) / float64(numTotal)
}

func evaluateAll(inSeqs, outSeqs [][]linalg.Vector, n *Network,
	f func(actual, expected linalg.Vector)) {
	runner := &rnn.Runner{Block: n.Block}
	for i := 0; i < len(inSeqs); i += evaluateBatchSize {
		batchSize := evaluateBatchSize
		if batchSize > len(inSeqs)-i {
			batchSize = len(inSeqs) - i
		}
		batch := inSeqs[i : i+batchSize]
		output := runner.RunAll(batch)
		for j, outSeq := range outSeqs[i : i+batchSize] {
			for t, actual := range output[j] {
				expected := outSeq[t]
				f(actual, expected)
			}
		}
	}
}
