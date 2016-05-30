package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)

func StatsCmd(dataFile string) error {
	data, err := ioutil.ReadFile(dataFile)
	if err != nil {
		return err
	}

	var solves []ReconstructedSolve
	if err := json.Unmarshal(data, &solves); err != nil {
		return err
	}

	var totalEntries int
	var parsedScrambles int
	var parsedReconstructions int
	var correctReconstructions int

SolveLoop:
	for _, solve := range solves {
		totalEntries++
		cube, err := CubeForMoves(solve.Scramble)
		if err != nil {
			continue
		}
		parsedScrambles++
		for _, move := range strings.Fields(solve.Reconstruction) {
			if err := Move(cube, move); err != nil {
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

	return nil
}
