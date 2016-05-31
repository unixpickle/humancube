package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		dieUsage()
	}
	var err error
	switch os.Args[1] {
	case "scrape":
		if len(os.Args) != 3 {
			dieUsage()
		}
		err = ScrapeCmd(os.Args[2])
	case "stats":
		if len(os.Args) != 3 {
			dieUsage()
		}
		err = StatsCmd(os.Args[2])
	case "train":
		if len(os.Args) != 4 {
			dieUsage()
		}
		err = TrainCmd(os.Args[2], os.Args[3])
	default:
		dieUsage()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func dieUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <command> <options>\n\n"+
		"Available commands are:\n\n"+
		" scrape <output.json>\n"+
		" stats <data.json>\n"+
		" train <data.json> <net_out.json>"+
		"\n\n", os.Args[0])
	os.Exit(1)
}
