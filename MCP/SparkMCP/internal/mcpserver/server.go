// Package mcpserver implements the MCP server and tool handlers.
package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spark-debug-mcp/spark-debug-mcp/internal/analyzer"
	"github.com/spark-debug-mcp/spark-debug-mcp/internal/config"
	"github.com/spark-debug-mcp/spark-debug-mcp/internal/diagnostics"
	"github.com/spark-debug-mcp/spark-debug-mcp/internal/models"
	"github.com/spark-debug-mcp/spark-debug-mcp/pkg/output"
	"go.uber.org/zap"
)

// Server wraps the MCP server and dependencies.
type Server struct {
	mcpServer *mcp.Server
	collector diagnostics.CollectorInterface
	analyzer  analyzer.AnalyzerInterface
	cfg       *config.Config
	logger    *zap.Logger
}

// NewServer creates and configures the MCP server.
func NewServer(cfg *config.Config, collector diagnostics.CollectorInterface, anal analyzer.AnalyzerInterface, logger *zap.Logger) *Server {
	s := &Server{
		mcpServer: mcp.NewServer(&mcp.Implementation{
			Name:    cfg.ServerName,
			Version: cfg.ServerVersion,
		}, nil),
		collector: collector,
		analyzer:  anal,
		cfg:       cfg,
		logger:    logger,
	}
	s.registerTools()
	return s
}

// Run starts the MCP server on stdio transport.
func (s *Server) Run(ctx context.Context) error {
	s.logger.Info("starting MCP server", zap.String("name", s.cfg.ServerName))
	return s.mcpServer.Run(ctx, &mcp.StdioTransport{})
}

func (s *Server) registerTools() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "list_spark_apps",
		Description: "List Spark applications in a Kubernetes namespace with state, age, driver, and executor count.",
	}, s.HandleListSparkApps)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_application",
		Description: "Get detailed Spark application info including YAML, status, driver, executors, and restart count.",
	}, s.HandleGetApplication)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "driver_logs",
		Description: "Download driver pod logs. Supports previous logs, tail, and since duration.",
	}, s.HandleDriverLogs)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "executor_logs",
		Description: "Download logs from executor pods. Supports filtering by executor name and previous logs.",
	}, s.HandleExecutorLogs)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "describe_driver",
		Description: "Get kubectl describe output for the Spark driver pod.",
	}, s.HandleDescribeDriver)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "describe_executor",
		Description: "Get kubectl describe output for a Spark executor pod.",
	}, s.HandleDescribeExecutor)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_events",
		Description: "Collect filtered Kubernetes events (Warning, FailedScheduling, OOMKilled, etc.).",
	}, s.HandleGetEvents)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "spark_conf",
		Description: "Collect Spark configuration from SparkApplication CR, ConfigMaps, env vars, and mounted files.",
	}, s.HandleSparkConf)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "pod_resources",
		Description: "Collect CPU, memory, limits, requests, restart count, node, and QoS for Spark pods.",
	}, s.HandlePodResources)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "analyze_failure",
		Description: "Comprehensive failure analysis: collects all diagnostics, runs rule engine, produces root cause report.",
	}, s.HandleAnalyzeFailure)
}

// --- Input types ---

type NamespaceInput struct {
	Namespace string `json:"namespace" jsonschema:"Kubernetes namespace"`
}

type AppInput struct {
	Namespace   string `json:"namespace" jsonschema:"Kubernetes namespace"`
	Application string `json:"application" jsonschema:"Spark application name"`
}

type DriverLogsInput struct {
	Namespace   string `json:"namespace" jsonschema:"Kubernetes namespace"`
	Application string `json:"application" jsonschema:"Spark application name"`
	Previous    bool   `json:"previous,omitempty" jsonschema:"Fetch previous container logs"`
	Tail        int64  `json:"tail,omitempty" jsonschema:"Number of lines from end of logs"`
	Since       string `json:"since,omitempty" jsonschema:"Duration e.g. 1h, 30m"`
}

type ExecutorLogsInput struct {
	Namespace   string `json:"namespace" jsonschema:"Kubernetes namespace"`
	Application string `json:"application" jsonschema:"Spark application name"`
	Executor    string `json:"executor,omitempty" jsonschema:"Filter by executor pod name"`
	Previous    bool   `json:"previous,omitempty" jsonschema:"Fetch previous container logs"`
	Tail        int64  `json:"tail,omitempty" jsonschema:"Number of lines from end of logs"`
	Since       string `json:"since,omitempty" jsonschema:"Duration e.g. 1h, 30m"`
}

type DescribeExecutorInput struct {
	Namespace   string `json:"namespace" jsonschema:"Kubernetes namespace"`
	Application string `json:"application" jsonschema:"Spark application name"`
	Executor    string `json:"executor,omitempty" jsonschema:"Executor pod name (defaults to first executor)"`
}

type AnalyzeInput struct {
	Namespace   string `json:"namespace" jsonschema:"Kubernetes namespace"`
	Application string `json:"application" jsonschema:"Spark application name"`
	Format      string `json:"format,omitempty" jsonschema:"Output format: json, markdown, or console"`
}

// ToolOutput is the standard MCP tool response payload.
type ToolOutput struct {
	Result string `json:"result" jsonschema:"JSON-formatted tool result"`
}

// --- Handlers ---

func (s *Server) HandleListSparkApps(ctx context.Context, _ *mcp.CallToolRequest, input NamespaceInput) (*mcp.CallToolResult, ToolOutput, error) {
	ns := s.resolveNamespace(input.Namespace)
	apps, err := s.collector.ListApps(ctx, ns)
	if err != nil {
		return errorResult(err), ToolOutput{}, nil
	}
	text, err := output.FormatAny(apps)
	if err != nil {
		return errorResult(err), ToolOutput{}, nil
	}
	return nil, ToolOutput{Result: text}, nil
}

func (s *Server) HandleGetApplication(ctx context.Context, _ *mcp.CallToolRequest, input AppInput) (*mcp.CallToolResult, ToolOutput, error) {
	ns := s.resolveNamespace(input.Namespace)
	app, err := s.collector.GetApplication(ctx, ns, input.Application)
	if err != nil {
		return errorResult(err), ToolOutput{}, nil
	}
	text, err := output.FormatAny(app)
	if err != nil {
		return errorResult(err), ToolOutput{}, nil
	}
	return nil, ToolOutput{Result: text}, nil
}

func (s *Server) HandleDriverLogs(ctx context.Context, _ *mcp.CallToolRequest, input DriverLogsInput) (*mcp.CallToolResult, ToolOutput, error) {
	ns := s.resolveNamespace(input.Namespace)
	logs, err := s.collector.DriverLogs(ctx, ns, input.Application, models.LogOptions{
		Previous: input.Previous,
		Tail:     input.Tail,
		Since:    input.Since,
	})
	if err != nil {
		return errorResult(err), ToolOutput{}, nil
	}
	text, err := output.FormatAny(logs)
	if err != nil {
		return errorResult(err), ToolOutput{}, nil
	}
	return nil, ToolOutput{Result: text}, nil
}

func (s *Server) HandleExecutorLogs(ctx context.Context, _ *mcp.CallToolRequest, input ExecutorLogsInput) (*mcp.CallToolResult, ToolOutput, error) {
	ns := s.resolveNamespace(input.Namespace)
	logs, err := s.collector.ExecutorLogs(ctx, ns, input.Application, models.LogOptions{
		Previous: input.Previous,
		Tail:     input.Tail,
		Since:    input.Since,
		Executor: input.Executor,
	})
	if err != nil {
		return errorResult(err), ToolOutput{}, nil
	}
	text, err := output.FormatAny(logs)
	if err != nil {
		return errorResult(err), ToolOutput{}, nil
	}
	return nil, ToolOutput{Result: text}, nil
}

func (s *Server) HandleDescribeDriver(ctx context.Context, _ *mcp.CallToolRequest, input AppInput) (*mcp.CallToolResult, ToolOutput, error) {
	ns := s.resolveNamespace(input.Namespace)
	desc, err := s.collector.DescribeDriver(ctx, ns, input.Application)
	if err != nil {
		return errorResult(err), ToolOutput{}, nil
	}
	text, err := output.FormatAny(desc)
	if err != nil {
		return errorResult(err), ToolOutput{}, nil
	}
	return nil, ToolOutput{Result: text}, nil
}

func (s *Server) HandleDescribeExecutor(ctx context.Context, _ *mcp.CallToolRequest, input DescribeExecutorInput) (*mcp.CallToolResult, ToolOutput, error) {
	ns := s.resolveNamespace(input.Namespace)
	desc, err := s.collector.DescribeExecutor(ctx, ns, input.Application, input.Executor)
	if err != nil {
		return errorResult(err), ToolOutput{}, nil
	}
	text, err := output.FormatAny(desc)
	if err != nil {
		return errorResult(err), ToolOutput{}, nil
	}
	return nil, ToolOutput{Result: text}, nil
}

func (s *Server) HandleGetEvents(ctx context.Context, _ *mcp.CallToolRequest, input NamespaceInput) (*mcp.CallToolResult, ToolOutput, error) {
	ns := s.resolveNamespace(input.Namespace)
	events, err := s.collector.GetEvents(ctx, ns)
	if err != nil {
		return errorResult(err), ToolOutput{}, nil
	}
	text, err := output.FormatAny(events)
	if err != nil {
		return errorResult(err), ToolOutput{}, nil
	}
	return nil, ToolOutput{Result: text}, nil
}

func (s *Server) HandleSparkConf(ctx context.Context, _ *mcp.CallToolRequest, input AppInput) (*mcp.CallToolResult, ToolOutput, error) {
	ns := s.resolveNamespace(input.Namespace)
	conf, err := s.collector.SparkConf(ctx, ns, input.Application)
	if err != nil {
		return errorResult(err), ToolOutput{}, nil
	}
	text, err := output.FormatAny(conf)
	if err != nil {
		return errorResult(err), ToolOutput{}, nil
	}
	return nil, ToolOutput{Result: text}, nil
}

func (s *Server) HandlePodResources(ctx context.Context, _ *mcp.CallToolRequest, input AppInput) (*mcp.CallToolResult, ToolOutput, error) {
	ns := s.resolveNamespace(input.Namespace)
	resources, err := s.collector.PodResources(ctx, ns, input.Application)
	if err != nil {
		return errorResult(err), ToolOutput{}, nil
	}
	text, err := output.FormatAny(resources)
	if err != nil {
		return errorResult(err), ToolOutput{}, nil
	}
	return nil, ToolOutput{Result: text}, nil
}

func (s *Server) HandleAnalyzeFailure(ctx context.Context, _ *mcp.CallToolRequest, input AnalyzeInput) (*mcp.CallToolResult, ToolOutput, error) {
	ns := s.resolveNamespace(input.Namespace)
	report, err := s.analyzer.AnalyzeFailure(ctx, ns, input.Application)
	if err != nil {
		return errorResult(err), ToolOutput{}, nil
	}

	format := models.OutputFormat(input.Format)
	if format == "" {
		format = models.OutputFormat(s.cfg.OutputFormat)
	}

	text, err := output.Format(report, format)
	if err != nil {
		text, _ = output.FormatJSON(report)
	}
	return nil, ToolOutput{Result: text}, nil
}

func (s *Server) resolveNamespace(ns string) string {
	if ns == "" {
		return s.cfg.Namespace
	}
	return ns
}

func errorResult(err error) *mcp.CallToolResult {
	msg, _ := json.Marshal(map[string]string{"error": err.Error()})
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(msg)},
		},
		IsError: true,
	}
}

// TextResult creates a text tool result (used in tests).
func TextResult(text string) (*mcp.CallToolResult, ToolOutput, error) {
	return nil, ToolOutput{Result: text}, nil
}

// ErrorText returns formatted error string.
func ErrorText(err error) string {
	if err == nil {
		return `{"error": "unknown error"}`
	}
	return fmt.Sprintf(`{"error": %q}`, err.Error())
}
