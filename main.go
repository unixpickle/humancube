package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		dieUsage()
	}
	var cmdErr error
	switch os.Args[1] {
	case "scrape":
		if len(os.Args) != 3 {
			dieUsage()
		}
		cmdErr = ScrapeCmd(os.Args[2])
	case "stats":
		if len(os.Args) != 3 {
			dieUsage()
		}
		cmdErr = StatsCmd(os.Args[2])
	case "train":
		if len(os.Args) != 7 {
			dieUsage()
		}
		stepSize, err := strconv.ParseFloat(os.Args[4], 64)
		if err != nil {
			cmdErr = err
			break
		}
		trainingCount, err := strconv.Atoi(os.Args[5])
		if err != nil {
			cmdErr = err
			break
		}
		batchSize, err := strconv.Atoi(os.Args[6])
		if err != nil {
			cmdErr = err
			break
		}
		cmdErr = TrainCmd(os.Args[2], os.Args[3], stepSize, trainingCount, batchSize)
	case "run":
		if len(os.Args) != 4 {
			dieUsage()
		}
		cmdErr = RunCmd(os.Args[2], os.Args[3])
	case "runmany":
		if len(os.Args) != 3 {
			dieUsage()
		}
		cmdErr = RunManyCmd(os.Args[2])
	default:
		dieUsage()
	}
	if cmdErr != nil {
		fmt.Fprintln(os.Stderr, cmdErr)
		os.Exit(1)
	}
}

func dieUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <command> <options>\n\n"+
		"Available commands are:\n\n"+
		" scrape <output.json>\n"+
		" stats <data.json>\n"+
		" train <data.json> <network file> <stepsize> <sample count> <batch size>\n"+
		" run <network_file> <scramble>\n"+
		" runmany <network_file>"+
		"\n\n", os.Args[0])
	os.Exit(1)
}
