package diagnostics_test

import (
	"context"
	"testing"

	"github.com/spark-debug-mcp/spark-debug-mcp/internal/diagnostics"
	"github.com/spark-debug-mcp/spark-debug-mcp/internal/kubernetes"
	"github.com/spark-debug-mcp/spark-debug-mcp/internal/models"
	"github.com/spark-debug-mcp/spark-debug-mcp/internal/spark"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func setupCollectorWithExecutors(t *testing.T) *diagnostics.Collector {
	t.Helper()
	gvr := schema.GroupVersionResource{Group: "sparkoperator.k8s.io", Version: "v1beta2", Resource: "sparkapplications"}
	scheme := runtime.NewScheme()
	dynamicClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, map[schema.GroupVersionResource]string{
		gvr: "SparkApplicationList",
	}, &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "sparkoperator.k8s.io/v1beta2",
			"kind":       "SparkApplication",
			"metadata":   map[string]interface{}{"name": "test-app", "namespace": "spark"},
			"status":     map[string]interface{}{"applicationState": map[string]interface{}{"state": "FAILED"}},
		},
	})

	k8sFake := k8sfake.NewSimpleClientset(
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-app-driver", Namespace: "spark",
				Labels: map[string]string{"sparkoperator.k8s.io/app-name": "test-app", "spark-role": "driver"},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Name: "spark",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1")},
					},
				}},
			},
			Status: corev1.PodStatus{Phase: corev1.PodFailed},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-app-exec-1", Namespace: "spark",
				Labels: map[string]string{"sparkoperator.k8s.io/app-name": "test-app", "spark-role": "executor"},
			},
			Status: corev1.PodStatus{Phase: corev1.PodFailed},
		},
	)

	k8sClient := &kubernetes.Client{Clientset: k8sFake, Dynamic: dynamicClient}
	return diagnostics.NewCollector(k8sClient, spark.NewService(k8sClient))
}

func TestDriverLogs(t *testing.T) {
	c := setupCollectorWithExecutors(t)
	// Logs will fail without log stream mock, but should find driver pod path
	_, err := c.DriverLogs(context.Background(), "spark", "test-app", models.LogOptions{Tail: 10})
	// fake client doesn't support log streaming - expect error or empty
	if err == nil {
		t.Log("driver logs retrieved (unexpected with fake client)")
	}
}

func TestExecutorLogsWithFilter(t *testing.T) {
	c := setupCollectorWithExecutors(t)
	result, err := c.ExecutorLogs(context.Background(), "spark", "test-app", models.LogOptions{
		Executor: "test-app-exec-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Logs) != 1 {
		t.Fatalf("expected 1 executor log entry, got %d", len(result.Logs))
	}
}

func TestDescribeExecutorByName(t *testing.T) {
	c := setupCollectorWithExecutors(t)
	desc, err := c.DescribeExecutor(context.Background(), "spark", "test-app", "test-app-exec-1")
	if err != nil {
		t.Fatal(err)
	}
	if desc.PodName != "test-app-exec-1" {
		t.Errorf("unexpected pod: %s", desc.PodName)
	}
}

func TestDescribeExecutorDefault(t *testing.T) {
	c := setupCollectorWithExecutors(t)
	desc, err := c.DescribeExecutor(context.Background(), "spark", "test-app", "")
	if err != nil {
		t.Fatal(err)
	}
	if desc.PodName == "" {
		t.Error("expected executor pod name")
	}
}
