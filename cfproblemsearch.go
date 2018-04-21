package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

var tags = `
implementation
dp
math
greedy
brute force
data structures
constructive algorithms
dfs and similar
sortings
binary search
graphs
trees
strings
number theory
geometry
combinatorics
two pointers
dsu
bitmasks
probabilities
shortest paths
hashing
divide and conquer
games
matrices
flows
string suffix structures
expression parsing
graph matchings
ternary search
meet-in-the-middle
fft
2-sat
chinese remainder theorem
schedules`

func main() {
	tagFlag := flag.String("tag", "", "problem type ")
	tagList := flag.Bool("list", false, "list problem tags")
	flag.Parse()

	switch {
	case *tagFlag == "" && flag.NFlag() == 0:
		fmt.Println("Tag argument is obligatory to in order to search for problems")
		return
	case *tagList == true:
		fmt.Println("List of valid tags: ")
		for r := bufio.NewReader(strings.NewReader(tags)); ; {
			l1, _, err := r.ReadLine()
			if err != nil && err != io.EOF {
				fmt.Println(err)
				return
			}

			l2, _, err := r.ReadLine()
			if err != nil && err != io.EOF {
				fmt.Println(err)
				return
			}

			if io.EOF == err {
				break
			}

			fmt.Printf("\t|%-30s|%-30s|\n", l1, l2)
		}

		return
	}

	const api = "http://codeforces.com/api/problemset.problems?tags=%v"

	c := &http.Client{}
	r, err := c.Get(fmt.Sprintf(api, *tagFlag))
	if err != nil {
		fmt.Println(err)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	var resp cfResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		fmt.Println(err)
		return
	}

	if resp.Status == "OK" {
		for i := len(resp.Result.Problems) - 1; i > 0; i-- {
			p := resp.Result.Problems[i]
			if p.Type != "PROGRAMMING" {
				continue
			}
			fmt.Printf("|%-4d%-2s|%-50s|%+v\n", p.ContestID, p.Index, p.Name, p.Tags)

		}

	} else {
		fmt.Println("Something went wrong request failed...")
	}

}

type cfResponse struct {
	Result struct {
		ProblemStatistics []struct {
			ContestID   int    `json:"contestId"`
			Index       string `json:"index"`
			SolvedCount int    `json:"solvedCount"`
		} `json:"problemStatistics"`
		Problems []struct {
			ContestID int      `json:"contestId"`
			Index     string   `json:"index"`
			Name      string   `json:"name"`
			Points    float64  `json:"points"`
			Tags      []string `json:"tags"`
			Type      string   `json:"type"`
		} `json:"problems"`
	} `json:"result"`
	Status string `json:"status"`
}
