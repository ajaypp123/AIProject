package output_test

import (
	"strings"
	"testing"
	"time"

	"github.com/spark-debug-mcp/spark-debug-mcp/internal/models"
	"github.com/spark-debug-mcp/spark-debug-mcp/pkg/output"
)

func sampleReport() *models.FailureReport {
	return &models.FailureReport{
		Application:     "test-app",
		Namespace:       "spark",
		FailureSummary:  "Application failed due to OOM",
		LikelyRootCause: "Executor OOM: Executor killed due to out-of-memory",
		Confidence:      0.85,
		FailureScore:    90,
		ClusterHealth:   65,
		Evidence:        []string{"OOMKilled event on exec-1"},
		MatchedRules: []models.RuleMatch{
			{ID: "R001", Name: "Executor OOM", Severity: models.SeverityCritical, Score: 90},
		},
		Recommendations: []models.Recommendation{
			{Priority: 1, Action: "Increase memory", Description: "Increase spark.executor.memory"},
		},
		GeneratedAt: time.Now(),
	}
}

func TestFormatJSON(t *testing.T) {
	text, err := output.FormatJSON(sampleReport())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "test-app") {
		t.Error("JSON should contain application name")
	}
	if !strings.Contains(text, "Executor OOM") {
		t.Error("JSON should contain matched rule")
	}
}

func TestFormatMarkdown(t *testing.T) {
	text, err := output.FormatMarkdown(sampleReport())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "# Spark Failure Analysis") {
		t.Error("markdown should have header")
	}
	if !strings.Contains(text, "Root Cause") {
		t.Error("markdown should have root cause section")
	}
}

func TestFormatConsole(t *testing.T) {
	text, err := output.FormatConsole(sampleReport())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "SPARK FAILURE ANALYSIS") {
		t.Error("console should have banner")
	}
}

func TestFormatWithFormatEnum(t *testing.T) {
	report := sampleReport()
	for _, fmt := range []models.OutputFormat{models.FormatJSON, models.FormatMarkdown, models.FormatConsole} {
		text, err := output.Format(report, fmt)
		if err != nil {
			t.Errorf("format %s failed: %v", fmt, err)
		}
		if text == "" {
			t.Errorf("format %s returned empty", fmt)
		}
	}
}

func TestFormatAny(t *testing.T) {
	text, err := output.FormatAny(map[string]string{"key": "value"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "value") {
		t.Error("expected formatted value")
	}
}
