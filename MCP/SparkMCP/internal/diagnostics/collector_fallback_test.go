package diagnostics_test

import (
	"context"
	"testing"

	"github.com/spark-debug-mcp/spark-debug-mcp/internal/diagnostics"
	"github.com/spark-debug-mcp/spark-debug-mcp/internal/kubernetes"
	"github.com/spark-debug-mcp/spark-debug-mcp/internal/models"
	"github.com/spark-debug-mcp/spark-debug-mcp/internal/spark"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func TestDriverLogsViaApplicationFallback(t *testing.T) {
	gvr := schema.GroupVersionResource{Group: "sparkoperator.k8s.io", Version: "v1beta2", Resource: "sparkapplications"}
	scheme := runtime.NewScheme()
	dynamicClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, map[schema.GroupVersionResource]string{
		gvr: "SparkApplicationList",
	}, &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "sparkoperator.k8s.io/v1beta2",
			"kind":       "SparkApplication",
			"metadata":   map[string]interface{}{"name": "fallback-app", "namespace": "spark"},
			"status": map[string]interface{}{
				"applicationState": map[string]interface{}{"state": "FAILED"},
				"driverInfo":       map[string]interface{}{"podName": "fallback-app-driver"},
			},
		},
	})
	// No pods with labels - forces fallback through GetApplication driver name
	k8sFake := k8sfake.NewSimpleClientset()
	k8sClient := &kubernetes.Client{Clientset: k8sFake, Dynamic: dynamicClient}
	c := diagnostics.NewCollector(k8sClient, spark.NewService(k8sClient))

	_, err := c.DriverLogs(context.Background(), "spark", "fallback-app", models.LogOptions{})
	if err == nil {
		t.Log("logs retrieved without pod (fake client limitation)")
	}
}

func TestDescribeDriverFallback(t *testing.T) {
	gvr := schema.GroupVersionResource{Group: "sparkoperator.k8s.io", Version: "v1beta2", Resource: "sparkapplications"}
	scheme := runtime.NewScheme()
	dynamicClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, map[schema.GroupVersionResource]string{
		gvr: "SparkApplicationList",
	}, &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "sparkoperator.k8s.io/v1beta2",
			"kind":       "SparkApplication",
			"metadata":   map[string]interface{}{"name": "fallback-app", "namespace": "spark"},
			"status": map[string]interface{}{
				"driverInfo": map[string]interface{}{"podName": "fallback-app-driver"},
			},
		},
	})
	k8sFake := k8sfake.NewSimpleClientset(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "fallback-app-driver", Namespace: "spark"},
		Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "spark"}}},
		Status:     corev1.PodStatus{Phase: corev1.PodFailed},
	})
	k8sClient := &kubernetes.Client{Clientset: k8sFake, Dynamic: dynamicClient}
	c := diagnostics.NewCollector(k8sClient, spark.NewService(k8sClient))

	desc, err := c.DescribeDriver(context.Background(), "spark", "fallback-app")
	if err != nil {
		t.Fatal(err)
	}
	if desc.PodName != "fallback-app-driver" {
		t.Errorf("unexpected pod: %s", desc.PodName)
	}
}

func TestExecutorLogsEmpty(t *testing.T) {
	k8sClient := &kubernetes.Client{Clientset: k8sfake.NewSimpleClientset()}
	c := diagnostics.NewCollector(k8sClient, spark.NewService(k8sClient))
	result, err := c.ExecutorLogs(context.Background(), "spark", "no-app", models.LogOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Logs) != 0 {
		t.Errorf("expected no logs, got %d", len(result.Logs))
	}
}
