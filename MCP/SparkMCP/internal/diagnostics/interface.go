// Package diagnostics provides data collection for Spark debugging.
package diagnostics

import (
	"context"

	"github.com/spark-debug-mcp/spark-debug-mcp/internal/models"
)

// CollectorInterface defines diagnostic collection operations.
type CollectorInterface interface {
	ListApps(ctx context.Context, namespace string) ([]models.SparkAppSummary, error)
	GetApplication(ctx context.Context, namespace, name string) (*models.SparkApplicationDetail, error)
	DriverLogs(ctx context.Context, namespace, appName string, opts models.LogOptions) (*models.PodLogs, error)
	ExecutorLogs(ctx context.Context, namespace, appName string, opts models.LogOptions) (*models.ExecutorLogs, error)
	DescribeDriver(ctx context.Context, namespace, appName string) (*models.PodDescription, error)
	DescribeExecutor(ctx context.Context, namespace, appName, executorName string) (*models.PodDescription, error)
	GetEvents(ctx context.Context, namespace string) ([]models.EventSummary, error)
	SparkConf(ctx context.Context, namespace, appName string) (*models.SparkConf, error)
	PodResources(ctx context.Context, namespace, appName string) ([]models.PodResources, error)
	CollectAll(ctx context.Context, namespace, appName string) (*models.AnalysisContext, error)
}

// Ensure Collector implements CollectorInterface.
var _ CollectorInterface = (*Collector)(nil)
