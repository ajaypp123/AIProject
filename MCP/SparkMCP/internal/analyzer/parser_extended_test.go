package analyzer_test

import (
	"testing"

	"github.com/spark-debug-mcp/spark-debug-mcp/internal/analyzer"
	"github.com/spark-debug-mcp/spark-debug-mcp/internal/models"
)

func TestParseExecutorLogs(t *testing.T) {
	parser := analyzer.NewLogParser()
	parsed := parser.ParseExecutorLogs("ERROR executor failed", "exec-1")
	if len(parsed) == 0 {
		t.Error("expected parsed executor logs")
	}
}

func TestParseEvents(t *testing.T) {
	parser := analyzer.NewLogParser()
	events := []models.EventSummary{
		{Type: "Warning", Reason: "Failed", Message: "Pod failed", Object: "Pod/p1"},
	}
	parsed := parser.ParseEvents(events)
	if len(parsed) != 1 {
		t.Fatalf("expected 1 parsed event, got %d", len(parsed))
	}
	if parsed[0].Source != "event:Pod/p1" {
		t.Errorf("unexpected source: %s", parsed[0].Source)
	}
}

func TestSummarizeLogsNoErrors(t *testing.T) {
	logs := []models.ParsedLog{{Level: "INFO", Message: "all good"}}
	summary := analyzer.SummarizeLogs(logs, 5)
	if summary == "" {
		t.Error("expected summary")
	}
}

func TestBuildTimelineWithSubmission(t *testing.T) {
	ctx := &models.AnalysisContext{
		Application: models.SparkApplicationDetail{
			SubmissionTime: "2024-01-15T10:00:00Z",
		},
	}
	timeline := analyzer.BuildTimeline(ctx, nil)
	if len(timeline) == 0 {
		t.Error("expected timeline with submission")
	}
}

func TestComputeFailureScoreEmpty(t *testing.T) {
	if analyzer.ComputeFailureScore(nil) != 0 {
		t.Error("expected 0 for empty matches")
	}
}

func TestComputeClusterHealthNoIssues(t *testing.T) {
	health := analyzer.ComputeClusterHealth(&models.AnalysisContext{}, nil)
	if health != 100 {
		t.Errorf("expected 100 health, got %f", health)
	}
}
