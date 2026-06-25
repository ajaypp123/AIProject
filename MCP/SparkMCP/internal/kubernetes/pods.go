package kubernetes

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spark-debug-mcp/spark-debug-mcp/internal/models"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// ListPods lists pods in a namespace matching label selector.
func (c *Client) ListPods(ctx context.Context, namespace, labelSelector string) ([]corev1.Pod, error) {
	opts := metav1.ListOptions{}
	if labelSelector != "" {
		opts.LabelSelector = labelSelector
	}
	list, err := c.Clientset.CoreV1().Pods(namespace).List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("list pods: %w", err)
	}
	return list.Items, nil
}

// GetPod retrieves a pod by name.
func (c *Client) GetPod(ctx context.Context, namespace, name string) (*corev1.Pod, error) {
	pod, err := c.Clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get pod %s/%s: %w", namespace, name, err)
	}
	return pod, nil
}

// DescribePod generates kubectl-style describe output for a pod.
func (c *Client) DescribePod(ctx context.Context, namespace, name string) (string, error) {
	pod, err := c.GetPod(ctx, namespace, name)
	if err != nil {
		return "", err
	}
	return formatPodDescribe(pod), nil
}

func formatPodDescribe(pod *corev1.Pod) string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)

	fmt.Fprintf(w, "Name:\t%s\n", pod.Name)
	fmt.Fprintf(w, "Namespace:\t%s\n", pod.Namespace)
	if pod.Spec.Priority != nil {
		fmt.Fprintf(w, "Priority:\t%d\n", *pod.Spec.Priority)
	}
	fmt.Fprintf(w, "Node:\t%s\n", pod.Spec.NodeName)
	fmt.Fprintf(w, "Start Time:\t%s\n", formatTime(pod.Status.StartTime))
	fmt.Fprintf(w, "Phase:\t%s\n", pod.Status.Phase)
	fmt.Fprintf(w, "QoS Class:\t%s\n", pod.Status.QOSClass)
	fmt.Fprintf(w, "IP:\t%s\n", pod.Status.PodIP)

	fmt.Fprintf(w, "\nLabels:\n")
	for k, v := range pod.Labels {
		fmt.Fprintf(w, "  %s=%s\n", k, v)
	}

	fmt.Fprintf(w, "\nConditions:\n")
	for _, cond := range pod.Status.Conditions {
		fmt.Fprintf(w, "  Type=%s Status=%s Reason=%s Message=%s\n",
			cond.Type, cond.Status, cond.Reason, cond.Message)
	}

	fmt.Fprintf(w, "\nContainers:\n")
	for _, cs := range pod.Status.ContainerStatuses {
		state := containerState(cs)
		fmt.Fprintf(w, "  %s:\n", cs.Name)
		fmt.Fprintf(w, "    Ready:\t%v\n", cs.Ready)
		fmt.Fprintf(w, "    Restart Count:\t%d\n", cs.RestartCount)
		fmt.Fprintf(w, "    State:\t%s\n", state)
		if cs.LastTerminationState.Terminated != nil {
			t := cs.LastTerminationState.Terminated
			fmt.Fprintf(w, "    Last Termination Reason:\t%s\n", t.Reason)
			fmt.Fprintf(w, "    Last Termination Message:\t%s\n", t.Message)
			fmt.Fprintf(w, "    Exit Code:\t%d\n", t.ExitCode)
		}
	}

	fmt.Fprintf(w, "\nEvents:\n")
	for _, ev := range pod.Status.Conditions {
		if ev.Type == corev1.PodScheduled && ev.Status == corev1.ConditionFalse {
			fmt.Fprintf(w, "  %s: %s\n", ev.Reason, ev.Message)
		}
	}

	w.Flush()
	return buf.String()
}

func containerState(cs corev1.ContainerStatus) string {
	if cs.State.Running != nil {
		return "Running"
	}
	if cs.State.Waiting != nil {
		return fmt.Sprintf("Waiting (%s): %s", cs.State.Waiting.Reason, cs.State.Waiting.Message)
	}
	if cs.State.Terminated != nil {
		return fmt.Sprintf("Terminated (%s): exit %d", cs.State.Terminated.Reason, cs.State.Terminated.ExitCode)
	}
	return "Unknown"
}

func formatTime(t *metav1.Time) string {
	if t == nil {
		return "<none>"
	}
	return t.Format(time.RFC3339)
}

// PodResources collects resource info for a pod.
func (c *Client) PodResources(ctx context.Context, namespace, name string) (*models.PodResources, error) {
	pod, err := c.GetPod(ctx, namespace, name)
	if err != nil {
		return nil, err
	}
	return PodResourcesFromPod(pod), nil
}

// PodResourcesFromPod extracts resource info from a pod object.
func PodResourcesFromPod(pod *corev1.Pod) *models.PodResources {
	res := &models.PodResources{
		PodName:    pod.Name,
		Namespace:  pod.Namespace,
		Node:       pod.Spec.NodeName,
		QoS:        string(pod.Status.QOSClass),
		Containers: make([]models.ContainerRes, 0, len(pod.Spec.Containers)),
	}
	for _, cs := range pod.Status.ContainerStatuses {
		res.RestartCount += cs.RestartCount
	}
	for i, container := range pod.Spec.Containers {
		cr := models.ContainerRes{Name: container.Name}
		if container.Resources.Requests != nil {
			if cpu, ok := container.Resources.Requests[corev1.ResourceCPU]; ok {
				cr.CPURequest = cpu.String()
				if i == 0 {
					res.CPURequest = cpu.String()
				}
			}
			if mem, ok := container.Resources.Requests[corev1.ResourceMemory]; ok {
				cr.MemoryRequest = mem.String()
				if i == 0 {
					res.MemoryRequest = mem.String()
				}
			}
		}
		if container.Resources.Limits != nil {
			if cpu, ok := container.Resources.Limits[corev1.ResourceCPU]; ok {
				cr.CPULimit = cpu.String()
				if i == 0 {
					res.CPULimit = cpu.String()
				}
			}
			if mem, ok := container.Resources.Limits[corev1.ResourceMemory]; ok {
				cr.MemoryLimit = mem.String()
				if i == 0 {
					res.MemoryLimit = mem.String()
				}
			}
		}
		if i < len(pod.Status.ContainerStatuses) {
			cs := pod.Status.ContainerStatuses[i]
			cr.RestartCount = cs.RestartCount
			cr.State = containerState(cs)
		}
		res.Containers = append(res.Containers, cr)
	}
	return res
}

// FindPodsByLabel finds pods matching labels in namespace.
func (c *Client) FindPodsByLabel(ctx context.Context, namespace string, match labels.Set) ([]corev1.Pod, error) {
	selector := labels.SelectorFromSet(match)
	list, err := c.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

// FindSparkPods finds driver and executor pods for a Spark application.
func (c *Client) FindSparkPods(ctx context.Context, namespace, appName string) (driver *corev1.Pod, executors []corev1.Pod, err error) {
	pods, err := c.ListPods(ctx, namespace, fmt.Sprintf("sparkoperator.k8s.io/app-name=%s", appName))
	if err != nil || len(pods) == 0 {
		pods, err = c.ListPods(ctx, namespace, fmt.Sprintf("spark-app-name=%s", appName))
		if err != nil {
			return nil, nil, err
		}
	}
	if len(pods) == 0 {
		allPods, err := c.ListPods(ctx, namespace, "")
		if err != nil {
			return nil, nil, err
		}
		for _, p := range allPods {
			if strings.Contains(p.Name, appName) {
				pods = append(pods, p)
			}
		}
	}

	for i := range pods {
		pod := &pods[i]
		role := strings.ToLower(pod.Labels["spark-role"])
		if role == "" {
			role = strings.ToLower(pod.Labels["sparkoperator.k8s.io/role"])
		}
		switch role {
		case "driver":
			driver = pod
		case "executor":
			executors = append(executors, *pod)
		default:
			if strings.Contains(pod.Name, "-driver") {
				driver = pod
			} else if strings.Contains(pod.Name, "-exec-") {
				executors = append(executors, *pod)
			}
		}
	}
	return driver, executors, nil
}
