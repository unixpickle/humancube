package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"time"

	"github.com/unixpickle/num-analysis/kahan"
	"github.com/unixpickle/num-analysis/linalg"
	"github.com/unixpickle/serializer"
	"github.com/unixpickle/weakai/rnn"
)

const (
	trainingSamples = 1000
	testingSamples  = 200
)

type SolveData struct {
	InSeqs  [][]linalg.Vector
	OutSeqs [][]linalg.Vector
	MoveMap map[string]int
}

func TrainCmd(solveFile, outFile string, stepSize float64) error {
	data, err := generateSolveData(solveFile)
	if err != nil {
		return err
	}

	trainer := rnn.Trainer{
		InSeqs:   data.InSeqs[:trainingSamples],
		OutSeqs:  data.OutSeqs[:trainingSamples],
		CostFunc: rnn.MeanSquaredCost{},
		StepSize: stepSize,
		Epochs:   1,
	}

	crossIn := data.InSeqs[trainingSamples : trainingSamples+testingSamples]
	crossOut := data.OutSeqs[trainingSamples : trainingSamples+testingSamples]

	net, err := ReadNetwork(outFile)
	if err != nil {
		log.Println("Could not load network. Creating new one.")
		rand.Seed(time.Now().UnixNano())
		net = NewNetwork(len(data.InSeqs[0][0]), data.MoveMap)
	} else {
		log.Println("Loaded existing network from file.")
	}

	log.Println("Training (Ctrl+C to finish)...")

	var killed uint32
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		signal.Stop(c)
		atomic.StoreUint32(&killed, 1)
		fmt.Println("\nCaught interrupt. Ctrl+C again to terminate.")
	}()

	for i := 0; atomic.LoadUint32(&killed) == 0; i++ {
		log.Printf("Epoch %d: error=%f, correct=%f, cross=%f", i,
			totalError(trainer.InSeqs, trainer.OutSeqs, net),
			totalCorrect(trainer.InSeqs, trainer.OutSeqs, net),
			totalCorrect(crossIn, crossOut, net))
		trainer.Train(net.RNN)
	}

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

func totalError(inSeqs, outSeqs [][]linalg.Vector, n *Network) float64 {
	var res kahan.Summer64
	for i, inSeq := range inSeqs {
		outSeq := outSeqs[i]
		for j, inVec := range inSeq {
			outVec := outSeq[j]
			actual := n.RNN.StepTime(inVec)
			for k, a := range actual {
				res.Add(math.Pow(a-outVec[k], 2))
			}
		}
		n.RNN.Reset()
	}
	return res.Sum()
}

func totalCorrect(inSeqs, outSeqs [][]linalg.Vector, n *Network) float64 {
	var numCorrect int
	var numTotal int
	for i, inSeq := range inSeqs {
		outSeq := outSeqs[i]
		for j, inVec := range inSeq {
			outVec := outSeq[j]
			actual := n.RNN.StepTime(inVec)
			if maxValueIndex(actual) == maxValueIndex(outVec) {
				numCorrect++
			}
			numTotal++
		}
		n.RNN.Reset()
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
