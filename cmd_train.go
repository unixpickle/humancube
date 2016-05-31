package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/unixpickle/num-analysis/kahan"
	"github.com/unixpickle/num-analysis/linalg"
	"github.com/unixpickle/weakai/rnn"
	"github.com/unixpickle/weakai/rnn/lstm"
	"github.com/unixpickle/weakai/rnn/softmax"
)

const (
	lstmStepSize   = 0.01
	lstmEpochs     = 1000
	lstmHiddenSize = 100
)

func TrainCmd(solveFile, outFile string) error {
	rand.Seed(time.Now().UnixNano())

	ins, outs, err := generateSolveVecs(solveFile)
	if err != nil {
		return err
	}

	trainer := rnn.Trainer{
		InSeqs:   ins[:10],
		OutSeqs:  outs[:10],
		CostFunc: rnn.MeanSquaredCost{},
		StepSize: lstmStepSize,
		Epochs:   1,
	}
	lstmNet := lstm.NewNet(rnn.ReLU{}, len(ins[0][0]), lstmHiddenSize, len(outs[0][0]))
	softmaxLayer := softmax.NewSoftmax(len(outs[0][0]))
	net := rnn.DeepRNN{lstmNet, softmaxLayer}
	net.Randomize()
	log.Println("Training...")
	for i := 0; i < lstmEpochs; i++ {
		log.Printf("Epoch %d: error=%f, correct=%f", i, totalError(&trainer, net),
			totalCorrect(&trainer, net))
		trainer.Train(net)
		if i == 60 {
			trainer.StepSize = 0.001
		}
	}
	log.Println("Saving...")

	return errors.New("saving not yet implemented")
}

func generateSolveVecs(solveFile string) (ins, outs [][]linalg.Vector, err error) {
	data, err := ioutil.ReadFile(solveFile)
	if err != nil {
		return nil, nil, err
	}

	var solves []ReconstructedSolve
	if err := json.Unmarshal(data, &solves); err != nil {
		return nil, nil, err
	}

	solves = usableSolves(solves)

	moveWords := map[string]int{}
	for _, solve := range solves {
		for _, move := range strings.Fields(solve.Scramble + " " + solve.Reconstruction) {
			if _, ok := moveWords[move]; !ok {
				moveWords[move] = len(moveWords)
			}
		}
	}

	for _, solve := range solves {
		i, o := generateSolveVec(solve, moveWords)
		ins = append(ins, i)
		outs = append(outs, o)
	}

	return
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

func totalError(t *rnn.Trainer, r rnn.RNN) float64 {
	var res kahan.Summer64
	for i, inSeq := range t.InSeqs {
		outSeq := t.OutSeqs[i]
		for j, inVec := range inSeq {
			outVec := outSeq[j]
			actual := r.StepTime(inVec)
			for k, a := range actual {
				res.Add(math.Pow(a-outVec[k], 2))
			}
		}
		r.Reset()
	}
	return res.Sum()
}

func totalCorrect(t *rnn.Trainer, r rnn.RNN) float64 {
	var numCorrect int
	var numTotal int
	for i, inSeq := range t.InSeqs {
		outSeq := t.OutSeqs[i]
		for j, inVec := range inSeq {
			outVec := outSeq[j]
			actual := r.StepTime(inVec)
			if maxValueIndex(actual) == maxValueIndex(outVec) {
				numCorrect++
			}
			numTotal++
		}
		r.Reset()
	}
	return float64(numCorrect) / float64(numTotal)
}

func maxValueIndex(v linalg.Vector) int {
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
