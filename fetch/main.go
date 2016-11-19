package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/unixpickle/humancube"
)

const LogInterval = 100

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "output_file")
		os.Exit(1)
	}
	outFile := os.Args[1]
	data := []humancube.ReconstructedSolve{}

	ch, errs := humancube.FetchReconstructions()
	for rec := range ch {
		data = append(data, rec)
		if len(data)%LogInterval == 0 {
			log.Println("Down to solve ID", rec.ID)
		}
	}

	if err := <-errs; err != nil {
		fmt.Fprintln(os.Stderr, "Error while scraping:", err)
	}

	binData, _ := json.Marshal(data)
	if err := ioutil.WriteFile(outFile, binData, 0755); err != nil {
		fmt.Fprintln(os.Stderr, "Error saving result:", err)
		os.Exit(1)
	}
}
