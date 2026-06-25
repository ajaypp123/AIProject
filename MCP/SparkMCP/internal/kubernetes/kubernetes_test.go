package kubernetes_test

import (
	"context"
	"testing"

	"github.com/spark-debug-mcp/spark-debug-mcp/internal/kubernetes"
	"github.com/spark-debug-mcp/spark-debug-mcp/internal/models"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestPodResourcesFromPod(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "driver", Namespace: "spark"},
		Spec: corev1.PodSpec{
			NodeName: "node-1",
			Containers: []corev1.Container{
				{
					Name: "spark-kubernetes-driver",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("2"),
							corev1.ResourceMemory: resource.MustParse("4Gi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("4"),
							corev1.ResourceMemory: resource.MustParse("8Gi"),
						},
					},
				},
			},
		},
		Status: corev1.PodStatus{
			QOSClass: corev1.PodQOSBurstable,
			ContainerStatuses: []corev1.ContainerStatus{
				{Name: "spark-kubernetes-driver", Ready: true, RestartCount: 2},
			},
		},
	}

	res := kubernetes.PodResourcesFromPod(pod)
	if res.CPURequest != "2" {
		t.Errorf("expected CPU request 2, got %s", res.CPURequest)
	}
	if res.RestartCount != 2 {
		t.Errorf("expected restart count 2, got %d", res.RestartCount)
	}
	if res.QoS != "Burstable" {
		t.Errorf("expected Burstable QoS, got %s", res.QoS)
	}
}

func TestGetPodLogsMock(t *testing.T) {
	// Verify fake client can be used for interface testing
	client := fake.NewSimpleClientset()
	_, err := client.CoreV1().Pods("default").Get(context.Background(), "test", metav1.GetOptions{})
	if err == nil {
		t.Error("expected not found error")
	}
}

func TestShouldIncludeEventPatterns(t *testing.T) {
	// Test via GetEvents with fake client
	client := fake.NewSimpleClientset(&corev1.Event{
		ObjectMeta: metav1.ObjectMeta{Name: "ev1", Namespace: "spark"},
		Type:       "Warning",
		Reason:     "FailedScheduling",
		Message:    "Insufficient cpu",
		InvolvedObject: corev1.ObjectReference{Kind: "Pod", Name: "driver"},
	})

	k8sClient := &kubernetes.Client{Clientset: client}
	events, err := k8sClient.GetEvents(context.Background(), "spark", true)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Reason != "FailedScheduling" {
		t.Errorf("unexpected reason: %s", events[0].Reason)
	}
}

func TestPodFromCore(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "exec-1", Namespace: "spark",
			Labels: map[string]string{"spark-role": "executor"},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				{Ready: true, RestartCount: 1},
			},
		},
	}
	summary := models.PodFromCore(pod)
	if summary.Name != "exec-1" {
		t.Errorf("unexpected name: %s", summary.Name)
	}
	if summary.RestartCount != 1 {
		t.Errorf("unexpected restart count: %d", summary.RestartCount)
	}
}

func TestFindSparkPodsByName(t *testing.T) {
	client := fake.NewSimpleClientset(
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "myapp-driver", Namespace: "spark",
				Labels: map[string]string{
					"sparkoperator.k8s.io/app-name": "myapp",
					"spark-role":                      "driver",
				},
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "myapp-exec-1", Namespace: "spark",
				Labels: map[string]string{
					"sparkoperator.k8s.io/app-name": "myapp",
					"spark-role":                      "executor",
				},
			},
		},
	)
	k8sClient := &kubernetes.Client{Clientset: client}
	driver, execs, err := k8sClient.FindSparkPods(context.Background(), "spark", "myapp")
	if err != nil {
		t.Fatal(err)
	}
	if driver == nil {
		t.Fatal("expected driver pod")
	}
	if len(execs) != 1 {
		t.Fatalf("expected 1 executor, got %d", len(execs))
	}
}
