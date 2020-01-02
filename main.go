package main

import (
	"context"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

var (
	ErrTimesUp = errors.New("Times up!")
)

func main() {
	const (
		defaultCSVPath        = "problems.csv"
		defaultTimeoutSeconds = 30
		defaultShuffle        = false
	)
	var (
		csvPath        string
		timeoutSeconds int
		shuffle        bool
	)

	flag.StringVar(&csvPath, "csv", defaultCSVPath, "Path to a CSV file containing problems.")
	flag.IntVar(&timeoutSeconds, "timeout", defaultTimeoutSeconds, "Time in seconds to complete the quiz.")
	flag.BoolVar(&shuffle, "shuffle", defaultShuffle, "Shuffle the problems.")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	const (
		CORRECT int = iota
		INCORRECT
	)
	answers := map[int][]string{
		CORRECT:   []string{},
		INCORRECT: []string{},
	}

	qs, err := questions(csvPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Relies on the default seed to provide natural order when not shuffling.
	indexOrder := make([]int, 0, len(qs))
	if shuffle {
		rand.Seed(time.Now().UnixNano())
	}
	indexOrder = rand.Perm(len(qs))

	fmt.Println("Type your answer followed by the return key.")
	fmt.Println()

outer:
	for _, n := range indexOrder {
		q := qs[n]

		ans, err := prompt(ctx, q)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			break outer
		}

		if ans == q[1] {
			answers[CORRECT] = append(answers[CORRECT], ans)
		} else {
			answers[INCORRECT] = append(answers[INCORRECT], ans)
		}
	}

	fmt.Printf("Total: %d (%d correct, %d incorrect, %d missed)\n",
		len(qs),
		len(answers[CORRECT]),
		len(answers[INCORRECT]),
		len(qs)-(len(answers[CORRECT])+len(answers[INCORRECT])))
}

func prompt(ctx context.Context, q []string) (string, error) {
	var ans string
	var err error

	fmt.Printf("%s = ", q[0])

	ansCh := make(chan string)
	go func() {
		s := ""
		fmt.Scanf("%s\n", &s)
		ansCh <- strings.TrimSpace(s)
	}()

	select {
	case <-ctx.Done():
		ans = ""
		err = ErrTimesUp
		fmt.Fprintf(os.Stdin, "\n")
	case ans = <-ansCh:
	}

	return ans, err
}

func questions(csvPath string) ([][]string, error) {
	f, err := os.Open(csvPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	all, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	return all, nil
}
