// Package diagnostics provides data collection for Spark debugging.
package diagnostics

import (
	"context"
	"fmt"

	"github.com/spark-debug-mcp/spark-debug-mcp/internal/kubernetes"
	"github.com/spark-debug-mcp/spark-debug-mcp/internal/models"
	"github.com/spark-debug-mcp/spark-debug-mcp/internal/spark"
)

// Collector orchestrates diagnostic data collection.
type Collector struct {
	k8s   *kubernetes.Client
	spark *spark.Service
}

// NewCollector creates a diagnostics collector.
func NewCollector(k8s *kubernetes.Client, sparkSvc *spark.Service) *Collector {
	return &Collector{k8s: k8s, spark: sparkSvc}
}

// ListApps lists Spark applications.
func (c *Collector) ListApps(ctx context.Context, namespace string) ([]models.SparkAppSummary, error) {
	return c.spark.ListApplications(ctx, namespace)
}

// GetApplication returns application details.
func (c *Collector) GetApplication(ctx context.Context, namespace, name string) (*models.SparkApplicationDetail, error) {
	return c.spark.GetApplication(ctx, namespace, name)
}

// DriverLogs collects driver pod logs.
func (c *Collector) DriverLogs(ctx context.Context, namespace, appName string, opts models.LogOptions) (*models.PodLogs, error) {
	driver, _, err := c.k8s.FindSparkPods(ctx, namespace, appName)
	if err != nil {
		return nil, err
	}
	if driver == nil {
		app, err := c.spark.GetApplication(ctx, namespace, appName)
		if err != nil {
			return nil, fmt.Errorf("driver pod not found for %s", appName)
		}
		if app.Driver.Name == "" {
			return nil, fmt.Errorf("driver pod not found for %s", appName)
		}
		logs, err := c.k8s.GetPodLogs(ctx, namespace, app.Driver.Name, opts)
		if err != nil {
			return nil, err
		}
		return &models.PodLogs{PodName: app.Driver.Name, Logs: logs, Previous: opts.Previous}, nil
	}
	logs, err := c.k8s.GetPodLogs(ctx, namespace, driver.Name, opts)
	if err != nil {
		return nil, err
	}
	return &models.PodLogs{PodName: driver.Name, Logs: logs, Previous: opts.Previous}, nil
}

// ExecutorLogs collects logs from all executor pods.
func (c *Collector) ExecutorLogs(ctx context.Context, namespace, appName string, opts models.LogOptions) (*models.ExecutorLogs, error) {
	_, executors, err := c.k8s.FindSparkPods(ctx, namespace, appName)
	if err != nil {
		return nil, err
	}

	result := &models.ExecutorLogs{
		Application: appName,
		Namespace:   namespace,
		Logs:        make([]models.PodLogs, 0),
	}

	for _, exec := range executors {
		if opts.Executor != "" && exec.Name != opts.Executor {
			continue
		}
		logs, err := c.k8s.GetPodLogs(ctx, namespace, exec.Name, opts)
		if err != nil {
			result.Logs = append(result.Logs, models.PodLogs{
				PodName: exec.Name,
				Logs:    fmt.Sprintf("Error fetching logs: %v", err),
			})
			continue
		}
		result.Logs = append(result.Logs, models.PodLogs{
			PodName:  exec.Name,
			Logs:     logs,
			Previous: opts.Previous,
		})
	}
	return result, nil
}

// DescribeDriver returns kubectl describe output for the driver pod.
func (c *Collector) DescribeDriver(ctx context.Context, namespace, appName string) (*models.PodDescription, error) {
	driver, _, err := c.k8s.FindSparkPods(ctx, namespace, appName)
	if err != nil {
		return nil, err
	}
	podName := ""
	if driver != nil {
		podName = driver.Name
	} else {
		app, err := c.spark.GetApplication(ctx, namespace, appName)
		if err != nil {
			return nil, err
		}
		podName = app.Driver.Name
	}
	desc, err := c.k8s.DescribePod(ctx, namespace, podName)
	if err != nil {
		return nil, err
	}
	return &models.PodDescription{PodName: podName, Namespace: namespace, Description: desc}, nil
}

// DescribeExecutor returns kubectl describe output for an executor pod.
func (c *Collector) DescribeExecutor(ctx context.Context, namespace, appName, executorName string) (*models.PodDescription, error) {
	if executorName == "" {
		_, executors, err := c.k8s.FindSparkPods(ctx, namespace, appName)
		if err != nil {
			return nil, err
		}
		if len(executors) == 0 {
			return nil, fmt.Errorf("no executors found for %s", appName)
		}
		executorName = executors[0].Name
	}
	desc, err := c.k8s.DescribePod(ctx, namespace, executorName)
	if err != nil {
		return nil, err
	}
	return &models.PodDescription{PodName: executorName, Namespace: namespace, Description: desc}, nil
}

// GetEvents collects filtered namespace events.
func (c *Collector) GetEvents(ctx context.Context, namespace string) ([]models.EventSummary, error) {
	return c.k8s.GetEvents(ctx, namespace, true)
}

// SparkConf collects Spark configuration.
func (c *Collector) SparkConf(ctx context.Context, namespace, appName string) (*models.SparkConf, error) {
	return c.spark.CollectSparkConf(ctx, namespace, appName)
}

// PodResources collects resource info for driver and executors.
func (c *Collector) PodResources(ctx context.Context, namespace, appName string) ([]models.PodResources, error) {
	driver, executors, err := c.k8s.FindSparkPods(ctx, namespace, appName)
	if err != nil {
		return nil, err
	}

	var resources []models.PodResources
	if driver != nil {
		res, err := c.k8s.PodResources(ctx, namespace, driver.Name)
		if err == nil {
			resources = append(resources, *res)
		}
	}
	for _, exec := range executors {
		res, err := c.k8s.PodResources(ctx, namespace, exec.Name)
		if err == nil {
			resources = append(resources, *res)
		}
	}
	return resources, nil
}

// CollectAll gathers all diagnostic data for analysis.
func (c *Collector) CollectAll(ctx context.Context, namespace, appName string) (*models.AnalysisContext, error) {
	app, err := c.spark.GetApplication(ctx, namespace, appName)
	if err != nil {
		return nil, err
	}

	driverLogs, _ := c.DriverLogs(ctx, namespace, appName, models.LogOptions{Tail: 5000})
	execLogs, _ := c.ExecutorLogs(ctx, namespace, appName, models.LogOptions{Tail: 2000})
	events, _ := c.GetEvents(ctx, namespace)
	sparkConf, _ := c.SparkConf(ctx, namespace, appName)
	driverDesc, _ := c.DescribeDriver(ctx, namespace, appName)

	execDescs := make(map[string]string)
	if execLogs != nil {
		for _, el := range execLogs.Logs {
			if desc, err := c.k8s.DescribePod(ctx, namespace, el.PodName); err == nil {
				execDescs[el.PodName] = desc
			}
		}
	}

	podRes, _ := c.PodResources(ctx, namespace, appName)
	var driverRes models.PodResources
	var execRes []models.PodResources
	for _, r := range podRes {
		if r.PodName == app.Driver.Name {
			driverRes = r
		} else {
			execRes = append(execRes, r)
		}
	}

	driverLogStr := ""
	if driverLogs != nil {
		driverLogStr = driverLogs.Logs
	}
	var execLogList []models.PodLogs
	if execLogs != nil {
		execLogList = execLogs.Logs
	}
	driverDescStr := ""
	if driverDesc != nil {
		driverDescStr = driverDesc.Description
	}
	sparkConfVal := models.SparkConf{}
	if sparkConf != nil {
		sparkConfVal = *sparkConf
	}

	return &models.AnalysisContext{
		Application:   *app,
		DriverLogs:    driverLogStr,
		ExecutorLogs:  execLogList,
		Events:        events,
		SparkConf:     sparkConfVal,
		DriverDesc:    driverDescStr,
		ExecutorDescs: execDescs,
		DriverRes:     driverRes,
		ExecutorRes:   execRes,
	}, nil
}
