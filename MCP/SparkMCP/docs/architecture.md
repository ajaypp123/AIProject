# Spark Debug MCP - Architecture

## Overview

Spark Debug MCP is a Model Context Protocol server that enables LLMs to diagnose
Spark applications running on Kubernetes. It integrates with Spark Operator and
supports standalone Spark deployments.

## Components

```
┌─────────────────────────────────────────────────────────────┐
│                     MCP Client (LLM)                        │
│              Claude Desktop / Cursor / VS Code              │
└─────────────────────────┬───────────────────────────────────┘
                          │ stdio (JSON-RPC)
┌─────────────────────────▼───────────────────────────────────┐
│                    MCP Server Layer                          │
│  list_spark_apps │ driver_logs │ analyze_failure │ ...       │
└─────────────────────────┬───────────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────────┐
│                  Diagnostics Collector                       │
│  Logs │ Events │ Pod Describe │ Spark Conf │ Resources      │
└──────────┬──────────────────────────────┬───────────────────┘
           │                              │
┌──────────▼──────────┐      ┌────────────▼──────────────────┐
│  Kubernetes Client  │      │      Spark Service            │
│  client-go          │      │  SparkApplication CR (dynamic)│
└──────────┬──────────┘      └────────────┬──────────────────┘
           │                              │
┌──────────▼──────────────────────────────▼───────────────────┐
│                   Kubernetes Cluster                         │
│  Spark Operator │ Driver Pods │ Executor Pods │ Events      │
└─────────────────────────────────────────────────────────────┘
```

## Rule Engine

The analyzer uses a pluggable rule engine with 45+ diagnostic rules. Each rule:

- Has a unique ID, name, description, and severity
- Implements detection logic against an `AnalysisContext`
- Returns evidence and remediation recommendations
- Can be extended without modifying core logic via `Engine.Register()`

## Data Flow for `analyze_failure`

1. Collect SparkApplication CR and pod metadata
2. Fetch driver and executor logs
3. Collect filtered Kubernetes events
4. Gather Spark configuration from CR, ConfigMaps, env vars
5. Parse logs for exceptions and stack traces
6. Run rule engine against collected context
7. Build timeline, compute scores, generate report
8. Return JSON, Markdown, or console formatted output

## Authentication

- **KubeConfig**: Default `~/.kube/config` or `KUBECONFIG` env var
- **InCluster**: Service account token when running inside Kubernetes
- **Custom**: `--kubeconfig` flag or config file

## Supported Clusters

- Minikube, Kind (local development)
- EKS, GKE (cloud managed)
- Any Kubernetes cluster with Spark Operator
