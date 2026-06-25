package diagnostics_test

import (
	"context"
	"testing"

	"github.com/spark-debug-mcp/spark-debug-mcp/internal/diagnostics"
	"github.com/spark-debug-mcp/spark-debug-mcp/internal/kubernetes"
	"github.com/spark-debug-mcp/spark-debug-mcp/internal/spark"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func setupCollector(t *testing.T) *diagnostics.Collector {
	t.Helper()
	scheme := runtime.NewScheme()
	gvr := schema.GroupVersionResource{Group: "sparkoperator.k8s.io", Version: "v1beta2", Resource: "sparkapplications"}
	dynamicClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, map[schema.GroupVersionResource]string{
		gvr: "SparkApplicationList",
	}, &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "sparkoperator.k8s.io/v1beta2",
			"kind":       "SparkApplication",
			"metadata": map[string]interface{}{
				"name":      "test-app",
				"namespace": "spark",
			},
			"status": map[string]interface{}{
				"applicationState": map[string]interface{}{
					"state": "FAILED",
				},
				"driverInfo": map[string]interface{}{
					"podName": "test-app-driver",
				},
			},
			"spec": map[string]interface{}{
				"sparkVersion": "3.5.0",
				"sparkConf": map[string]interface{}{
					"spark.executor.memory": "1g",
				},
			},
		},
	})

	k8sFake := k8sfake.NewSimpleClientset(
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-app-driver", Namespace: "spark",
				Labels: map[string]string{
					"sparkoperator.k8s.io/app-name": "test-app",
					"spark-role":                      "driver",
				},
			},
			Status: corev1.PodStatus{Phase: corev1.PodFailed},
		},
		&corev1.Event{
			ObjectMeta: metav1.ObjectMeta{Name: "ev1", Namespace: "spark"},
			Type:       "Warning",
			Reason:     "Failed",
			Message:    "Pod failed",
			InvolvedObject: corev1.ObjectReference{Kind: "Pod", Name: "test-app-driver"},
		},
	)

	k8sClient := &kubernetes.Client{Clientset: k8sFake, Dynamic: dynamicClient}
	sparkSvc := spark.NewService(k8sClient)
	return diagnostics.NewCollector(k8sClient, sparkSvc)
}

func TestListApps(t *testing.T) {
	c := setupCollector(t)
	apps, err := c.ListApps(context.Background(), "spark")
	if err != nil {
		t.Fatal(err)
	}
	if len(apps) != 1 {
		t.Fatalf("expected 1 app, got %d", len(apps))
	}
	if apps[0].Name != "test-app" {
		t.Errorf("unexpected app name: %s", apps[0].Name)
	}
}

func TestGetApplication(t *testing.T) {
	c := setupCollector(t)
	app, err := c.GetApplication(context.Background(), "spark", "test-app")
	if err != nil {
		t.Fatal(err)
	}
	if app.State != "FAILED" {
		t.Errorf("expected FAILED state, got %s", app.State)
	}
}

func TestGetEvents(t *testing.T) {
	c := setupCollector(t)
	events, err := c.GetEvents(context.Background(), "spark")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) == 0 {
		t.Error("expected events")
	}
}

func TestSparkConf(t *testing.T) {
	c := setupCollector(t)
	conf, err := c.SparkConf(context.Background(), "spark", "test-app")
	if err != nil {
		t.Fatal(err)
	}
	if conf.AllConf["spark.executor.memory"] != "1g" {
		t.Errorf("unexpected conf: %v", conf.AllConf)
	}
}

func TestDescribeDriver(t *testing.T) {
	c := setupCollector(t)
	desc, err := c.DescribeDriver(context.Background(), "spark", "test-app")
	if err != nil {
		t.Fatal(err)
	}
	if desc.PodName != "test-app-driver" {
		t.Errorf("unexpected pod: %s", desc.PodName)
	}
}

func TestPodResources(t *testing.T) {
	c := setupCollector(t)
	res, err := c.PodResources(context.Background(), "spark", "test-app")
	if err != nil {
		t.Fatal(err)
	}
	if len(res) == 0 {
		t.Error("expected pod resources")
	}
}

func TestCollectAll(t *testing.T) {
	c := setupCollector(t)
	ctx, err := c.CollectAll(context.Background(), "spark", "test-app")
	if err != nil {
		t.Fatal(err)
	}
	if ctx.Application.Name != "test-app" {
		t.Errorf("unexpected app: %s", ctx.Application.Name)
	}
}
