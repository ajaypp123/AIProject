package kubernetes_test

import (
	"context"
	"strings"
	"testing"

	"github.com/spark-debug-mcp/spark-debug-mcp/internal/kubernetes"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestDescribePodTerminated(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "failed-pod", Namespace: "spark"},
		Spec:       corev1.PodSpec{NodeName: "node-1", Containers: []corev1.Container{{Name: "spark"}}},
		Status: corev1.PodStatus{
			Phase:    corev1.PodFailed,
			QOSClass: corev1.PodQOSGuaranteed,
			ContainerStatuses: []corev1.ContainerStatus{{
				Name: "spark", Ready: false, RestartCount: 3,
				State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{
					Reason: "CrashLoopBackOff", Message: "back-off 5m0s restarting failed container",
				}},
				LastTerminationState: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{
					Reason: "OOMKilled", Message: "memory limit exceeded", ExitCode: 137,
				}},
			}},
			Conditions: []corev1.PodCondition{{
				Type: corev1.PodScheduled, Status: corev1.ConditionFalse,
				Reason: "Unschedulable", Message: "0/3 nodes available",
			}},
		},
	}
	client := fake.NewSimpleClientset(pod)
	k8sClient := &kubernetes.Client{Clientset: client}
	desc, err := k8sClient.DescribePod(context.Background(), "spark", "failed-pod")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"CrashLoopBackOff", "OOMKilled", "Unschedulable"} {
		if !strings.Contains(desc, want) {
			t.Errorf("describe missing %s", want)
		}
	}
}

func TestFindSparkPodsByNamePattern(t *testing.T) {
	client := fake.NewSimpleClientset(
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "myapp-driver", Namespace: "spark"},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "myapp-exec-1", Namespace: "spark"},
		},
	)
	k8sClient := &kubernetes.Client{Clientset: client}
	driver, execs, err := k8sClient.FindSparkPods(context.Background(), "spark", "myapp")
	if err != nil {
		t.Fatal(err)
	}
	if driver == nil || driver.Name != "myapp-driver" {
		t.Error("expected driver by name pattern")
	}
	if len(execs) != 1 {
		t.Fatalf("expected 1 executor, got %d", len(execs))
	}
}

func TestEventFilterOOM(t *testing.T) {
	client := fake.NewSimpleClientset(
		&corev1.Event{
			ObjectMeta: metav1.ObjectMeta{Name: "ev1", Namespace: "spark"},
			Type:       "Normal", Reason: "Started", Message: "Started container",
			InvolvedObject: corev1.ObjectReference{Kind: "Pod", Name: "p1"},
		},
		&corev1.Event{
			ObjectMeta: metav1.ObjectMeta{Name: "ev2", Namespace: "spark"},
			Type:       "Normal", Reason: "Killing", Message: "Stopping container due to oom",
			InvolvedObject: corev1.ObjectReference{Kind: "Pod", Name: "p2"},
		},
	)
	k8sClient := &kubernetes.Client{Clientset: client}
	filtered, err := k8sClient.GetEvents(context.Background(), "spark", true)
	if err != nil {
		t.Fatal(err)
	}
	if len(filtered) != 1 {
		t.Fatalf("expected 1 filtered event, got %d", len(filtered))
	}
}

func TestPodResourcesMultipleContainers(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "multi", Namespace: "spark"},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "c1"}, {Name: "c2"},
			},
		},
		Status: corev1.PodStatus{
			QOSClass: corev1.PodQOSBestEffort,
			ContainerStatuses: []corev1.ContainerStatus{
				{Name: "c1", RestartCount: 1, State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}}},
				{Name: "c2", RestartCount: 2, State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{Reason: "Error"}}},
			},
		},
	}
	res := kubernetes.PodResourcesFromPod(pod)
	if res.RestartCount != 3 {
		t.Errorf("expected total restarts 3, got %d", res.RestartCount)
	}
	if len(res.Containers) != 2 {
		t.Fatalf("expected 2 containers, got %d", len(res.Containers))
	}
}
