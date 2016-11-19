package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/unixpickle/humancube"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "data_file")
		os.Exit(1)
	}
	data, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Read data:", err)
		os.Exit(1)
	}

	var solves []humancube.ReconstructedSolve
	if err := json.Unmarshal(data, &solves); err != nil {
		fmt.Fprintln(os.Stderr, "Unmarshal data:", err)
		os.Exit(1)
	}

	var totalEntries int
	var parsedScrambles int
	var parsedReconstructions int
	var correctReconstructions int

SolveLoop:
	for _, solve := range solves {
		totalEntries++
		cube, err := humancube.CubeForMoves(solve.Scramble)
		if err != nil {
			continue
		}
		parsedScrambles++
		for _, move := range strings.Fields(solve.Reconstruction) {
			if err := humancube.Move(cube, move); err != nil {
				continue SolveLoop
			}
		}
		parsedReconstructions++
		if cube.Solved() {
			correctReconstructions++
		}
	}

	fmt.Println("Processed", totalEntries, "data items:")
	fmt.Println("  Valid scrambles:", parsedScrambles)
	fmt.Println("  Valid solutions:", parsedReconstructions)
	fmt.Println("Correct solutions:", correctReconstructions)
}
