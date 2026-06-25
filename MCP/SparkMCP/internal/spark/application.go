// Package spark provides Spark Operator CR and standalone Spark support.
package spark

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spark-debug-mcp/spark-debug-mcp/internal/kubernetes"
	"github.com/spark-debug-mcp/spark-debug-mcp/internal/models"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	sparkAppGroup    = "sparkoperator.k8s.io"
	sparkAppVersion  = "v1beta2"
	sparkAppResource = "sparkapplications"
)

var sparkAppGVR = schema.GroupVersionResource{
	Group:    sparkAppGroup,
	Version:  sparkAppVersion,
	Resource: sparkAppResource,
}

// Service provides Spark application operations.
type Service struct {
	k8s *kubernetes.Client
}

// NewService creates a Spark service.
func NewService(k8s *kubernetes.Client) *Service {
	return &Service{k8s: k8s}
}

// ListApplications lists Spark applications in a namespace.
func (s *Service) ListApplications(ctx context.Context, namespace string) ([]models.SparkAppSummary, error) {
	if s.k8s.Dynamic == nil {
		return s.listStandaloneApps(ctx, namespace)
	}
	list, err := s.k8s.Dynamic.Resource(sparkAppGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return s.listStandaloneApps(ctx, namespace)
	}

	summaries := make([]models.SparkAppSummary, 0, len(list.Items))
	for _, item := range list.Items {
		summaries = append(summaries, s.summaryFromCR(&item))
	}
	return summaries, nil
}

func (s *Service) listStandaloneApps(ctx context.Context, namespace string) ([]models.SparkAppSummary, error) {
	pods, err := s.k8s.ListPods(ctx, namespace, "spark-role=driver")
	if err != nil {
		return nil, err
	}
	summaries := make([]models.SparkAppSummary, 0, len(pods))
	for _, pod := range pods {
		appName := pod.Labels["spark-app-name"]
		if appName == "" {
			appName = strings.TrimSuffix(pod.Name, "-driver")
		}
		_, execs, _ := s.k8s.FindSparkPods(ctx, namespace, appName)
		summaries = append(summaries, models.SparkAppSummary{
			Name:      appName,
			Namespace: namespace,
			State:     string(pod.Status.Phase),
			Age:       ageSince(pod.CreationTimestamp.Time),
			Driver:    pod.Name,
			Executors: len(execs),
			AppType:   "standalone",
		})
	}
	return summaries, nil
}

func (s *Service) summaryFromCR(obj *unstructured.Unstructured) models.SparkAppSummary {
	name := obj.GetName()
	ns := obj.GetNamespace()
	state, _, _ := unstructured.NestedString(obj.Object, "status", "applicationState", "state")
	driverPod, _, _ := unstructured.NestedString(obj.Object, "status", "driverInfo", "podName")
	execCount, _, _ := unstructured.NestedInt64(obj.Object, "status", "executorState", "size")

	if driverPod == "" {
		driverPod = name + "-driver"
	}

	image, _, _ := unstructured.NestedString(obj.Object, "spec", "image")

	return models.SparkAppSummary{
		Name:       name,
		Namespace:  ns,
		State:      state,
		Age:        ageSince(obj.GetCreationTimestamp().Time),
		Driver:     driverPod,
		Executors:  int(execCount),
		AppType:    "spark-operator",
		SparkImage: image,
	}
}

// GetApplication returns detailed Spark application info.
func (s *Service) GetApplication(ctx context.Context, namespace, name string) (*models.SparkApplicationDetail, error) {
	if s.k8s.Dynamic != nil {
		obj, err := s.k8s.Dynamic.Resource(sparkAppGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
		if err == nil {
			return s.detailFromCR(ctx, obj)
		}
	}
	return s.getStandaloneApplication(ctx, namespace, name)
}

func (s *Service) detailFromCR(ctx context.Context, obj *unstructured.Unstructured) (*models.SparkApplicationDetail, error) {
	yamlBytes, _ := yaml.Marshal(obj.Object)
	name := obj.GetName()
	ns := obj.GetNamespace()

	state, _, _ := unstructured.NestedString(obj.Object, "status", "applicationState", "state")
	errorMsg, _, _ := unstructured.NestedString(obj.Object, "status", "applicationState", "errorMessage")
	subTime, _, _ := unstructured.NestedString(obj.Object, "status", "lastSubmissionAttemptTime")
	compTime, _, _ := unstructured.NestedString(obj.Object, "status", "terminationTime")
	sparkVer, _, _ := unstructured.NestedString(obj.Object, "spec", "sparkVersion")
	mainClass, _, _ := unstructured.NestedString(obj.Object, "spec", "mainClass")
	mainApp, _, _ := unstructured.NestedString(obj.Object, "spec", "mainApplicationFile")

	status := map[string]interface{}{
		"state": state,
	}
	if errorMsg != "" {
		status["errorMessage"] = errorMsg
	}

	driver, executors, _ := s.k8s.FindSparkPods(ctx, ns, name)
	detail := &models.SparkApplicationDetail{
		Name:            name,
		Namespace:       ns,
		YAML:            string(yamlBytes),
		Status:          status,
		State:           state,
		SubmissionTime:  subTime,
		CompletionTime:  compTime,
		SparkVersion:    sparkVer,
		MainClass:       mainClass,
		MainApplication: mainApp,
	}

	if driver != nil {
		detail.Driver = models.PodFromCore(driver)
		detail.RestartCount = detail.Driver.RestartCount
	} else {
		driverName, _, _ := unstructured.NestedString(obj.Object, "status", "driverInfo", "podName")
		if driverName != "" {
			if pod, err := s.k8s.GetPod(ctx, ns, driverName); err == nil {
				detail.Driver = models.PodFromCore(pod)
				detail.RestartCount = detail.Driver.RestartCount
			}
		}
	}

	for _, exec := range executors {
		detail.Executors = append(detail.Executors, models.PodFromCore(&exec))
		detail.RestartCount += models.PodFromCore(&exec).RestartCount
	}

	return detail, nil
}

func (s *Service) getStandaloneApplication(ctx context.Context, namespace, name string) (*models.SparkApplicationDetail, error) {
	driver, executors, err := s.k8s.FindSparkPods(ctx, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("application %s not found: %w", name, err)
	}
	if driver == nil {
		return nil, fmt.Errorf("application %s not found in namespace %s", name, namespace)
	}

	detail := &models.SparkApplicationDetail{
		Name:      name,
		Namespace: namespace,
		State:     string(driver.Status.Phase),
		Status:    map[string]interface{}{"state": string(driver.Status.Phase)},
		Driver:    models.PodFromCore(driver),
	}
	for _, exec := range executors {
		detail.Executors = append(detail.Executors, models.PodFromCore(&exec))
	}
	detail.RestartCount = detail.Driver.RestartCount
	for _, e := range detail.Executors {
		detail.RestartCount += e.RestartCount
	}
	return detail, nil
}

// GetSparkApplicationYAML returns raw YAML for a SparkApplication CR.
func (s *Service) GetSparkApplicationYAML(ctx context.Context, namespace, name string) (string, error) {
	obj, err := s.k8s.Dynamic.Resource(sparkAppGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	yamlBytes, err := yaml.Marshal(obj.Object)
	if err != nil {
		return "", err
	}
	return string(yamlBytes), nil
}

func ageSince(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}

// CollectSparkConf gathers Spark configuration from multiple sources.
func (s *Service) CollectSparkConf(ctx context.Context, namespace, appName string) (*models.SparkConf, error) {
	conf := &models.SparkConf{
		Application:   appName,
		Namespace:     namespace,
		FromCR:        make(map[string]string),
		FromConfigMap: make(map[string]string),
		FromEnv:       make(map[string]string),
		FromFiles:     make(map[string]string),
		AllConf:       make(map[string]string),
	}

	obj, err := s.k8s.Dynamic.Resource(sparkAppGVR).Namespace(namespace).Get(ctx, appName, metav1.GetOptions{})
	if err == nil {
		sparkConf, found, _ := unstructured.NestedStringMap(obj.Object, "spec", "sparkConf")
		if found {
			for k, v := range sparkConf {
				conf.FromCR[k] = v
				conf.AllConf[k] = v
			}
		}
		deps, found, _ := unstructured.NestedMap(obj.Object, "spec", "deps")
		if found {
			for k, v := range deps {
				key := "spark.deps." + k
				conf.FromCR[key] = fmt.Sprintf("%v", v)
				conf.AllConf[key] = fmt.Sprintf("%v", v)
			}
		}
	}

	driver, _, _ := s.k8s.FindSparkPods(ctx, namespace, appName)
	if driver != nil {
		for _, container := range driver.Spec.Containers {
			for _, env := range container.Env {
				if strings.HasPrefix(env.Name, "SPARK_") || strings.HasPrefix(env.Name, "spark.") {
					val := env.Value
					if val == "" && env.ValueFrom != nil {
						val = fmt.Sprintf("%v", env.ValueFrom)
					}
					conf.FromEnv[env.Name] = val
					conf.AllConf[env.Name] = val
				}
			}
		}
		s.collectConfigMaps(ctx, namespace, driver, conf)
	}

	return conf, nil
}

func (s *Service) collectConfigMaps(ctx context.Context, namespace string, pod *corev1.Pod, conf *models.SparkConf) {
	for _, vol := range pod.Spec.Volumes {
		if vol.ConfigMap != nil {
			cm, err := s.k8s.Clientset.CoreV1().ConfigMaps(namespace).Get(ctx, vol.ConfigMap.Name, metav1.GetOptions{})
			if err != nil {
				continue
			}
			for k, v := range cm.Data {
				conf.FromConfigMap[k] = v
				if strings.Contains(k, "spark") || strings.HasPrefix(v, "spark.") {
					conf.AllConf[k] = v
				}
			}
		}
	}
}
