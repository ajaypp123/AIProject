package kubernetes

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/spark-debug-mcp/spark-debug-mcp/internal/models"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var eventFilterReasons = map[string]bool{
	"Warning":           true,
	"FailedScheduling":  true,
	"OOMKilled":         true,
	"BackOff":           true,
	"Evicted":           true,
	"DeadlineExceeded":  true,
	"NodePressure":      true,
	"FailedMount":       true,
	"FailedAttachVolume": true,
	"Unhealthy":         true,
	"Failed":            true,
	"Killing":           true,
}

// GetEvents collects namespace events, optionally filtered.
func (c *Client) GetEvents(ctx context.Context, namespace string, filter bool) ([]models.EventSummary, error) {
	list, err := c.Clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}

	summaries := make([]models.EventSummary, 0, len(list.Items))
	for _, ev := range list.Items {
		if filter && !shouldIncludeEvent(&ev) {
			continue
		}
		summaries = append(summaries, eventToSummary(&ev))
	}
	return summaries, nil
}

// GetEventsForObject collects events for a specific object.
func (c *Client) GetEventsForObject(ctx context.Context, namespace, kind, name string) ([]models.EventSummary, error) {
	fieldSelector := fmt.Sprintf("involvedObject.kind=%s,involvedObject.name=%s", kind, name)
	list, err := c.Clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return nil, err
	}
	summaries := make([]models.EventSummary, 0, len(list.Items))
	for _, ev := range list.Items {
		summaries = append(summaries, eventToSummary(&ev))
	}
	return summaries, nil
}

func shouldIncludeEvent(ev *corev1.Event) bool {
	if ev.Type == "Warning" {
		return true
	}
	if eventFilterReasons[ev.Reason] {
		return true
	}
	msg := strings.ToLower(ev.Message)
	keywords := []string{"oom", "backoff", "evict", "pressure", "failed", "timeout", "denied", "pull"}
	for _, kw := range keywords {
		if strings.Contains(msg, kw) {
			return true
		}
	}
	return false
}

func eventToSummary(ev *corev1.Event) models.EventSummary {
	return models.EventSummary{
		Type:      ev.Type,
		Reason:    ev.Reason,
		Message:   ev.Message,
		Object:    fmt.Sprintf("%s/%s", ev.InvolvedObject.Kind, ev.InvolvedObject.Name),
		Count:     ev.Count,
		FirstSeen: ev.FirstTimestamp.Time,
		LastSeen:  ev.LastTimestamp.Time,
	}
}

// GetPodLogs retrieves logs from a pod.
func (c *Client) GetPodLogs(ctx context.Context, namespace, podName string, opts models.LogOptions) (string, error) {
	logOpts := &corev1.PodLogOptions{
		Previous: opts.Previous,
	}
	if opts.Tail > 0 {
		logOpts.TailLines = &opts.Tail
	}
	if opts.Since != "" {
		d, err := time.ParseDuration(opts.Since)
		if err == nil {
			sec := int64(d.Seconds())
			logOpts.SinceSeconds = &sec
		}
	}

	req := c.Clientset.CoreV1().Pods(namespace).GetLogs(podName, logOpts)
	stream, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("stream logs for %s: %w", podName, err)
	}
	defer stream.Close()

	data, err := io.ReadAll(stream)
	if err != nil {
		return "", fmt.Errorf("read logs: %w", err)
	}
	return string(data), nil
}
