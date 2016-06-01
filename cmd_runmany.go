package main

import (
	"fmt"
	"log"

	"github.com/unixpickle/gocube"
)

const ScramblePrintInterval = 50

func RunManyCmd(netFile string) error {
	net, err := ReadNetwork(netFile)
	if err != nil {
		return err
	}

	log.Println("Running on scrambles until one gets solved.")

	runIdx := 0
	for {
		cube := gocube.RandomCubieCube()
		for i := 0; i < MaxRunLength; i++ {
			res := net.RNN.StepTime(CubeVector(&cube))
			move := MoveForOutput(net, res)
			Move(&cube, move)
			if cube.Solved() {
				fmt.Println("Solved cube after", runIdx, "tries.")
				return nil
			}
		}
		net.RNN.Reset()
		runIdx++
		if runIdx%ScramblePrintInterval == 0 {
			log.Println("Made", runIdx, "failed solve attempts.")
		}
	}
}
