package main

import (
	"errors"
	"fmt"
	"regexp"
)

const expectedMatches = 2

// FailureContext represents the context of a workflow failure.
type FailureContext struct {
	RunID     string
	Jobs      []FailedJob
	IssueBody string
}

// FailedJob represents a failed job within a workflow run.
type FailedJob struct {
	Name string
	Logs string
}

func extractRunID(body string) (string, error) {
	re := regexp.MustCompile(`actions/runs/(\d+)`)
	matches := re.FindStringSubmatch(body)
	if len(matches) < expectedMatches {
		return "", errors.New("no run ID found in issue body")
	}
	return matches[1], nil
}

func main() {
	//nolint:forbidigo // Script output
	fmt.Println("Analysis script placeholder")
}
