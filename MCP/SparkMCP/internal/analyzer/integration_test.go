package analyzer_test

import (
	"context"
	"testing"

	"github.com/spark-debug-mcp/spark-debug-mcp/internal/analyzer"
	"github.com/spark-debug-mcp/spark-debug-mcp/internal/diagnostics"
	"github.com/spark-debug-mcp/spark-debug-mcp/internal/kubernetes"
	"github.com/spark-debug-mcp/spark-debug-mcp/internal/models"
	"github.com/spark-debug-mcp/spark-debug-mcp/internal/spark"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func TestAnalyzeFailureIntegration(t *testing.T) {
	gvr := schema.GroupVersionResource{Group: "sparkoperator.k8s.io", Version: "v1beta2", Resource: "sparkapplications"}
	scheme := runtime.NewScheme()
	dynamicClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, map[schema.GroupVersionResource]string{
		gvr: "SparkApplicationList",
	}, &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "sparkoperator.k8s.io/v1beta2",
			"kind":       "SparkApplication",
			"metadata": map[string]interface{}{
				"name": "failed-app", "namespace": "spark",
			},
			"status": map[string]interface{}{
				"applicationState": map[string]interface{}{
					"state":        "FAILED",
					"errorMessage": "driver pod failed",
				},
				"driverInfo": map[string]interface{}{"podName": "failed-app-driver"},
			},
			"spec": map[string]interface{}{"sparkVersion": "3.5.0"},
		},
	})

	k8sFake := k8sfake.NewSimpleClientset(
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "failed-app-driver", Namespace: "spark",
				Labels: map[string]string{
					"sparkoperator.k8s.io/app-name": "failed-app",
					"spark-role": "driver",
				},
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodFailed,
				ContainerStatuses: []corev1.ContainerStatus{
					{RestartCount: 3, State: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff"},
					}},
				},
			},
		},
		&corev1.Event{
			ObjectMeta: metav1.ObjectMeta{Name: "ev1", Namespace: "spark"},
			Type: "Warning", Reason: "BackOff",
			Message: "Back-off restarting failed container",
			InvolvedObject: corev1.ObjectReference{Kind: "Pod", Name: "failed-app-driver"},
		},
	)

	k8sClient := &kubernetes.Client{Clientset: k8sFake, Dynamic: dynamicClient}
	sparkSvc := spark.NewService(k8sClient)
	collector := diagnostics.NewCollector(k8sClient, sparkSvc)
	anal := analyzer.NewAnalyzer(collector)

	report, err := anal.AnalyzeFailure(context.Background(), "spark", "failed-app")
	if err != nil {
		t.Fatal(err)
	}
	if report.Application != "failed-app" {
		t.Errorf("unexpected app: %s", report.Application)
	}
	if report.FailureSummary == "" {
		t.Error("expected failure summary")
	}
}

func TestAnalyzeFailureWithOOMLogs(t *testing.T) {
	collector := &mockCollector{
		ctx: &models.AnalysisContext{
			Application: models.SparkApplicationDetail{
				Name: "oom-app", Namespace: "spark", State: "FAILED",
			},
			DriverLogs: "java.lang.OutOfMemoryError: Java heap space\nExecutorLost: executor 1",
			Events: []models.EventSummary{
				{Reason: "OOMKilled", Message: "Container killed", Object: "Pod/exec-1", Type: "Warning"},
			},
		},
	}
	anal := analyzer.NewAnalyzer(collector)
	report, err := anal.AnalyzeFailure(context.Background(), "spark", "oom-app")
	if err != nil {
		t.Fatal(err)
	}
	if len(report.MatchedRules) == 0 {
		t.Error("expected matched rules for OOM scenario")
	}
	if report.Confidence < 0.3 {
		t.Errorf("expected reasonable confidence, got %f", report.Confidence)
	}
}

type mockCollector struct {
	ctx *models.AnalysisContext
}

func (m *mockCollector) ListApps(ctx context.Context, namespace string) ([]models.SparkAppSummary, error) {
	return nil, nil
}
func (m *mockCollector) GetApplication(ctx context.Context, namespace, name string) (*models.SparkApplicationDetail, error) {
	return &m.ctx.Application, nil
}
func (m *mockCollector) DriverLogs(ctx context.Context, namespace, appName string, opts models.LogOptions) (*models.PodLogs, error) {
	return &models.PodLogs{Logs: m.ctx.DriverLogs}, nil
}
func (m *mockCollector) ExecutorLogs(ctx context.Context, namespace, appName string, opts models.LogOptions) (*models.ExecutorLogs, error) {
	return &models.ExecutorLogs{}, nil
}
func (m *mockCollector) DescribeDriver(ctx context.Context, namespace, appName string) (*models.PodDescription, error) {
	return &models.PodDescription{Description: m.ctx.DriverDesc}, nil
}
func (m *mockCollector) DescribeExecutor(ctx context.Context, namespace, appName, executorName string) (*models.PodDescription, error) {
	return nil, nil
}
func (m *mockCollector) GetEvents(ctx context.Context, namespace string) ([]models.EventSummary, error) {
	return m.ctx.Events, nil
}
func (m *mockCollector) SparkConf(ctx context.Context, namespace, appName string) (*models.SparkConf, error) {
	return &m.ctx.SparkConf, nil
}
func (m *mockCollector) PodResources(ctx context.Context, namespace, appName string) ([]models.PodResources, error) {
	return nil, nil
}
func (m *mockCollector) CollectAll(ctx context.Context, namespace, appName string) (*models.AnalysisContext, error) {
	return m.ctx, nil
}
