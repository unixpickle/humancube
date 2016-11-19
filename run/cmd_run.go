package main

import (
	"errors"
	"fmt"
	"math"
	"math/rand"

	"github.com/unixpickle/humancube"
	"github.com/unixpickle/num-analysis/linalg"
	"github.com/unixpickle/weakai/rnn"
)

const MaxRunLength = 200

func RunCmd(netFile, scramble string) error {
	cube, err := humancube.CubeForMoves(scramble)
	if err != nil {
		return errors.New("bad scramble: " + err.Error())
	}

	net, err := humancube.ReadNetwork(netFile)
	if err != nil {
		return err
	}

	fmt.Println("Moves:")

	runner := &rnn.Runner{Block: net.Block}
	for i := 0; i < MaxRunLength; i++ {
		res := runner.StepTime(humancube.CubeVector(cube))
		move := randomMove(net, res)
		fmt.Print(move + " ")
		humancube.Move(cube, move)
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

func randomMove(n *humancube.Network, out linalg.Vector) string {
	num := rand.Float64()
	outIdx := 0
	for outIdx < len(out)-1 && num > 0 {
		num -= math.Exp(out[outIdx])
		if num > 0 {
			outIdx++
		}
	}

	for name, idx := range n.MoveMap {
		if idx == outIdx {
			return name
		}
	}
	return "?"
}
