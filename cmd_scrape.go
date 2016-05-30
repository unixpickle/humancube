package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

const LogInterval = 100

func ScrapeCmd(outputFile string) error {
	var data []ReconstructedSolve

	ch, errs := FetchReconstructions()
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
	return ioutil.WriteFile(outputFile, binData, 0755)
}
