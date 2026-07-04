// Package models defines shared data structures for Spark debugging.
package models

import (
	"time"

	corev1 "k8s.io/api/core/v1"
)

// OutputFormat specifies the report output format.
type OutputFormat string

const (
	FormatJSON     OutputFormat = "json"
	FormatMarkdown OutputFormat = "markdown"
	FormatConsole  OutputFormat = "console"
)

// SparkAppSummary is a lightweight view of a Spark application.
type SparkAppSummary struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	State      string `json:"state"`
	Age        string `json:"age"`
	Driver     string `json:"driver"`
	Executors  int    `json:"executors"`
	AppType    string `json:"appType,omitempty"`
	SparkImage string `json:"sparkImage,omitempty"`
}

// SparkApplicationDetail contains full application details.
type SparkApplicationDetail struct {
	Name            string                 `json:"name"`
	Namespace       string                 `json:"namespace"`
	YAML            string                 `json:"yaml"`
	Status          map[string]interface{} `json:"status"`
	Driver          PodSummary             `json:"driver"`
	Executors       []PodSummary           `json:"executors"`
	RestartCount    int32                  `json:"restartCount"`
	SubmissionTime  string                 `json:"submissionTime,omitempty"`
	CompletionTime  string                 `json:"completionTime,omitempty"`
	State           string                 `json:"state"`
	SparkVersion    string                 `json:"sparkVersion,omitempty"`
	MainClass       string                 `json:"mainClass,omitempty"`
	MainApplication string                 `json:"mainApplicationFile,omitempty"`
}

// PodSummary summarizes a Kubernetes pod.
type PodSummary struct {
	Name         string            `json:"name"`
	Namespace    string            `json:"namespace"`
	Phase        string            `json:"phase"`
	Node         string            `json:"node,omitempty"`
	RestartCount int32             `json:"restartCount"`
	Labels       map[string]string `json:"labels,omitempty"`
	StartTime    string            `json:"startTime,omitempty"`
	Ready        bool              `json:"ready"`
}

// LogOptions configures log retrieval.
type LogOptions struct {
	Previous bool   `json:"previous"`
	Tail     int64  `json:"tail,omitempty"`
	Since    string `json:"since,omitempty"`
	Executor string `json:"executor,omitempty"`
}

// PodLogs holds logs for a pod.
type PodLogs struct {
	PodName  string `json:"podName"`
	Logs     string `json:"logs"`
	Previous bool   `json:"previous"`
}

// ExecutorLogs holds logs from all executors.
type ExecutorLogs struct {
	Application string    `json:"application"`
	Namespace   string    `json:"namespace"`
	Logs        []PodLogs `json:"logs"`
}

// PodDescription holds kubectl describe output for a pod.
type PodDescription struct {
	PodName     string `json:"podName"`
	Namespace   string `json:"namespace"`
	Description string `json:"description"`
}

// EventSummary summarizes a Kubernetes event.
type EventSummary struct {
	Type      string    `json:"type"`
	Reason    string    `json:"reason"`
	Message   string    `json:"message"`
	Object    string    `json:"object"`
	Count     int32     `json:"count"`
	FirstSeen time.Time `json:"firstSeen"`
	LastSeen  time.Time `json:"lastSeen"`
}

// SparkConf holds collected Spark configuration.
type SparkConf struct {
	Application   string            `json:"application"`
	Namespace     string            `json:"namespace"`
	FromCR        map[string]string `json:"fromSparkApplication"`
	FromConfigMap map[string]string `json:"fromConfigMaps"`
	FromEnv       map[string]string `json:"fromEnvironment"`
	FromFiles     map[string]string `json:"fromMountedFiles"`
	AllConf       map[string]string `json:"allConf"`
}

// PodResources holds resource usage and limits for a pod.
type PodResources struct {
	PodName       string            `json:"podName"`
	Namespace     string            `json:"namespace"`
	Node          string            `json:"node,omitempty"`
	QoS           string            `json:"qos"`
	RestartCount  int32             `json:"restartCount"`
	CPURequest    string            `json:"cpuRequest,omitempty"`
	CPULimit      string            `json:"cpuLimit,omitempty"`
	MemoryRequest string            `json:"memoryRequest,omitempty"`
	MemoryLimit   string            `json:"memoryLimit,omitempty"`
	Containers    []ContainerRes    `json:"containers"`
}

// ContainerRes holds per-container resource info.
type ContainerRes struct {
	Name          string `json:"name"`
	CPURequest    string `json:"cpuRequest,omitempty"`
	CPULimit      string `json:"cpuLimit,omitempty"`
	MemoryRequest string `json:"memoryRequest,omitempty"`
	MemoryLimit   string `json:"memoryLimit,omitempty"`
	RestartCount  int32  `json:"restartCount"`
	State         string `json:"state"`
}

// RuleSeverity indicates the severity of a detected issue.
type RuleSeverity string

const (
	SeverityCritical RuleSeverity = "critical"
	SeverityHigh     RuleSeverity = "high"
	SeverityMedium   RuleSeverity = "medium"
	SeverityLow      RuleSeverity = "low"
	SeverityInfo     RuleSeverity = "info"
)

// RuleMatch represents a rule that matched during analysis.
type RuleMatch struct {
	ID             string       `json:"id"`
	Name           string       `json:"name"`
	Description    string       `json:"description"`
	Severity       RuleSeverity `json:"severity"`
	Evidence       []string     `json:"evidence"`
	Recommendation string       `json:"recommendation"`
	Score          float64      `json:"score"`
	DocLinks       []string     `json:"docLinks,omitempty"`
}

// ParsedLog holds parsed log entries.
type ParsedLog struct {
	Level      string   `json:"level,omitempty"`
	Message    string   `json:"message"`
	Timestamp  string   `json:"timestamp,omitempty"`
	Exceptions []string `json:"exceptions,omitempty"`
	StackTrace string   `json:"stackTrace,omitempty"`
	Source     string   `json:"source"`
}

// TimelineEntry represents an event in the failure timeline.
type TimelineEntry struct {
	Timestamp   time.Time `json:"timestamp"`
	Source      string    `json:"source"`
	Event       string    `json:"event"`
	Description string    `json:"description"`
	Severity    string    `json:"severity,omitempty"`
}

// FailureReport is the comprehensive analysis output.
type FailureReport struct {
	Application      string            `json:"application"`
	Namespace        string            `json:"namespace"`
	FailureSummary   string            `json:"failureSummary"`
	LikelyRootCause  string            `json:"likelyRootCause"`
	Confidence       float64           `json:"confidence"`
	Evidence         []string          `json:"evidence"`
	KubernetesEvents []EventSummary    `json:"kubernetesEvents"`
	RelevantLogs     []ParsedLog       `json:"relevantLogs"`
	SparkConfig      SparkConf         `json:"sparkConfig"`
	Recommendations  []Recommendation  `json:"recommendations"`
	MatchedRules     []RuleMatch       `json:"matchedRules"`
	Timeline         []TimelineEntry   `json:"timeline"`
	ClusterHealth    float64           `json:"clusterHealthScore"`
	FailureScore     float64           `json:"failureScore"`
	LogSummary       string            `json:"logSummary,omitempty"`
	ApplicationDetail SparkApplicationDetail `json:"applicationDetail,omitempty"`
	GeneratedAt      time.Time         `json:"generatedAt"`
}

// Recommendation is a prioritized remediation suggestion.
type Recommendation struct {
	Priority    int    `json:"priority"`
	Action      string `json:"action"`
	Description string `json:"description"`
	DocLink     string `json:"docLink,omitempty"`
}

// AnalysisContext holds all collected data for rule evaluation.
type AnalysisContext struct {
	Application   SparkApplicationDetail
	DriverLogs    string
	ExecutorLogs  []PodLogs
	Events        []EventSummary
	SparkConf     SparkConf
	DriverDesc    string
	ExecutorDescs map[string]string
	DriverRes     PodResources
	ExecutorRes   []PodResources
	ParsedLogs    []ParsedLog
}

// PodFromCore converts a corev1 Pod to PodSummary.
func PodFromCore(pod *corev1.Pod) PodSummary {
	summary := PodSummary{
		Name:      pod.Name,
		Namespace: pod.Namespace,
		Phase:     string(pod.Status.Phase),
		Node:      pod.Spec.NodeName,
		Labels:    pod.Labels,
	}
	if pod.Status.StartTime != nil {
		summary.StartTime = pod.Status.StartTime.Format(time.RFC3339)
	}
	for _, cs := range pod.Status.ContainerStatuses {
		summary.RestartCount += cs.RestartCount
		summary.Ready = cs.Ready
	}
	return summary
}
