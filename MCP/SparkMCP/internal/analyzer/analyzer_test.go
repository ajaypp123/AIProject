package analyzer_test

import (
	"testing"
	"time"

	"github.com/spark-debug-mcp/spark-debug-mcp/internal/analyzer"
	"github.com/spark-debug-mcp/spark-debug-mcp/internal/models"
)

func TestLogParserParseExceptions(t *testing.T) {
	parser := analyzer.NewLogParser()
	logs := `2024-01-15 10:00:00 ERROR org.apache.spark.SparkException: Job aborted
	at org.apache.spark.scheduler.DAGScheduler.failJobAndIndependentStages(DAGScheduler.scala:1234)
java.lang.OutOfMemoryError: Java heap space`

	parsed := parser.ParseDriverLogs(logs, "driver")
	if len(parsed) == 0 {
		t.Fatal("expected parsed logs")
	}
	hasException := false
	for _, p := range parsed {
		if len(p.Exceptions) > 0 {
			hasException = true
		}
	}
	if !hasException {
		t.Error("expected exception to be parsed")
	}
}

func TestSummarizeLogs(t *testing.T) {
	logs := []models.ParsedLog{
		{Level: "ERROR", Message: "Something failed"},
		{Level: "INFO", Message: "Normal operation"},
	}
	summary := analyzer.SummarizeLogs(logs, 5)
	if summary == "" {
		t.Error("expected non-empty summary")
	}
}

func TestComputeFailureScore(t *testing.T) {
	matches := []models.RuleMatch{
		{Score: 90},
		{Score: 75},
	}
	score := analyzer.ComputeFailureScore(matches)
	if score != 100 {
		t.Errorf("expected capped score 100, got %f", score)
	}
}

func TestComputeClusterHealth(t *testing.T) {
	matches := []models.RuleMatch{
		{Severity: models.SeverityCritical},
		{Severity: models.SeverityHigh},
	}
	ctx := &models.AnalysisContext{}
	health := analyzer.ComputeClusterHealth(ctx, matches)
	if health >= 100 {
		t.Errorf("expected reduced health, got %f", health)
	}
}

func TestBuildTimeline(t *testing.T) {
	ctx := &models.AnalysisContext{
		Events: []models.EventSummary{
			{
				Reason:    "Failed",
				Message:   "Pod failed",
				Object:    "Pod/test",
				LastSeen:  time.Now(),
				Type:      "Warning",
			},
		},
		Application: models.SparkApplicationDetail{
			SubmissionTime: time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
		},
	}
	matches := []models.RuleMatch{
		{Name: "Test", Severity: models.SeverityCritical, Description: "test"},
	}
	timeline := analyzer.BuildTimeline(ctx, matches)
	if len(timeline) == 0 {
		t.Error("expected timeline entries")
	}
}

func TestComputeConfidence(t *testing.T) {
	matches := []models.RuleMatch{{Score: 85}}
	conf := analyzer.ComputeConfidence(matches)
	if conf < 0.3 || conf > 0.95 {
		t.Errorf("confidence out of range: %f", conf)
	}
}

func TestComputeConfidenceEmpty(t *testing.T) {
	conf := analyzer.ComputeConfidence(nil)
	if conf != 0.1 {
		t.Errorf("expected 0.1 for empty matches, got %f", conf)
	}
}
