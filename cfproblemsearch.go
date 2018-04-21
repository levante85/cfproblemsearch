package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

func main() {
	tflag, err := commandLine()
	switch {
	case tflag == nil && err == nil:
		return
	case tflag != nil && err == nil:
		break
	case tflag == nil && err != nil:
		fmt.Println(err)
		return
	}

	resp, err := readCfProblems(tflag)
	if err != nil {
		fmt.Println(err)
		return
	}

	var wg sync.WaitGroup
	cmd := exec.Command("/usr/bin/less")

	r, stdin := io.Pipe()

	cmd.Stdin = r
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	wg.Add(1)
	go func(wg *sync.WaitGroup, stdin io.WriteCloser) {
		defer stdin.Close()

		if resp.Status == "OK" {
			for i := len(resp.Result.Problems) - 1; i > 0; i-- {
				p := resp.Result.Problems[i]
				if p.Type != "PROGRAMMING" {
					continue
				}

				fmt.Fprintf(stdin, "|%-4d%-2s|%-50s|%+v\n", p.ContestID, p.Index, p.Name, p.Tags)
			}
		} else {
			fmt.Println("Something went wrong request failed...")
		}

		wg.Done()
	}(&wg, stdin)

	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	wg.Wait()
}

func readCfProblems(tflag *string) (*cfResponse, error) {
	const api = "http://codeforces.com/api/problemset.problems?tags=%v"

	c := &http.Client{}
	r, err := c.Get(fmt.Sprintf(api, *tflag))
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return nil, err
	}

	var resp cfResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func commandLine() (*string, error) {
	tagFlag := flag.String("tag", "", "problem type ")
	tagList := flag.Bool("list", false, "list problem tags")
	flag.Parse()

	switch {
	case *tagFlag == "" && flag.NFlag() == 0:
		return nil, errors.New("Tag argument is obligatory to search for problems")
	case *tagList == true:
		fmt.Println("List of valid tags: ")
		for r := bufio.NewReader(strings.NewReader(tags)); ; {
			l1, _, err := r.ReadLine()
			if err != nil && err != io.EOF {
				return nil, err
			}

			l2, _, err := r.ReadLine()
			if err != nil && err != io.EOF {
				return nil, err
			}

			if io.EOF == err {
				break
			}

			fmt.Printf("\t|%-30s|%-30s|\n", l1, l2)
		}

		return nil, nil
	}

	return tagFlag, nil
}

func getTerminalSize() (int, int) {
	type winsize struct {
		Row    uint16
		Col    uint16
		Xpixel uint16
		Ypixel uint16
	}

	ws := &winsize{}
	retCode, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))

	if int(retCode) == -1 {
		panic(errno)
	}
	return int(ws.Row), int(ws.Col)
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
