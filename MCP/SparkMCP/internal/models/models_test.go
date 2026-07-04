package models_test

import (
	"testing"
	"time"

	"github.com/spark-debug-mcp/spark-debug-mcp/internal/models"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPodFromCore(t *testing.T) {
	start := metav1.NewTime(time.Now())
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "driver", Namespace: "spark",
			Labels: map[string]string{"app": "test"},
		},
		Spec: corev1.PodSpec{NodeName: "node-1"},
		Status: corev1.PodStatus{
			Phase:     corev1.PodRunning,
			StartTime: &start,
			ContainerStatuses: []corev1.ContainerStatus{
				{Ready: true, RestartCount: 2},
				{Ready: false, RestartCount: 1},
			},
		},
	}
	summary := models.PodFromCore(pod)
	if summary.Name != "driver" {
		t.Errorf("unexpected name: %s", summary.Name)
	}
	if summary.RestartCount != 3 {
		t.Errorf("expected restart count 3, got %d", summary.RestartCount)
	}
	if summary.Node != "node-1" {
		t.Errorf("unexpected node: %s", summary.Node)
	}
	if summary.StartTime == "" {
		t.Error("expected start time")
	}
}
