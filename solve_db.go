package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var solveLinkExp = regexp.MustCompile("^/solve/([0-9]*)$")

const fetchRoutineCount = 10

type ReconstructedSolve struct {
	ID             int
	Scramble       string
	Reconstruction string
	Commented      string
}

func FetchReconstructions() (<-chan ReconstructedSolve, <-chan error) {
	resChan := make(chan ReconstructedSolve)
	errChan := make(chan error, 1)

	go func() {
		defer close(resChan)
		defer close(errChan)
		page := 1
		for {
			links, err := linksOnPage(page)
			page++
			if err != nil {
				errChan <- err
				return
			} else if len(links) == 0 {
				return
			}
			fetchPages(links, resChan)
		}
	}()

	return resChan, errChan
}

func linksOnPage(page int) ([]int, error) {
	pageURL := fmt.Sprintf("http://cubesolv.es/?page=%d", page)
	resp, err := http.Get(pageURL)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	parsed, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}
	links := scrape.FindAll(parsed, scrape.ByTag(atom.A))

	var res []int
	for _, tag := range links {
		href := scrape.Attr(tag, "href")
		parsed := solveLinkExp.FindStringSubmatch(href)
		if parsed == nil {
			continue
		}
		id, _ := strconv.Atoi(parsed[1])
		res = append(res, id)
	}

	return res, nil
}

func fetchPages(ids []int, resChan chan<- ReconstructedSolve) {
	idChan := make(chan int, len(ids))
	for _, id := range ids {
		idChan <- id
	}
	close(idChan)

	var wg sync.WaitGroup

	for i := 0; i < fetchRoutineCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for id := range idChan {
				solve, err := fetchReconstruction(id)
				if err != nil {
					log.Printf("Error fetching %d: %s", id, err)
				} else {
					resChan <- *solve
				}
			}
		}()
	}

	wg.Wait()
}

func fetchReconstruction(id int) (*ReconstructedSolve, error) {
	pageURL := fmt.Sprintf("http://cubesolv.es/solve/%d", id)
	resp, err := http.Get(pageURL)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	parsed, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	algWells := scrape.FindAll(parsed, scrape.ByClass("algorithm"))
	if len(algWells) != 2 {
		return nil, errors.New("expected exactly 2 algorithm wells")
	}

	var res ReconstructedSolve
	res.ID = id
	res.Scramble = rawAlgText(algWells[0])
	res.Reconstruction = rawAlgText(algWells[1])
	res.Commented = commentedAlgText(algWells[1])

	return &res, nil
}

func rawAlgText(n *html.Node) string {
	var res string
	child := n.FirstChild
	for child != nil {
		if child.Type == html.TextNode {
			res += child.Data
		}
		child = child.NextSibling
	}
	res = strings.Replace(res, "(", "", -1)
	res = strings.Replace(res, ")", "", -1)
	return strings.Join(strings.Fields(res), " ")
}

func commentedAlgText(n *html.Node) string {
	var res string
	child := n.FirstChild
	for child != nil {
		if child.Type == html.TextNode {
			res += removeNewlines(child.Data)
		} else if child.Type == html.ElementNode {
			if child.DataAtom == atom.Br {
				res += "\n"
			} else if child.DataAtom == atom.Span {
				res += removeNewlines(scrape.Text(child))
			}
		}
		child = child.NextSibling
	}
	return strings.TrimSpace(res)
}

func removeNewlines(s string) string {
	return strings.Replace(s, "\n", "", -1)
}
