package conformance

import (
	"fmt"
	"strings"
	"time"
)

// CertificationReport represents the result of a plugin certification process.
type CertificationReport struct {
	PluginName    string    `json:"plugin_name"`
	PluginVersion string    `json:"plugin_version"`
	Certified     bool      `json:"certified"`
	CertifiedAt   time.Time `json:"certified_at"`
	Issues        []string  `json:"issues,omitempty"`
	SuiteSummary  Summary   `json:"suite_summary"`
}

// Certify evaluates a SuiteReport and generates a CertificationReport.
func Certify(suiteReport *SuiteReport, name, version string) *CertificationReport {
	report := &CertificationReport{
		PluginName:    name,
		PluginVersion: version,
		CertifiedAt:   time.Now(),
		SuiteSummary:  suiteReport.Summary,
	}

	if suiteReport.Summary.Failed > 0 {
		report.Certified = false
		for _, res := range suiteReport.Results {
			if res.Status == StatusFail || res.Status == StatusError {
				report.Issues = append(report.Issues, fmt.Sprintf("%s: %s", res.TestName, res.Error))
			}
		}
	} else {
		report.Certified = true
	}

	return report
}

// GenerateMarkdown generates a markdown report for certification.
func (r *CertificationReport) GenerateMarkdown() string {
	status := "❌ FAILED"
	if r.Certified {
		status = "✅ CERTIFIED"
	}

	var sb strings.Builder
	sb.WriteString("# Pulumicost Plugin Certification\n\n")
	sb.WriteString(fmt.Sprintf("**Plugin**: %s\n", r.PluginName))
	sb.WriteString(fmt.Sprintf("**Version**: %s\n", r.PluginVersion))
	sb.WriteString(fmt.Sprintf("**Status**: %s\n", status))
	sb.WriteString(fmt.Sprintf("**Date**: %s\n\n", r.CertifiedAt.Format(time.RFC1123)))

	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("- Total Tests: %d\n", r.SuiteSummary.Total))
	sb.WriteString(fmt.Sprintf("- Passed: %d\n", r.SuiteSummary.Passed))
	sb.WriteString(fmt.Sprintf("- Failed: %d\n", r.SuiteSummary.Failed))
	sb.WriteString(fmt.Sprintf("- Skipped: %d\n\n", r.SuiteSummary.Skipped))

	if len(r.Issues) > 0 {
		sb.WriteString("## Issues\n\n")
		for _, issue := range r.Issues {
			sb.WriteString("- " + issue + "\n")
		}
	}

	return sb.String()
}
