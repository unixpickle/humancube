package main

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var solveLinkExp = regexp.MustCompile("^/solve/([0-9]*)$")

type ReconstructedSolve struct {
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
			if err != nil {
				errChan <- err
				return
			} else if len(links) == 0 {
				return
			}
			for _, id := range links {
				solve, err := fetchReconstruction(id)
				if err != nil {
					errChan <- err
					return
				}
				resChan <- *solve
			}
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
	res.Scramble = rawAlgText(algWells[0])
	res.Reconstruction = rawAlgText(algWells[1])
	res.Commented = commentedAlgText(algWells[2])

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
	return strings.TrimSpace(res)
}

func commentedAlgText(n *html.Node) string {
	var res string
	child := n.FirstChild
	for child != nil {
		if child.Type == html.TextNode {
			res += child.Data
		} else if child.Type == html.ElementNode {
			if child.DataAtom == atom.Br {
				res += "\n"
			} else if child.DataAtom == atom.Span {
				res += scrape.Text(child)
			}
		}
		child = child.NextSibling
	}
	return strings.TrimSpace(res)
}
