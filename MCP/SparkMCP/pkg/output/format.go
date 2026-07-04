// Package output provides report formatting utilities.
package output

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spark-debug-mcp/spark-debug-mcp/internal/models"
)

// Format formats a failure report in the requested format.
func Format(report *models.FailureReport, format models.OutputFormat) (string, error) {
	switch format {
	case models.FormatJSON:
		return FormatJSON(report)
	case models.FormatMarkdown:
		return FormatMarkdown(report)
	case models.FormatConsole:
		return FormatConsole(report)
	default:
		return FormatJSON(report)
	}
}

// FormatJSON returns JSON formatted report.
func FormatJSON(report *models.FailureReport) (string, error) {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FormatMarkdown returns Markdown formatted report.
func FormatMarkdown(report *models.FailureReport) (string, error) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Spark Failure Analysis: %s\n\n", report.Application))
	sb.WriteString(fmt.Sprintf("**Namespace:** %s  \n", report.Namespace))
	sb.WriteString(fmt.Sprintf("**Generated:** %s  \n\n", report.GeneratedAt.Format("2006-01-02 15:04:05 UTC")))

	sb.WriteString("## Summary\n\n")
	sb.WriteString(report.FailureSummary + "\n\n")

	sb.WriteString("## Root Cause Analysis\n\n")
	sb.WriteString(fmt.Sprintf("- **Likely Root Cause:** %s\n", report.LikelyRootCause))
	sb.WriteString(fmt.Sprintf("- **Confidence:** %.0f%%\n", report.Confidence*100))
	sb.WriteString(fmt.Sprintf("- **Failure Score:** %.0f/100\n", report.FailureScore))
	sb.WriteString(fmt.Sprintf("- **Cluster Health:** %.0f/100\n\n", report.ClusterHealth))

	if len(report.Evidence) > 0 {
		sb.WriteString("### Evidence\n\n")
		for _, e := range report.Evidence {
			sb.WriteString(fmt.Sprintf("- %s\n", e))
		}
		sb.WriteByte('\n')
	}

	if len(report.MatchedRules) > 0 {
		sb.WriteString("## Detected Issues\n\n")
		sb.WriteString("| Rule | Severity | Score |\n|------|----------|-------|\n")
		for _, m := range report.MatchedRules {
			sb.WriteString(fmt.Sprintf("| %s | %s | %.0f |\n", m.Name, m.Severity, m.Score))
		}
		sb.WriteByte('\n')
	}

	if len(report.Recommendations) > 0 {
		sb.WriteString("## Recommendations\n\n")
		for _, r := range report.Recommendations {
			sb.WriteString(fmt.Sprintf("### %d. %s\n\n", r.Priority, r.Action))
			sb.WriteString(r.Description + "\n")
			if r.DocLink != "" {
				sb.WriteString(fmt.Sprintf("\n[Documentation](%s)\n", r.DocLink))
			}
			sb.WriteByte('\n')
		}
	}

	if report.LogSummary != "" {
		sb.WriteString("## Log Summary\n\n")
		sb.WriteString("```\n" + report.LogSummary + "\n```\n\n")
	}

	if len(report.Timeline) > 0 {
		sb.WriteString("## Timeline\n\n")
		for _, t := range report.Timeline {
			sb.WriteString(fmt.Sprintf("- `%s` **%s** (%s): %s\n",
				t.Timestamp.Format("15:04:05"), t.Event, t.Source, t.Description))
		}
	}

	return sb.String(), nil
}

// FormatConsole returns human-readable console output.
func FormatConsole(report *models.FailureReport) (string, error) {
	var sb strings.Builder
	sb.WriteString("╔══════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║           SPARK FAILURE ANALYSIS REPORT                      ║\n")
	sb.WriteString("╚══════════════════════════════════════════════════════════════╝\n\n")
	sb.WriteString(fmt.Sprintf("Application: %s (namespace: %s)\n", report.Application, report.Namespace))
	sb.WriteString(fmt.Sprintf("Generated:   %s\n\n", report.GeneratedAt.Format("2006-01-02 15:04:05 UTC")))

	sb.WriteString("── Summary ──────────────────────────────────────────────────\n")
	sb.WriteString(report.FailureSummary + "\n\n")

	sb.WriteString("── Root Cause ─────────────────────────────────────────────────\n")
	sb.WriteString(fmt.Sprintf("  Cause:      %s\n", report.LikelyRootCause))
	sb.WriteString(fmt.Sprintf("  Confidence: %.0f%%\n", report.Confidence*100))
	sb.WriteString(fmt.Sprintf("  Score:      %.0f/100\n\n", report.FailureScore))

	if len(report.MatchedRules) > 0 {
		sb.WriteString("── Issues Detected ────────────────────────────────────────────\n")
		for i, m := range report.MatchedRules {
			if i >= 5 {
				sb.WriteString(fmt.Sprintf("  ... and %d more\n", len(report.MatchedRules)-5))
				break
			}
			sb.WriteString(fmt.Sprintf("  [%s] %s (score: %.0f)\n", m.Severity, m.Name, m.Score))
		}
		sb.WriteByte('\n')
	}

	if len(report.Recommendations) > 0 {
		sb.WriteString("── Recommendations ────────────────────────────────────────────\n")
		for _, r := range report.Recommendations {
			sb.WriteString(fmt.Sprintf("  %d. %s\n     %s\n", r.Priority, r.Action, r.Description))
		}
	}

	return sb.String(), nil
}

// FormatAny formats data as JSON string for MCP tool responses.
func FormatAny(v interface{}) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
