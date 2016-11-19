package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		dieUsage()
	}
	var cmdErr error
	if len(os.Args) == 3 {
		cmdErr = RunCmd(os.Args[1], os.Args[2])
	} else if len(os.Args) == 2 {
		cmdErr = RunManyCmd(os.Args[1])
	} else {
		dieUsage()
	}
	if cmdErr != nil {
		fmt.Fprintln(os.Stderr, cmdErr)
	}
}

func dieUsage() {
	fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "network_file [scramble]")
	os.Exit(1)
}
