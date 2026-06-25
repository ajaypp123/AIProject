package mcpserver_test

import (
	"context"
	"errors"
	"testing"

	"github.com/spark-debug-mcp/spark-debug-mcp/internal/config"
	mcpserver "github.com/spark-debug-mcp/spark-debug-mcp/internal/mcpserver"
	"github.com/spark-debug-mcp/spark-debug-mcp/internal/models"
	"go.uber.org/zap"
)

type mockCollector struct{}
type mockAnalyzer struct{}

func (m *mockCollector) ListApps(_ context.Context, _ string) ([]models.SparkAppSummary, error) {
	return []models.SparkAppSummary{{Name: "app1", Namespace: "spark", State: "RUNNING"}}, nil
}
func (m *mockCollector) GetApplication(_ context.Context, _, _ string) (*models.SparkApplicationDetail, error) {
	return &models.SparkApplicationDetail{Name: "app1", State: "FAILED"}, nil
}
func (m *mockCollector) DriverLogs(_ context.Context, _, _ string, _ models.LogOptions) (*models.PodLogs, error) {
	return &models.PodLogs{PodName: "driver", Logs: "log line"}, nil
}
func (m *mockCollector) ExecutorLogs(_ context.Context, _, _ string, _ models.LogOptions) (*models.ExecutorLogs, error) {
	return &models.ExecutorLogs{Logs: []models.PodLogs{{PodName: "exec-1", Logs: "exec log"}}}, nil
}
func (m *mockCollector) DescribeDriver(_ context.Context, _, _ string) (*models.PodDescription, error) {
	return &models.PodDescription{PodName: "driver", Description: "Name: driver"}, nil
}
func (m *mockCollector) DescribeExecutor(_ context.Context, _, _, _ string) (*models.PodDescription, error) {
	return &models.PodDescription{PodName: "exec-1", Description: "Name: exec-1"}, nil
}
func (m *mockCollector) GetEvents(_ context.Context, _ string) ([]models.EventSummary, error) {
	return []models.EventSummary{{Reason: "Failed", Type: "Warning"}}, nil
}
func (m *mockCollector) SparkConf(_ context.Context, _, _ string) (*models.SparkConf, error) {
	return &models.SparkConf{AllConf: map[string]string{"spark.executor.memory": "2g"}}, nil
}
func (m *mockCollector) PodResources(_ context.Context, _, _ string) ([]models.PodResources, error) {
	return []models.PodResources{{PodName: "driver", QoS: "Burstable"}}, nil
}
func (m *mockCollector) CollectAll(_ context.Context, _, _ string) (*models.AnalysisContext, error) {
	return &models.AnalysisContext{}, nil
}

func (m *mockAnalyzer) AnalyzeFailure(_ context.Context, _, _ string) (*models.FailureReport, error) {
	return &models.FailureReport{
		Application:     "app1",
		FailureSummary:  "Failed due to OOM",
		LikelyRootCause: "Executor OOM",
		Confidence:      0.85,
		MatchedRules:    []models.RuleMatch{{Name: "Executor OOM", Score: 90}},
	}, nil
}

type errCollector struct{ mockCollector }

func (e *errCollector) ListApps(_ context.Context, _ string) ([]models.SparkAppSummary, error) {
	return nil, errors.New("list failed")
}

func newTestServer() *mcpserver.Server {
	cfg := &config.Config{Namespace: "spark", OutputFormat: "json", ServerName: "test", ServerVersion: "1.0.0"}
	return mcpserver.NewServer(cfg, &mockCollector{}, &mockAnalyzer{}, zap.NewNop())
}

func TestListSparkApps(t *testing.T) {
	s := newTestServer()
	_, out, err := s.HandleListSparkApps(context.Background(), nil, mcpserver.NamespaceInput{Namespace: "spark"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Result == "" || !contains(out.Result, "app1") {
		t.Errorf("unexpected output: %s", out.Result)
	}
}

func TestGetApplication(t *testing.T) {
	s := newTestServer()
	_, out, err := s.HandleGetApplication(context.Background(), nil, mcpserver.AppInput{Namespace: "spark", Application: "app1"})
	if err != nil {
		t.Fatal(err)
	}
	if !contains(out.Result, "app1") {
		t.Errorf("unexpected output: %s", out.Result)
	}
}

func TestDriverLogs(t *testing.T) {
	s := newTestServer()
	_, out, err := s.HandleDriverLogs(context.Background(), nil, mcpserver.DriverLogsInput{
		Namespace: "spark", Application: "app1", Tail: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !contains(out.Result, "driver") {
		t.Errorf("unexpected output: %s", out.Result)
	}
}

func TestExecutorLogs(t *testing.T) {
	s := newTestServer()
	_, out, err := s.HandleExecutorLogs(context.Background(), nil, mcpserver.ExecutorLogsInput{
		Namespace: "spark", Application: "app1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !contains(out.Result, "exec-1") {
		t.Errorf("unexpected output: %s", out.Result)
	}
}

func TestDescribeDriver(t *testing.T) {
	s := newTestServer()
	_, out, err := s.HandleDescribeDriver(context.Background(), nil, mcpserver.AppInput{Namespace: "spark", Application: "app1"})
	if err != nil {
		t.Fatal(err)
	}
	if !contains(out.Result, "driver") {
		t.Errorf("unexpected output: %s", out.Result)
	}
}

func TestDescribeExecutor(t *testing.T) {
	s := newTestServer()
	_, out, err := s.HandleDescribeExecutor(context.Background(), nil, mcpserver.DescribeExecutorInput{
		Namespace: "spark", Application: "app1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !contains(out.Result, "exec-1") {
		t.Errorf("unexpected output: %s", out.Result)
	}
}

func TestGetEvents(t *testing.T) {
	s := newTestServer()
	_, out, err := s.HandleGetEvents(context.Background(), nil, mcpserver.NamespaceInput{Namespace: "spark"})
	if err != nil {
		t.Fatal(err)
	}
	if !contains(out.Result, "Failed") {
		t.Errorf("unexpected output: %s", out.Result)
	}
}

func TestSparkConf(t *testing.T) {
	s := newTestServer()
	_, out, err := s.HandleSparkConf(context.Background(), nil, mcpserver.AppInput{Namespace: "spark", Application: "app1"})
	if err != nil {
		t.Fatal(err)
	}
	if !contains(out.Result, "spark.executor.memory") {
		t.Errorf("unexpected output: %s", out.Result)
	}
}

func TestPodResources(t *testing.T) {
	s := newTestServer()
	_, out, err := s.HandlePodResources(context.Background(), nil, mcpserver.AppInput{Namespace: "spark", Application: "app1"})
	if err != nil {
		t.Fatal(err)
	}
	if !contains(out.Result, "Burstable") {
		t.Errorf("unexpected output: %s", out.Result)
	}
}

func TestAnalyzeFailure(t *testing.T) {
	s := newTestServer()
	_, out, err := s.HandleAnalyzeFailure(context.Background(), nil, mcpserver.AnalyzeInput{
		Namespace: "spark", Application: "app1", Format: "json",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !contains(out.Result, "Executor OOM") {
		t.Errorf("unexpected output: %s", out.Result)
	}
}

func TestAnalyzeFailureMarkdown(t *testing.T) {
	s := newTestServer()
	_, out, err := s.HandleAnalyzeFailure(context.Background(), nil, mcpserver.AnalyzeInput{
		Namespace: "spark", Application: "app1", Format: "markdown",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !contains(out.Result, "# Spark Failure Analysis") {
		t.Errorf("unexpected markdown output")
	}
}

func TestErrorHandling(t *testing.T) {
	cfg := &config.Config{Namespace: "spark", ServerName: "test", ServerVersion: "1.0.0"}
	s := mcpserver.NewServer(cfg, &errCollector{}, &mockAnalyzer{}, zap.NewNop())
	result, _, _ := s.HandleListSparkApps(context.Background(), nil, mcpserver.NamespaceInput{Namespace: "spark"})
	if result == nil || !result.IsError {
		t.Error("expected error result")
	}
}

func TestErrorText(t *testing.T) {
	text := mcpserver.ErrorText(errors.New("test error"))
	if !contains(text, "test error") {
		t.Errorf("unexpected: %s", text)
	}
}

func TestTextResult(t *testing.T) {
	_, out, err := mcpserver.TextResult(`{"status":"ok"}`)
	if err != nil {
		t.Fatal(err)
	}
	if out.Result != `{"status":"ok"}` {
		t.Errorf("unexpected text: %s", out.Result)
	}
}

func TestDefaultNamespace(t *testing.T) {
	s := newTestServer()
	_, out, err := s.HandleListSparkApps(context.Background(), nil, mcpserver.NamespaceInput{})
	if err != nil {
		t.Fatal(err)
	}
	if out.Result == "" {
		t.Error("expected output with default namespace")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
