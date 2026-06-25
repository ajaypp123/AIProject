package analyzer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spark-debug-mcp/spark-debug-mcp/internal/analyzer/rules"
	"github.com/spark-debug-mcp/spark-debug-mcp/internal/diagnostics"
	"github.com/spark-debug-mcp/spark-debug-mcp/internal/models"
)

// Analyzer performs comprehensive failure analysis.
type Analyzer struct {
	collector diagnostics.CollectorInterface
	engine    *rules.Engine
	parser    *LogParser
}

// NewAnalyzer creates a failure analyzer.
func NewAnalyzer(collector diagnostics.CollectorInterface) *Analyzer {
	return &Analyzer{
		collector: collector,
		engine:    rules.NewEngine(),
		parser:    NewLogParser(),
	}
}

// AnalyzeFailure runs full diagnostic analysis and produces a report.
func (a *Analyzer) AnalyzeFailure(ctx context.Context, namespace, appName string) (*models.FailureReport, error) {
	actx, err := a.collector.CollectAll(ctx, namespace, appName)
	if err != nil {
		return nil, fmt.Errorf("collect diagnostics: %w", err)
	}

	actx.ParsedLogs = a.parser.ParseDriverLogs(actx.DriverLogs, "driver")
	for _, el := range actx.ExecutorLogs {
		actx.ParsedLogs = append(actx.ParsedLogs, a.parser.ParseExecutorLogs(el.Logs, el.PodName)...)
	}
	actx.ParsedLogs = append(actx.ParsedLogs, a.parser.ParseEvents(actx.Events)...)

	matches := a.engine.Evaluate(actx)

	report := &models.FailureReport{
		Application:       appName,
		Namespace:         namespace,
		KubernetesEvents:  actx.Events,
		RelevantLogs:      filterRelevantLogs(actx.ParsedLogs),
		SparkConfig:       actx.SparkConf,
		MatchedRules:      matches,
		ApplicationDetail: actx.Application,
		GeneratedAt:       time.Now(),
		Timeline:          BuildTimeline(actx, matches),
		FailureScore:      ComputeFailureScore(matches),
		ClusterHealth:     ComputeClusterHealth(actx, matches),
		LogSummary:        SummarizeLogs(actx.ParsedLogs, 15),
	}

	if len(matches) > 0 {
		top := matches[0]
		report.LikelyRootCause = top.Name + ": " + top.Description
		report.Confidence = ComputeConfidence(matches)
		report.Evidence = top.Evidence
		report.Recommendations = buildRecommendations(matches)
		report.FailureSummary = buildFailureSummary(actx, matches)
	} else {
		state := actx.Application.State
		report.LikelyRootCause = "No specific failure rule matched"
		report.Confidence = 0.2
		report.FailureSummary = fmt.Sprintf("Application %s in namespace %s has state '%s'. No definitive failure pattern detected.", appName, namespace, state)
		report.Recommendations = []models.Recommendation{
			{Priority: 1, Action: "Review logs manually", Description: "Examine driver and executor logs for errors not covered by automated rules."},
		}
	}

	return report, nil
}

func filterRelevantLogs(logs []models.ParsedLog) []models.ParsedLog {
	relevant := make([]models.ParsedLog, 0)
	for _, log := range logs {
		if log.Level == "ERROR" || log.Level == "Warning" || len(log.Exceptions) > 0 {
			relevant = append(relevant, log)
		}
	}
	if len(relevant) > 50 {
		return relevant[:50]
	}
	return relevant
}

func buildRecommendations(matches []models.RuleMatch) []models.Recommendation {
	recs := make([]models.Recommendation, 0, len(matches))
	seen := make(map[string]bool)
	priority := 1
	for _, m := range matches {
		if seen[m.Recommendation] {
			continue
		}
		seen[m.Recommendation] = true
		docLink := ""
		if len(m.DocLinks) > 0 {
			docLink = m.DocLinks[0]
		}
		recs = append(recs, models.Recommendation{
			Priority:    priority,
			Action:      m.Name,
			Description: m.Recommendation,
			DocLink:     docLink,
		})
		priority++
		if priority > 10 {
			break
		}
	}
	return recs
}

func buildFailureSummary(ctx *models.AnalysisContext, matches []models.RuleMatch) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Spark application '%s' in namespace '%s' ", ctx.Application.Name, ctx.Application.Namespace))
	sb.WriteString(fmt.Sprintf("is in state '%s'. ", ctx.Application.State))
	if len(matches) > 0 {
		sb.WriteString(fmt.Sprintf("Detected %d issue(s). Top issue: %s (%s). ",
			len(matches), matches[0].Name, matches[0].Severity))
	}
	if ctx.Application.RestartCount > 0 {
		sb.WriteString(fmt.Sprintf("Total container restarts: %d. ", ctx.Application.RestartCount))
	}
	return sb.String()
}
