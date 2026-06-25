// Command spark-debug-mcp starts the Spark Debugging MCP Server.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spark-debug-mcp/spark-debug-mcp/internal/analyzer"
	"github.com/spark-debug-mcp/spark-debug-mcp/internal/config"
	"github.com/spark-debug-mcp/spark-debug-mcp/internal/diagnostics"
	"github.com/spark-debug-mcp/spark-debug-mcp/internal/kubernetes"
	mcpserver "github.com/spark-debug-mcp/spark-debug-mcp/internal/mcpserver"
	"github.com/spark-debug-mcp/spark-debug-mcp/internal/spark"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	version = "1.0.0"
	cfgFile string
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spark-debug-mcp",
		Short: "Spark Debugging MCP Server for Kubernetes",
		Long:  "MCP server that diagnoses Spark applications running on Kubernetes using Spark Operator.",
	}

	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default ~/.spark-debug.yaml)")
	cmd.PersistentFlags().String("kubeconfig", "", "path to kubeconfig file")
	cmd.PersistentFlags().Bool("in-cluster", false, "use in-cluster Kubernetes config")
	cmd.PersistentFlags().String("namespace", "default", "default Kubernetes namespace")
	cmd.PersistentFlags().String("log-level", "info", "log level (debug, info, warn, error)")
	cmd.PersistentFlags().String("output-format", "json", "default output format (json, markdown, console)")

	viper.BindPFlag("kubeconfig", cmd.PersistentFlags().Lookup("kubeconfig"))
	viper.BindPFlag("inCluster", cmd.PersistentFlags().Lookup("in-cluster"))
	viper.BindPFlag("namespace", cmd.PersistentFlags().Lookup("namespace"))
	viper.BindPFlag("logLevel", cmd.PersistentFlags().Lookup("log-level"))
	viper.BindPFlag("outputFormat", cmd.PersistentFlags().Lookup("output-format"))

	cmd.AddCommand(newServeCmd())
	cmd.AddCommand(newVersionCmd())

	return cmd
}

func newServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the MCP server (stdio transport)",
		RunE:  runServe,
	}
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "spark-debug-mcp version %s\n", version)
		},
	}
}

func runServe(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	cfg.ServerVersion = version

	logger, err := initLogger(cfg.LogLevel)
	if err != nil {
		return err
	}
	defer logger.Sync()

	k8sClient, err := kubernetes.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("create kubernetes client: %w", err)
	}

	if err := k8sClient.Ping(context.Background()); err != nil {
		logger.Warn("kubernetes API ping failed", zap.Error(err))
	}

	sparkSvc := spark.NewService(k8sClient)
	collector := diagnostics.NewCollector(k8sClient, sparkSvc)
	anal := analyzer.NewAnalyzer(collector)

	server := mcpserver.NewServer(cfg, collector, anal, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		logger.Info("shutdown signal received")
		cancel()
	}()

	return server.Run(ctx)
}

func initLogger(level string) (*zap.Logger, error) {
	var cfg zap.Config
	switch level {
	case "debug":
		cfg = zap.NewDevelopmentConfig()
	default:
		cfg = zap.NewProductionConfig()
	}
	cfg.OutputPaths = []string{"stderr"}
	return cfg.Build()
}
