package main

import (
	"errors"
	"fmt"

	"github.com/unixpickle/num-analysis/linalg"
)

const MaxRunLength = 200

func RunCmd(netFile, scramble string) error {
	cube, err := CubeForMoves(scramble)
	if err != nil {
		return errors.New("bad scramble: " + err.Error())
	}

	net, err := ReadNetwork(netFile)
	if err != nil {
		return err
	}

	fmt.Println("Moves:")

	for i := 0; i < MaxRunLength; i++ {
		res := net.RNN.StepTime(CubeVector(cube))
		move := moveForOutput(net, res)
		fmt.Print(move + " ")
		Move(cube, move)
		if cube.Solved() {
			fmt.Println()
			fmt.Println("Cube solved!")
			return nil
		}
	}

	fmt.Println()
	fmt.Println("Cube not solved after", MaxRunLength, "moves.")

	return nil
}

func moveForOutput(n *Network, out linalg.Vector) string {
	moveIdx := MaxValueIndex(out)

	for name, idx := range n.MoveMap {
		if idx == moveIdx {
			return name
		}
	}
	panic("code should be unreachable")
}
