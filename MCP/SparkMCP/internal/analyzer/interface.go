package analyzer

import (
	"context"

	"github.com/spark-debug-mcp/spark-debug-mcp/internal/models"
)

// AnalyzerInterface defines failure analysis operations.
type AnalyzerInterface interface {
	AnalyzeFailure(ctx context.Context, namespace, appName string) (*models.FailureReport, error)
}

// Ensure Analyzer implements AnalyzerInterface.
var _ AnalyzerInterface = (*Analyzer)(nil)
