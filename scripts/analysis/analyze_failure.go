package main

import (
	"errors"
	"fmt"
	"regexp"
)

const expectedMatches = 2

// runIDRegex matches GitHub Actions run URLs to extract the run ID.
var runIDRegex = regexp.MustCompile(`actions/runs/(\d+)`)

// FailureContext represents the context of a workflow failure.
type FailureContext struct {
	RunID     string      `json:"run_id"`
	Jobs      []FailedJob `json:"jobs"`
	IssueBody string      `json:"issue_body"`
}

// FailedJob represents a failed job within a workflow run.
type FailedJob struct {
	Name string `json:"name"`
	Logs string `json:"logs"`
}

// extractRunID extracts the numeric workflow run ID from the provided issue body.
// It returns the captured run ID as a string when a substring matching "actions/runs/<digits>"
// is present in body. If no such match is found, it returns an error indicating that no run ID was found.
func extractRunID(body string) (string, error) {
	matches := runIDRegex.FindStringSubmatch(body)
	if len(matches) < expectedMatches {
		return "", errors.New("no run ID found in issue body: expected pattern 'actions/runs/<number>'")
	}
	return matches[1], nil
}

// main is the program entry point for a lightweight analysis scaffold.
// It prints a short placeholder message ("Analysis script placeholder") and performs no further actions.
func main() {
	//nolint:forbidigo // Script output
	fmt.Println("Analysis script placeholder")
}
