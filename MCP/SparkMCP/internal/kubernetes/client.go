// Package kubernetes provides Kubernetes API client wrappers.
package kubernetes

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spark-debug-mcp/spark-debug-mcp/internal/config"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client wraps Kubernetes clientsets.
type Client struct {
	Clientset     kubernetes.Interface
	Dynamic       dynamic.Interface
	RESTConfig    *rest.Config
	config        *config.Config
}

// NewClient creates a Kubernetes client from configuration.
func NewClient(cfg *config.Config) (*Client, error) {
	restConfig, err := buildRESTConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("build rest config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("create clientset: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("create dynamic client: %w", err)
	}

	return &Client{
		Clientset:  clientset,
		Dynamic:    dynamicClient,
		RESTConfig: restConfig,
		config:     cfg,
	}, nil
}

func buildRESTConfig(cfg *config.Config) (*rest.Config, error) {
	if cfg.InCluster {
		return rest.InClusterConfig()
	}

	kubeconfig := cfg.KubeConfig
	if env := os.Getenv("KUBECONFIG"); env != "" {
		kubeconfig = env
	}

	if kubeconfig == "" {
		return nil, fmt.Errorf("kubeconfig not configured")
	}

	absPath, err := filepath.Abs(kubeconfig)
	if err != nil {
		return nil, err
	}

	return clientcmd.BuildConfigFromFlags("", absPath)
}

// Ping verifies connectivity to the Kubernetes API.
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.Clientset.Discovery().ServerVersion()
	return err
}
