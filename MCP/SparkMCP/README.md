# Spark Debug MCP

Production-grade MCP server for diagnosing Apache Spark applications running on Kubernetes.

Connect Claude Desktop, Cursor, or VS Code to automatically investigate Spark failures — collect logs, events, configuration, and produce root cause analysis with remediation suggestions.

## Features

- **10 MCP tools** for Spark/Kubernetes diagnostics
- **45+ rule engine** for automated failure detection
- **Spark Operator** and standalone Spark on Kubernetes support
- **Multi-format output**: JSON, Markdown, console
- **AI-ready reports** with confidence scores, timelines, and recommendations
- Works with Minikube, Kind, EKS, GKE, and any Kubernetes cluster

## Architecture

See [docs/architecture.md](docs/architecture.md) for component diagrams and data flow.

```
LLM Client → MCP Server → Diagnostics Collector → Kubernetes / Spark Operator
                              ↓
                         Rule Engine → Failure Report
```

## Prerequisites

- Go 1.24+
- Kubernetes cluster with Spark Operator (optional for standalone Spark)
- Valid kubeconfig (`~/.kube/config`) or in-cluster service account

## Installation

```bash
git clone https://github.com/spark-debug-mcp/spark-debug-mcp.git
cd spark-debug-mcp
make build
```

Binary output: `bin/spark-debug-mcp`

## Building

```bash
# Build binary
make build

# Run tests (80%+ coverage on library packages)
make test

# Format and lint
make fmt
make lint

# Full build pipeline
./scripts/build.sh
```

## Running

### Local (stdio MCP transport)

```bash
./bin/spark-debug-mcp serve

# With options
./bin/spark-debug-mcp serve \
  --kubeconfig ~/.kube/config \
  --namespace spark \
  --log-level info \
  --output-format json
```

### Docker

```bash
make docker
docker run -i --rm \
  -v ~/.kube/config:/home/spark/.kube/config:ro \
  spark-debug-mcp:1.0.0 serve
```

### Docker Compose

```bash
docker compose up
```

Mounts kubeconfig and config directory automatically.

## Configuration

Configuration sources (in priority order):

1. CLI flags
2. Environment variables (`SPARK_DEBUG_*`, `KUBECONFIG`)
3. Config file (`~/.spark-debug.yaml`)

Example `~/.spark-debug.yaml`:

```yaml
namespace: spark
logLevel: info
outputFormat: json
inCluster: false
kubeconfig: ~/.kube/config
retryMax: 3
retryDelay: 500ms
```

## Connecting Claude Desktop

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "spark-debug": {
      "command": "/path/to/spark-debug-mcp",
      "args": ["serve", "--namespace", "spark"],
      "env": {
        "KUBECONFIG": "/Users/you/.kube/config"
      }
    }
  }
}
```

## Connecting Cursor

Add to Cursor MCP settings (`.cursor/mcp.json`):

```json
{
  "mcpServers": {
    "spark-debug": {
      "command": "/path/to/bin/spark-debug-mcp",
      "args": ["serve"],
      "env": {
        "KUBECONFIG": "/home/you/.kube/config",
        "SPARK_DEBUG_NAMESPACE": "spark"
      }
    }
  }
}
```

## Connecting VS Code

Use an MCP-compatible extension and configure:

```json
{
  "spark-debug-mcp": {
    "command": "spark-debug-mcp",
    "args": ["serve", "--kubeconfig", "${env:HOME}/.kube/config"]
  }
}
```

## MCP Tools

| Tool | Description |
|------|-------------|
| `list_spark_apps` | List Spark applications with state, age, driver, executors |
| `get_application` | Full SparkApplication details, YAML, status, restart count |
| `driver_logs` | Driver pod logs (tail, since, previous) |
| `executor_logs` | Executor logs with optional filtering |
| `describe_driver` | kubectl describe for driver pod |
| `describe_executor` | kubectl describe for executor pod |
| `get_events` | Filtered Kubernetes events (Warning, OOMKilled, etc.) |
| `spark_conf` | Spark config from CR, ConfigMaps, env vars |
| `pod_resources` | CPU, memory, limits, QoS, restart count |
| `analyze_failure` | Full automated root cause analysis |

## Example Prompts

```
Why did Spark application pi-calculator fail in the spark namespace?

Analyze the failure of my-spark-job and suggest remediation.

List all failed Spark applications in the spark namespace.

Get driver logs for app xyz and check for OOM errors.
```

## Supported Rules (45+)

| ID | Rule | Severity |
|----|------|----------|
| R001 | Executor OOM | Critical |
| R002 | Driver OOM | Critical |
| R003 | Node Pressure | High |
| R004 | PVC Pending | High |
| R005 | Image Pull Error | Critical |
| R006 | DNS Failure | High |
| R007 | Network Timeout | High |
| R008 | Executor Lost | Critical |
| R009 | Shuffle Fetch Failure | High |
| R010 | Disk Full | Critical |
| R011 | Memory Overhead Low | Medium |
| R012 | Container Restart Loop | High |
| R013 | Java Heap Space | Critical |
| R014 | GC Overhead | High |
| R015 | OutOfDirectMemory | High |
| R016 | Permission Denied | High |
| R017 | AWS Credential Failure | Critical |
| R018 | S3 Authentication | High |
| R019 | Missing ConfigMap | High |
| R020 | Missing Secret | High |
| R021 | RBAC Denied | High |
| R022 | ServiceAccount Missing | High |
| R023 | Webhook Failure | High |
| R024 | Admission Rejected | High |
| R025 | Insufficient CPU | High |
| R026 | Insufficient Memory | High |
| R027 | Node Tainted | Medium |
| R028 | ImagePullBackOff | Critical |
| R029 | CrashLoopBackOff | Critical |
| R030 | Executor Timeout | High |
| R031 | Shuffle Service Failure | Medium |
| R032 | Dynamic Allocation Failure | Medium |
| R033 | Speculation Issue | Medium |
| R034 | OOM Score High | Medium |
| R035 | ContainerKilled | High |
| R036 | SparkConf Invalid | Medium |
| R037 | Spark Version Mismatch | Medium |
| R038 | Classpath Issue | High |
| R039 | Jar Missing | Critical |
| R040 | Dependency Resolution Failure | High |
| R041 | Unsupported Hadoop Version | Medium |
| R042 | DeadlineExceeded | High |
| R043 | BackOff | Medium |
| R044 | Evicted | High |
| R045 | FailedScheduling | High |

Custom rules can be registered via `rules.Engine.Register()` without modifying core logic.

## Examples

- [SparkApplication sample](examples/sparkapplication.yaml)
- [Sample OOM driver log](examples/sample-driver-oom.log)
- [Sample analysis output](examples/sample-analysis-output.json)

## Screenshots

<!-- Placeholder: Claude Desktop showing analyze_failure output -->
![Analysis Report Placeholder](docs/screenshots/analysis-report.png)

<!-- Placeholder: Cursor MCP tools list -->
![MCP Tools Placeholder](docs/screenshots/mcp-tools.png)

## Roadmap

- [ ] SSE/HTTP MCP transport for remote deployment
- [ ] Prometheus metrics integration
- [ ] Spark History Server log correlation
- [ ] Multi-cluster support
- [ ] Slack/PagerDuty alert integration
- [ ] GPU/OOM profiling for ML workloads

## License

Apache License 2.0 — see [LICENSE](LICENSE).
