package kubernetes_test

import (
	"context"
	"strings"
	"testing"

	"github.com/spark-debug-mcp/spark-debug-mcp/internal/kubernetes"
	"github.com/spark-debug-mcp/spark-debug-mcp/internal/models"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

func TestDescribePod(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "driver", Namespace: "spark"},
		Spec:       corev1.PodSpec{NodeName: "node-1", Containers: []corev1.Container{{Name: "spark"}}},
		Status: corev1.PodStatus{
			Phase:    corev1.PodRunning,
			QOSClass: corev1.PodQOSBurstable,
			ContainerStatuses: []corev1.ContainerStatus{
				{Name: "spark", Ready: true, State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}}},
			},
		},
	}
	client := fake.NewSimpleClientset(pod)
	k8sClient := &kubernetes.Client{Clientset: client}
	desc, err := k8sClient.DescribePod(context.Background(), "spark", "driver")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(desc, "driver") {
		t.Error("describe should contain pod name")
	}
}

func TestGetPod(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
	}
	client := fake.NewSimpleClientset(pod)
	k8sClient := &kubernetes.Client{Clientset: client}
	got, err := k8sClient.GetPod(context.Background(), "default", "test")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "test" {
		t.Errorf("unexpected pod: %s", got.Name)
	}
}

func TestListPods(t *testing.T) {
	client := fake.NewSimpleClientset(
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "a", Namespace: "ns", Labels: map[string]string{"app": "x"}}},
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "b", Namespace: "ns", Labels: map[string]string{"app": "y"}}},
	)
	k8sClient := &kubernetes.Client{Clientset: client}
	pods, err := k8sClient.ListPods(context.Background(), "ns", "app=x")
	if err != nil {
		t.Fatal(err)
	}
	if len(pods) != 1 {
		t.Fatalf("expected 1 pod, got %d", len(pods))
	}
}

func TestGetEventsUnfiltered(t *testing.T) {
	client := fake.NewSimpleClientset(&corev1.Event{
		ObjectMeta: metav1.ObjectMeta{Name: "ev", Namespace: "ns"},
		Type:       "Normal",
		Reason:     "Started",
		Message:    "Started container",
		InvolvedObject: corev1.ObjectReference{Kind: "Pod", Name: "p"},
	})
	k8sClient := &kubernetes.Client{Clientset: client}
	events, err := k8sClient.GetEvents(context.Background(), "ns", false)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
}

func TestGetEventsForObject(t *testing.T) {
	client := fake.NewSimpleClientset(&corev1.Event{
		ObjectMeta: metav1.ObjectMeta{Name: "ev", Namespace: "ns"},
		Type:       "Warning", Reason: "Failed",
		InvolvedObject: corev1.ObjectReference{Kind: "Pod", Name: "driver"},
	})
	k8sClient := &kubernetes.Client{Clientset: client}
	events, err := k8sClient.GetEventsForObject(context.Background(), "ns", "Pod", "driver")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
}

func TestPodResources(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "driver", Namespace: "spark"},
		Spec: corev1.PodSpec{
			NodeName: "n1",
			Containers: []corev1.Container{{Name: "c1"}},
		},
		Status: corev1.PodStatus{QOSClass: corev1.PodQOSBestEffort},
	}
	client := fake.NewSimpleClientset(pod)
	k8sClient := &kubernetes.Client{Clientset: client}
	res, err := k8sClient.PodResources(context.Background(), "spark", "driver")
	if err != nil {
		t.Fatal(err)
	}
	if res.PodName != "driver" {
		t.Errorf("unexpected pod: %s", res.PodName)
	}
}

func TestPing(t *testing.T) {
	client := fake.NewSimpleClientset()
	k8sClient := &kubernetes.Client{Clientset: client, RESTConfig: &rest.Config{Host: "http://localhost"}}
	// Ping may fail with fake client but should not panic
	_ = k8sClient.Ping(context.Background())
}

func TestLogOptionsSince(t *testing.T) {
	// Verify LogOptions struct is usable
	opts := models.LogOptions{Since: "1h", Tail: 100, Previous: true}
	if opts.Since != "1h" {
		t.Error("unexpected since value")
	}
}
