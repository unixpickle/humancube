package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/unixpickle/humancube"
	"github.com/unixpickle/num-analysis/kahan"
	"github.com/unixpickle/num-analysis/linalg"
	"github.com/unixpickle/serializer"
	"github.com/unixpickle/sgd"
	"github.com/unixpickle/weakai/neuralnet"
	"github.com/unixpickle/weakai/rnn"
	"github.com/unixpickle/weakai/rnn/seqtoseq"
)

const (
	testingSamples    = 1
	evaluateBatchSize = 20
)

type SolveData struct {
	InSeqs  [][]linalg.Vector
	OutSeqs [][]linalg.Vector
	MoveMap map[string]int
}

func main() {
	if len(os.Args) != 6 {
		fmt.Fprintln(os.Stderr, "Usage:", os.Args[0],
			"data_file network_file step_size sample_count batch_size")
		os.Exit(1)
	}
	if err := RunCommand(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func RunCommand() error {
	stepSize, err := strconv.ParseFloat(os.Args[3], 64)
	if err != nil {
		return errors.New("bad step size")
	}
	trainingCount, err := strconv.Atoi(os.Args[4])
	if err != nil {
		return errors.New("bad training count")
	}
	batchSize, err := strconv.Atoi(os.Args[5])
	if err != nil {
		return errors.New("bad batch size")
	}
	return Train(os.Args[1], os.Args[2], stepSize, trainingCount, batchSize)
}

func Train(solveFile, outFile string, stepSize float64, trainingCount, batchSize int) error {
	data, err := generateSolveData(solveFile)
	if err != nil {
		return err
	}

	crossIn := data.InSeqs[trainingCount : trainingCount+testingSamples]
	crossOut := data.OutSeqs[trainingCount : trainingCount+testingSamples]
	trainingIn := data.InSeqs[:trainingCount]
	trainingOut := data.OutSeqs[:trainingCount]
	sampleSet := makeSampleSet(trainingIn, trainingOut)

	net, err := humancube.ReadNetwork(outFile)
	if err != nil {
		log.Println("Could not load network. Creating new one.")
		rand.Seed(time.Now().UnixNano())
		net = humancube.NewNetwork(len(data.InSeqs[0][0]), data.MoveMap)
	} else {
		log.Println("Loaded existing network from file.")
	}

	gradienter := &sgd.Adam{
		Gradienter: &seqtoseq.Gradienter{
			SeqFunc:  &rnn.BlockSeqFunc{B: net.Block},
			Learner:  net.Block,
			CostFunc: neuralnet.DotCost{},
		},
	}

	log.Println("Training (Ctrl+C to finish)...")

	epochIdx := 0
	net.Dropout(true)
	sgd.SGDInteractive(gradienter, sampleSet, stepSize, batchSize, func() bool {
		net.Dropout(false)
		defer net.Dropout(true)
		log.Printf("Epoch %d: error=%f, correct=%f, cross=%f", epochIdx,
			totalError(trainingIn, trainingOut, net),
			totalCorrect(trainingIn, trainingOut, net),
			totalCorrect(crossIn, crossOut, net))
		epochIdx++
		return true
	})

	net.Dropout(false)

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

	var solves []humancube.ReconstructedSolve
	if err := json.Unmarshal(data, &solves); err != nil {
		return nil, err
	}

	shuffled := make([]humancube.ReconstructedSolve, len(solves))
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

func generateSolveVec(r humancube.ReconstructedSolve,
	moveMap map[string]int) (ins, outs []linalg.Vector) {
	cube, _ := humancube.CubeForMoves(r.Scramble)

	for _, move := range strings.Fields(r.Reconstruction) {
		outputVec := make(linalg.Vector, len(moveMap))
		outputVec[moveMap[move]] = 1
		inputVec := humancube.CubeVector(cube)
		humancube.Move(cube, move)
		ins = append(ins, inputVec)
		outs = append(outs, outputVec)
	}

	return
}

func usableSolves(solves []humancube.ReconstructedSolve) []humancube.ReconstructedSolve {
	var res []humancube.ReconstructedSolve

SolveLoop:
	for _, solve := range solves {
		cube, err := humancube.CubeForMoves(solve.Scramble)
		if err != nil {
			continue
		}
		for _, move := range strings.Fields(solve.Reconstruction) {
			if err := humancube.Move(cube, move); err != nil {
				continue SolveLoop
			}
		}
		if cube.Solved() {
			res = append(res, solve)
		}
	}

	return res
}

func makeSampleSet(ins, outs [][]linalg.Vector) sgd.SampleSet {
	res := make(sgd.SliceSampleSet, len(ins))
	for i, in := range ins {
		res[i] = seqtoseq.Sample{Inputs: in, Outputs: outs[i]}
	}
	return res
}

func totalError(inSeqs, outSeqs [][]linalg.Vector, n *humancube.Network) float64 {
	var res kahan.Summer64
	evaluateAll(inSeqs, outSeqs, n, func(actual, expected linalg.Vector) {
		res.Add(actual.Dot(expected))
	})
	return -res.Sum()
}

func totalCorrect(inSeqs, outSeqs [][]linalg.Vector, n *humancube.Network) float64 {
	var numCorrect int
	var numTotal int
	evaluateAll(inSeqs, outSeqs, n, func(actual, expected linalg.Vector) {
		_, max1 := actual.Max()
		_, max2 := expected.Max()
		if max1 == max2 {
			numCorrect++
		}
		numTotal++
	})
	return float64(numCorrect) / float64(numTotal)
}

func evaluateAll(inSeqs, outSeqs [][]linalg.Vector, n *humancube.Network,
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