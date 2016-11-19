package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/unixpickle/gocube"
	"github.com/unixpickle/humancube"
	"github.com/unixpickle/num-analysis/linalg"
	"github.com/unixpickle/serializer"
	"github.com/unixpickle/sgd"
	"github.com/unixpickle/weakai/neuralnet"
	"github.com/unixpickle/weakai/rnn"
	"github.com/unixpickle/weakai/rnn/seqtoseq"
)

const (
	ValidationAmount = 0.1
)

type SolveData struct {
	InSeqs  [][]linalg.Vector
	OutSeqs [][]linalg.Vector
	MoveMap map[string]int
}

func main() {
	if len(os.Args) != 5 {
		fmt.Fprintln(os.Stderr, "Usage:", os.Args[0],
			"data_file network_file step_size batch_size")
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
	batchSize, err := strconv.Atoi(os.Args[4])
	if err != nil {
		return errors.New("bad batch size")
	}
	return Train(os.Args[1], os.Args[2], stepSize, batchSize)
}

func Train(solveFile, outFile string, stepSize float64, batchSize int) error {
	sampleSet, err := humancube.LoadSampleSet(solveFile)
	if err != nil {
		return errors.New("load sample set: " + err.Error())
	}

	// It is important to split before augmenting, to ensure
	// that the validation set doesn't include samples which
	// are closely related to the training set.
	validation, training := sgd.HashSplit(sampleSet, ValidationAmount)
	if validation.Len() < batchSize || training.Len() < batchSize {
		return errors.New("not enough samples")
	}
	validation = validation.Copy()
	training = training.Copy()

	log.Printf("Augmenting %d training and %d validation...", training.Len(), validation.Len())
	// Augment the samples the same way every time.
	rand.Seed(123123)
	augParams := &humancube.AugmentParams{
		LLCases:    10,
		CrossSkips: true,
		FirstSkips: true,
	}
	humancube.Augment(validation.(*humancube.SampleSet), augParams)
	humancube.Augment(training.(*humancube.SampleSet), augParams)
	log.Printf("Using %d training and %d validation...", training.Len(), validation.Len())

	net, err := humancube.ReadNetwork(outFile)
	if err != nil {
		log.Println("Could not load network. Creating new one.")
		rand.Seed(time.Now().UnixNano())
		solved := gocube.SolvedCubieCube()
		inLen := len(humancube.CubeVector(&solved))
		net = humancube.NewNetwork(inLen, sampleSet.MoveMap)
	} else {
		log.Println("Loaded existing network from file.")
	}

	costFunc := neuralnet.DotCost{}
	gradienter := &sgd.Adam{
		Gradienter: &seqtoseq.Gradienter{
			SeqFunc:  &rnn.BlockSeqFunc{B: net.Block},
			Learner:  net.Block,
			CostFunc: costFunc,
			MaxLanes: batchSize,
		},
	}

	log.Println("Training (Ctrl+C to finish)...")

	net.Dropout(true)
	var epochIdx int
	var lastBatch sgd.SampleSet
	sgd.SGDMini(gradienter, training, stepSize, batchSize, func(s sgd.SampleSet) bool {
		net.Dropout(false)
		defer net.Dropout(true)

		if epochIdx%4 == 0 {
			var lastCost float64
			if lastBatch != nil {
				lastCost = seqtoseq.TotalCostBlock(net.Block, batchSize, lastBatch, costFunc)
			}
			lastBatch = s

			batchCost := seqtoseq.TotalCostBlock(net.Block, batchSize, s, costFunc)

			sgd.ShuffleSampleSet(validation)
			subVal := validation.Subset(0, batchSize)
			validationCost := seqtoseq.TotalCostBlock(net.Block, batchSize, subVal, costFunc)

			log.Printf("Epoch %d: validation=%f cost=%f last=%f", epochIdx, validationCost,
				batchCost, lastCost)
		}

		epochIdx++
		return true
	})

	net.Dropout(false)

	log.Println("Saving...")
	return serializer.SaveAny(outFile, net)
}
