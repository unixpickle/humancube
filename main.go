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
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func dieUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <command> <options>\n\n"+
		"Available commands are:\n\n"+
		" scrape <output.json>"+
		"\n\n", os.Args[0])
	os.Exit(1)
}
