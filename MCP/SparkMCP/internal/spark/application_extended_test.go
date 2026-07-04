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

func TestGetStandaloneApplication(t *testing.T) {
	k8sClient := &kubernetes.Client{
		Clientset: k8sfake.NewSimpleClientset(
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "standalone-driver", Namespace: "spark",
					Labels: map[string]string{"spark-role": "driver", "spark-app-name": "standalone"},
				},
				Status: corev1.PodStatus{Phase: corev1.PodFailed},
			},
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "standalone-exec-1", Namespace: "spark",
					Labels: map[string]string{"spark-role": "executor", "spark-app-name": "standalone"},
				},
				Status: corev1.PodStatus{Phase: corev1.PodRunning},
			},
		),
	}
	svc := spark.NewService(k8sClient)
	app, err := svc.GetApplication(context.Background(), "spark", "standalone")
	if err != nil {
		t.Fatal(err)
	}
	if app.Name != "standalone" {
		t.Errorf("unexpected name: %s", app.Name)
	}
	if len(app.Executors) != 1 {
		t.Fatalf("expected 1 executor, got %d", len(app.Executors))
	}
}

func TestGetSparkApplicationYAML(t *testing.T) {
	svc := newSparkService()
	yaml, err := svc.GetSparkApplicationYAML(context.Background(), "default", "my-spark")
	if err != nil {
		t.Fatal(err)
	}
	if yaml == "" {
		t.Error("expected YAML output")
	}
}

func TestGetApplicationNotFound(t *testing.T) {
	k8sClient := &kubernetes.Client{Clientset: k8sfake.NewSimpleClientset()}
	svc := spark.NewService(k8sClient)
	_, err := svc.GetApplication(context.Background(), "spark", "missing")
	if err == nil {
		t.Error("expected error for missing application")
	}
}

func TestListApplicationsWithCR(t *testing.T) {
	gvr := schema.GroupVersionResource{Group: "sparkoperator.k8s.io", Version: "v1beta2", Resource: "sparkapplications"}
	scheme := runtime.NewScheme()
	dynamicClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, map[schema.GroupVersionResource]string{
		gvr: "SparkApplicationList",
	},
		&unstructured.Unstructured{Object: map[string]interface{}{
			"apiVersion": "sparkoperator.k8s.io/v1beta2", "kind": "SparkApplication",
			"metadata": map[string]interface{}{"name": "app-a", "namespace": "ns"},
			"status":   map[string]interface{}{"applicationState": map[string]interface{}{"state": "COMPLETED"}},
			"spec":     map[string]interface{}{"image": "spark:3.5.0"},
		}},
		&unstructured.Unstructured{Object: map[string]interface{}{
			"apiVersion": "sparkoperator.k8s.io/v1beta2", "kind": "SparkApplication",
			"metadata": map[string]interface{}{"name": "app-b", "namespace": "ns"},
			"status":   map[string]interface{}{"applicationState": map[string]interface{}{"state": "FAILED"}},
		}},
	)
	k8sClient := &kubernetes.Client{Clientset: k8sfake.NewSimpleClientset(), Dynamic: dynamicClient}
	svc := spark.NewService(k8sClient)
	apps, err := svc.ListApplications(context.Background(), "ns")
	if err != nil {
		t.Fatal(err)
	}
	if len(apps) != 2 {
		t.Fatalf("expected 2 apps, got %d", len(apps))
	}
}

func TestCollectSparkConfWithDriver(t *testing.T) {
	gvr := schema.GroupVersionResource{Group: "sparkoperator.k8s.io", Version: "v1beta2", Resource: "sparkapplications"}
	scheme := runtime.NewScheme()
	dynamicClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, map[schema.GroupVersionResource]string{
		gvr: "SparkApplicationList",
	}, &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "sparkoperator.k8s.io/v1beta2", "kind": "SparkApplication",
		"metadata": map[string]interface{}{"name": "conf-app", "namespace": "spark"},
		"spec": map[string]interface{}{
			"sparkConf": map[string]interface{}{"spark.driver.memory": "2g"},
			"deps":      map[string]interface{}{"jars": "local:///tmp/a.jar"},
		},
	}})
	k8sFake := k8sfake.NewSimpleClientset(
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "conf-app-driver", Namespace: "spark",
				Labels: map[string]string{"sparkoperator.k8s.io/app-name": "conf-app", "spark-role": "driver"},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Name: "spark",
					Env:  []corev1.EnvVar{{Name: "SPARK_HOME", Value: "/opt/spark"}},
				}},
				Volumes: []corev1.Volume{{Name: "cfg", VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{LocalObjectReference: corev1.LocalObjectReference{Name: "spark-cm"}},
				}}},
			},
		},
		&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: "spark-cm", Namespace: "spark"},
			Data:       map[string]string{"spark-defaults.conf": "spark.executor.instances=3"},
		},
	)
	k8sClient := &kubernetes.Client{Clientset: k8sFake, Dynamic: dynamicClient}
	svc := spark.NewService(k8sClient)
	conf, err := svc.CollectSparkConf(context.Background(), "spark", "conf-app")
	if err != nil {
		t.Fatal(err)
	}
	if conf.AllConf["spark.driver.memory"] != "2g" {
		t.Errorf("missing driver memory conf: %v", conf.AllConf)
	}
	if conf.FromEnv["SPARK_HOME"] != "/opt/spark" {
		t.Errorf("missing env conf: %v", conf.FromEnv)
	}
}
