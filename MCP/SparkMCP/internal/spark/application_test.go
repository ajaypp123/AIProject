package spark_test

import (
	"context"
	"testing"

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

func newSparkService() *spark.Service {
	gvr := schema.GroupVersionResource{Group: "sparkoperator.k8s.io", Version: "v1beta2", Resource: "sparkapplications"}
	scheme := runtime.NewScheme()
	dynamicClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, map[schema.GroupVersionResource]string{
		gvr: "SparkApplicationList",
	}, &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "sparkoperator.k8s.io/v1beta2",
			"kind":       "SparkApplication",
			"metadata": map[string]interface{}{
				"name":      "my-spark",
				"namespace": "default",
			},
			"status": map[string]interface{}{
				"applicationState": map[string]interface{}{"state": "RUNNING"},
			},
			"spec": map[string]interface{}{
				"image":        "spark:3.5.0",
				"sparkVersion": "3.5.0",
			},
		},
	})
	k8sClient := &kubernetes.Client{
		Clientset: k8sfake.NewSimpleClientset(),
		Dynamic:   dynamicClient,
	}
	return spark.NewService(k8sClient)
}

func TestListApplications(t *testing.T) {
	svc := newSparkService()
	apps, err := svc.ListApplications(context.Background(), "default")
	if err != nil {
		t.Fatal(err)
	}
	if len(apps) != 1 || apps[0].Name != "my-spark" {
		t.Fatalf("unexpected apps: %+v", apps)
	}
}

func TestGetApplication(t *testing.T) {
	svc := newSparkService()
	app, err := svc.GetApplication(context.Background(), "default", "my-spark")
	if err != nil {
		t.Fatal(err)
	}
	if app.State != "RUNNING" {
		t.Errorf("expected RUNNING, got %s", app.State)
	}
}

func TestStandaloneList(t *testing.T) {
	k8sClient := &kubernetes.Client{
		Clientset: k8sfake.NewSimpleClientset(&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "standalone-driver", Namespace: "spark",
				Labels: map[string]string{"spark-role": "driver", "spark-app-name": "standalone"},
			},
			Status: corev1.PodStatus{Phase: corev1.PodRunning},
		}),
	}
	svc := spark.NewService(k8sClient)
	apps, err := svc.ListApplications(context.Background(), "spark")
	if err != nil {
		t.Fatal(err)
	}
	if len(apps) != 1 {
		t.Fatalf("expected 1 standalone app, got %d", len(apps))
	}
	if apps[0].AppType != "standalone" {
		t.Errorf("expected standalone type, got %s", apps[0].AppType)
	}
}

func TestCollectSparkConf(t *testing.T) {
	svc := newSparkService()
	conf, err := svc.CollectSparkConf(context.Background(), "default", "my-spark")
	if err != nil {
		t.Fatal(err)
	}
	if conf.Application != "my-spark" {
		t.Errorf("unexpected application: %s", conf.Application)
	}
}
