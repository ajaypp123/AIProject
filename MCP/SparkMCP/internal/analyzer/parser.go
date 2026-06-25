// Package analyzer provides failure analysis and rule engine.
package analyzer

import (
	"regexp"
	"strings"
	"time"

	"github.com/spark-debug-mcp/spark-debug-mcp/internal/models"
)

// LogParser parses Spark and Kubernetes logs.
type LogParser struct {
	exceptionRe  *regexp.Regexp
	stackTraceRe *regexp.Regexp
	sparkErrorRe *regexp.Regexp
	timestampRe  *regexp.Regexp
}

// NewLogParser creates a log parser.
func NewLogParser() *LogParser {
	return &LogParser{
		exceptionRe:  regexp.MustCompile(`(?i)([\w$.]+Exception|[\w$.]+Error):\s*(.+)`),
		stackTraceRe: regexp.MustCompile(`(?m)^\s+at\s+[\w$.]+\([\w$.]+:\d+\)`),
		sparkErrorRe: regexp.MustCompile(`(?i)(ERROR|WARN|FATAL)\s+`),
		timestampRe:  regexp.MustCompile(`\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}`),
	}
}

// ParseDriverLogs parses driver log content.
func (p *LogParser) ParseDriverLogs(content, source string) []models.ParsedLog {
	return p.parseLogs(content, source)
}

// ParseExecutorLogs parses executor log content.
func (p *LogParser) ParseExecutorLogs(content, source string) []models.ParsedLog {
	return p.parseLogs(content, source)
}

// ParseEvents converts events to parsed log entries.
func (p *LogParser) ParseEvents(events []models.EventSummary) []models.ParsedLog {
	parsed := make([]models.ParsedLog, 0, len(events))
	for _, ev := range events {
		parsed = append(parsed, models.ParsedLog{
			Level:     ev.Type,
			Message:   ev.Reason + ": " + ev.Message,
			Timestamp: ev.LastSeen.Format(time.RFC3339),
			Source:    "event:" + ev.Object,
		})
	}
	return parsed
}

func (p *LogParser) parseLogs(content, source string) []models.ParsedLog {
	if content == "" {
		return nil
	}
	lines := strings.Split(content, "\n")
	parsed := make([]models.ParsedLog, 0)
	var current *models.ParsedLog
	var stackLines []string

	for _, line := range lines {
		if line == "" {
			continue
		}
		if m := p.exceptionRe.FindStringSubmatch(line); len(m) >= 3 {
			if current != nil && len(stackLines) > 0 {
				current.StackTrace = strings.Join(stackLines, "\n")
			}
			current = &models.ParsedLog{
				Message:    m[1] + ": " + m[2],
				Exceptions: []string{m[1]},
				Source:     source,
			}
			if ts := p.timestampRe.FindString(line); ts != "" {
				current.Timestamp = ts
			}
			if p.sparkErrorRe.MatchString(line) {
				current.Level = "ERROR"
			}
			parsed = append(parsed, *current)
			stackLines = nil
			continue
		}
		if p.stackTraceRe.MatchString(line) {
			stackLines = append(stackLines, line)
			continue
		}
		if p.sparkErrorRe.MatchString(line) {
			entry := models.ParsedLog{
				Level:   "ERROR",
				Message: line,
				Source:  source,
			}
			if ts := p.timestampRe.FindString(line); ts != "" {
				entry.Timestamp = ts
			}
			parsed = append(parsed, entry)
		}
	}
	if current != nil && len(stackLines) > 0 && len(parsed) > 0 {
		parsed[len(parsed)-1].StackTrace = strings.Join(stackLines, "\n")
	}
	return parsed
}

// SummarizeLogs produces a brief summary of log content.
func SummarizeLogs(logs []models.ParsedLog, maxEntries int) string {
	if len(logs) == 0 {
		return "No significant log entries found."
	}
	if maxEntries <= 0 {
		maxEntries = 10
	}
	var sb strings.Builder
	sb.WriteString("Key log entries:\n")
	count := 0
	for _, log := range logs {
		if log.Level == "ERROR" || len(log.Exceptions) > 0 {
			sb.WriteString("- ")
			sb.WriteString(log.Message)
			sb.WriteByte('\n')
			count++
			if count >= maxEntries {
				break
			}
		}
	}
	if count == 0 {
		sb.WriteString("- No errors detected in parsed logs\n")
	}
	return sb.String()
}

// BuildTimeline constructs a failure timeline from analysis context.
func BuildTimeline(ctx *models.AnalysisContext, matches []models.RuleMatch) []models.TimelineEntry {
	entries := make([]models.TimelineEntry, 0)

	for _, ev := range ctx.Events {
		entries = append(entries, models.TimelineEntry{
			Timestamp:   ev.LastSeen,
			Source:      "kubernetes",
			Event:       ev.Reason,
			Description: ev.Message,
			Severity:    ev.Type,
		})
	}

	if ctx.Application.SubmissionTime != "" {
		if t, err := time.Parse(time.RFC3339, ctx.Application.SubmissionTime); err == nil {
			entries = append(entries, models.TimelineEntry{
				Timestamp:   t,
				Source:      "spark",
				Event:       "Submission",
				Description: "Application submitted",
			})
		}
	}

	for _, log := range ctx.ParsedLogs {
		if log.Level == "ERROR" && log.Timestamp != "" {
			if t, err := time.Parse(time.RFC3339, log.Timestamp); err == nil {
				entries = append(entries, models.TimelineEntry{
					Timestamp:   t,
					Source:      log.Source,
					Event:       "Error",
					Description: truncate(log.Message, 200),
					Severity:    "ERROR",
				})
			}
		}
	}

	for _, m := range matches {
		if m.Severity == models.SeverityCritical || m.Severity == models.SeverityHigh {
			entries = append(entries, models.TimelineEntry{
				Timestamp:   time.Now(),
				Source:      "analyzer",
				Event:       m.Name,
				Description: m.Description,
				Severity:    string(m.Severity),
			})
		}
	}

	// Sort by timestamp
	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].Timestamp.Before(entries[i].Timestamp) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
	return entries
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// ComputeFailureScore calculates a failure severity score 0-100.
func ComputeFailureScore(matches []models.RuleMatch) float64 {
	if len(matches) == 0 {
		return 0
	}
	score := 0.0
	for _, m := range matches {
		score += m.Score
	}
	if score > 100 {
		return 100
	}
	return score
}

// ComputeClusterHealth calculates cluster health score 0-100.
func ComputeClusterHealth(ctx *models.AnalysisContext, matches []models.RuleMatch) float64 {
	health := 100.0
	for _, m := range matches {
		switch m.Severity {
		case models.SeverityCritical:
			health -= 25
		case models.SeverityHigh:
			health -= 15
		case models.SeverityMedium:
			health -= 8
		case models.SeverityLow:
			health -= 3
		}
	}
	pressureEvents := 0
	for _, ev := range ctx.Events {
		if strings.Contains(ev.Reason, "Pressure") || strings.Contains(ev.Reason, "FailedScheduling") {
			pressureEvents++
		}
	}
	health -= float64(pressureEvents) * 5
	if health < 0 {
		return 0
	}
	return health
}

// ComputeConfidence calculates root cause confidence.
func ComputeConfidence(matches []models.RuleMatch) float64 {
	if len(matches) == 0 {
		return 0.1
	}
	top := matches[0]
	conf := top.Score / 100.0
	if conf > 0.95 {
		return 0.95
	}
	if conf < 0.3 {
		return 0.3
	}
	return conf
}
