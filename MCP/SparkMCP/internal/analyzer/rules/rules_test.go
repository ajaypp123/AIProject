package rules_test

import (
	"testing"

	"github.com/spark-debug-mcp/spark-debug-mcp/internal/analyzer/rules"
	"github.com/spark-debug-mcp/spark-debug-mcp/internal/models"
)

func TestAllRulesCount(t *testing.T) {
	all := rules.AllRules()
	if len(all) < 40 {
		t.Fatalf("expected at least 40 rules, got %d", len(all))
	}
}

func TestExecutorOOMDetection(t *testing.T) {
	engine := rules.NewEngine()
	ctx := &models.AnalysisContext{
		DriverLogs: "ExecutorLost: executor 1 killed",
		Events: []models.EventSummary{
			{Reason: "OOMKilled", Message: "Container killed", Object: "Pod/exec-1"},
		},
	}
	matches := engine.Evaluate(ctx)
	found := false
	for _, m := range matches {
		if m.ID == "R001" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected Executor OOM rule to match")
	}
}

func TestDriverOOMDetection(t *testing.T) {
	engine := rules.NewEngine()
	ctx := &models.AnalysisContext{
		DriverLogs:  "java.lang.OutOfMemoryError: Java heap space",
		DriverDesc:  "Last Termination Reason: OOMKilled",
	}
	matches := engine.Evaluate(ctx)
	found := false
	for _, m := range matches {
		if m.ID == "R002" || m.ID == "R013" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected Driver OOM or Java Heap Space rule to match")
	}
}

func TestCrashLoopDetection(t *testing.T) {
	engine := rules.NewEngine()
	ctx := &models.AnalysisContext{
		DriverRes: models.PodResources{
			PodName:      "driver",
			RestartCount: 5,
			Containers: []models.ContainerRes{
				{State: "Waiting (CrashLoopBackOff): back-off restarting"},
			},
		},
	}
	matches := engine.Evaluate(ctx)
	found := false
	for _, m := range matches {
		if m.ID == "R012" || m.ID == "R029" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected CrashLoop rule to match")
	}
}

func TestNodePressureDetection(t *testing.T) {
	engine := rules.NewEngine()
	ctx := &models.AnalysisContext{
		Events: []models.EventSummary{
			{Reason: "NodePressure", Message: "Node has MemoryPressure", Object: "Node/worker-1"},
		},
	}
	matches := engine.Evaluate(ctx)
	found := false
	for _, m := range matches {
		if m.ID == "R003" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected Node Pressure rule to match")
	}
}

func TestImagePullBackOff(t *testing.T) {
	engine := rules.NewEngine()
	ctx := &models.AnalysisContext{
		DriverDesc: "State: Waiting (ImagePullBackOff): Back-off pulling image",
		DriverRes: models.PodResources{
			Containers: []models.ContainerRes{
				{State: "Waiting (ImagePullBackOff): Back-off pulling image"},
			},
		},
	}
	matches := engine.Evaluate(ctx)
	found := false
	for _, m := range matches {
		if m.ID == "R005" || m.ID == "R028" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected ImagePullBackOff rule to match")
	}
}

func TestFailedScheduling(t *testing.T) {
	engine := rules.NewEngine()
	ctx := &models.AnalysisContext{
		Events: []models.EventSummary{
			{Reason: "FailedScheduling", Message: "0/3 nodes available: Insufficient cpu", Object: "Pod/driver"},
		},
		DriverDesc: "FailedScheduling: 0/3 nodes available: Insufficient cpu",
	}
	matches := engine.Evaluate(ctx)
	if len(matches) == 0 {
		t.Error("expected scheduling failure rules to match")
	}
}

func TestCustomRuleRegistration(t *testing.T) {
	engine := rules.NewEngine()
	engine.Register(rules.Rule{
		ID: "CUSTOM", Name: "Custom Test", Description: "Test rule",
		Severity: models.SeverityInfo, Score: 10,
		Recommendation: "No action needed",
		Detect: func(ctx *models.AnalysisContext) []string {
			if ctx.DriverLogs == "custom-trigger" {
				return []string{"custom evidence"}
			}
			return nil
		},
	})
	matches := engine.Evaluate(&models.AnalysisContext{DriverLogs: "custom-trigger"})
	if len(matches) != 1 || matches[0].ID != "CUSTOM" {
		t.Error("custom rule should match")
	}
}

func TestNoMatchReturnsEmpty(t *testing.T) {
	engine := rules.NewEngine()
	matches := engine.Evaluate(&models.AnalysisContext{})
	if len(matches) != 0 {
		t.Errorf("expected no matches, got %d", len(matches))
	}
}
