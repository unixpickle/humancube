package main

import (
	"fmt"
	"log"

	"github.com/unixpickle/gocube"
	"github.com/unixpickle/humancube"
	"github.com/unixpickle/weakai/rnn"
)

const ScramblePrintInterval = 50

func RunManyCmd(netFile string) error {
	net, err := humancube.ReadNetwork(netFile)
	if err != nil {
		return err
	}

	log.Println("Running on scrambles until one gets solved.")

	runIdx := 0
	runner := &rnn.Runner{Block: net.Block}
	for {
		cube := gocube.RandomCubieCube()
		for i := 0; i < MaxRunLength; i++ {
			res := runner.StepTime(humancube.CubeVector(&cube))
			move := randomMove(net, res)
			humancube.Move(&cube, move)
			if cube.Solved() {
				fmt.Println("Solved cube after", runIdx, "tries.")
				return nil
			}
		}
		runner.Reset()
		runIdx++
		if runIdx%ScramblePrintInterval == 0 {
			log.Println("Made", runIdx, "failed solve attempts.")
		}
	}
}
