package rules

import (
	"strconv"

	"github.com/spark-debug-mcp/spark-debug-mcp/internal/models"
)

// AllRules returns all registered diagnostic rules (40+).
func AllRules() []Rule {
	return []Rule{
		{
			ID: "R001", Name: "Executor OOM", Description: "Executor killed due to out-of-memory",
			Severity: models.SeverityCritical, Score: 90,
			Recommendation: "Increase executor memory (spark.executor.memory) and memory overhead (spark.executor.memoryOverhead). Reduce partition size or enable spill.",
			DocLinks:       []string{"https://spark.apache.org/docs/latest/configuration.html#memory-management"},
			Detect: func(ctx *models.AnalysisContext) []string {
				ev := eventReasons(ctx, "OOMKilled")
				ev = append(ev, logPatterns(ctx, "ExecutorLost", "Container killed by YARN/K8s", "exit code 137", "OutOfMemoryError")...)
				ev = append(ev, descPatterns(ctx, "OOMKilled", "Reason: OOMKilled")...)
				return ev
			},
		},
		{
			ID: "R002", Name: "Driver OOM", Description: "Driver pod killed due to out-of-memory",
			Severity: models.SeverityCritical, Score: 92,
			Recommendation: "Increase driver memory (spark.driver.memory). Reduce collect() operations and broadcast large variables.",
			DocLinks:       []string{"https://spark.apache.org/docs/latest/configuration.html#application-properties"},
			Detect: func(ctx *models.AnalysisContext) []string {
				if containsAny(ctx.DriverDesc, "OOMKilled") {
					return []string{"Driver OOMKilled detected in pod description"}
				}
				if containsAny(ctx.DriverLogs, "OutOfMemoryError", "Java heap space") && !containsAny(ctx.DriverLogs, "executor") {
					return []string{"Driver OutOfMemoryError in logs"}
				}
				return eventReasons(ctx, "OOMKilled")
			},
		},
		{
			ID: "R003", Name: "Node Pressure", Description: "Node experiencing resource pressure",
			Severity: models.SeverityHigh, Score: 75,
			Recommendation: "Check node capacity, add nodes, or reduce resource requests. Review pod resource limits.",
			DocLinks:       []string{"https://kubernetes.io/docs/concepts/scheduling-eviction/node-pressure-eviction/"},
			Detect: func(ctx *models.AnalysisContext) []string {
				return eventReasons(ctx, "NodePressure", "EvictionThresholdMet", "DiskPressure", "MemoryPressure")
			},
		},
		{
			ID: "R004", Name: "PVC Pending", Description: "PersistentVolumeClaim is pending",
			Severity: models.SeverityHigh, Score: 70,
			Recommendation: "Verify storage class exists, PVC size is available, and provisioner is running.",
			DocLinks:       []string{"https://kubernetes.io/docs/concepts/storage/persistent-volumes/"},
			Detect: func(ctx *models.AnalysisContext) []string {
				return append(eventReasons(ctx, "FailedScheduling", "ProvisioningFailed"),
					descPatterns(ctx, "PersistentVolumeClaim", "Pending")...)
			},
		},
		{
			ID: "R005", Name: "Image Pull Error", Description: "Failed to pull container image",
			Severity: models.SeverityCritical, Score: 85,
			Recommendation: "Verify image name, tag, registry credentials, and network connectivity to registry.",
			DocLinks:       []string{"https://kubernetes.io/docs/concepts/containers/images/"},
			Detect: func(ctx *models.AnalysisContext) []string {
				ev := eventReasons(ctx, "Failed", "ErrImagePull", "ImagePullBackOff")
				ev = append(ev, containerWaiting(ctx, "ImagePullBackOff")...)
				ev = append(ev, descPatterns(ctx, "ErrImagePull", "ImagePullBackOff")...)
				return ev
			},
		},
		{
			ID: "R006", Name: "DNS Failure", Description: "DNS resolution failure detected",
			Severity: models.SeverityHigh, Score: 72,
			Recommendation: "Check CoreDNS pods, network policies, and service DNS configuration.",
			DocLinks:       []string{"https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/"},
			Detect: func(ctx *models.AnalysisContext) []string {
				return logPatterns(ctx, "UnknownHostException", "Name or service not known", "DNS resolution failed", "no such host")
			},
		},
		{
			ID: "R007", Name: "Network Timeout", Description: "Network connection timeout",
			Severity: models.SeverityHigh, Score: 68,
			Recommendation: "Check network policies, firewall rules, and service endpoints. Increase timeout settings.",
			Detect: func(ctx *models.AnalysisContext) []string {
				return logPatterns(ctx, "SocketTimeoutException", "Connection timed out", "ConnectTimeoutException", "network timeout")
			},
		},
		{
			ID: "R008", Name: "Executor Lost", Description: "Spark lost connection to executor",
			Severity: models.SeverityCritical, Score: 88,
			Recommendation: "Check executor pod stability, node health, and network. Review executor memory and GC settings.",
			DocLinks:       []string{"https://spark.apache.org/docs/latest/tuning.html"},
			Detect: func(ctx *models.AnalysisContext) []string {
				return logPatterns(ctx, "Lost executor", "ExecutorLostFailure", "ExecutorLost", "removed executor")
			},
		},
		{
			ID: "R009", Name: "Shuffle Fetch Failure", Description: "Failed to fetch shuffle blocks",
			Severity: models.SeverityHigh, Score: 80,
			Recommendation: "Increase spark.shuffle.io.maxRetries and spark.network.timeout. Check executor stability.",
			DocLinks:       []string{"https://spark.apache.org/docs/latest/configuration.html#shuffle-behavior"},
			Detect: func(ctx *models.AnalysisContext) []string {
				return logPatterns(ctx, "FetchFailedException", "shuffle fetch failed", "Missing shuffle output")
			},
		},
		{
			ID: "R010", Name: "Disk Full", Description: "Insufficient disk space on node or pod",
			Severity: models.SeverityCritical, Score: 82,
			Recommendation: "Increase ephemeral storage limits, clean up old data, or add storage capacity.",
			Detect: func(ctx *models.AnalysisContext) []string {
				return append(eventReasons(ctx, "Evicted", "DiskPressure"),
					logPatterns(ctx, "No space left on device", "disk full", "IOException: No space")...)
			},
		},
		{
			ID: "R011", Name: "Memory Overhead Low", Description: "Executor memory overhead may be insufficient",
			Severity: models.SeverityMedium, Score: 55,
			Recommendation: "Increase spark.executor.memoryOverhead (typically 10-20% of executor memory).",
			DocLinks:       []string{"https://spark.apache.org/docs/latest/configuration.html#memory-management"},
			Detect: func(ctx *models.AnalysisContext) []string {
				if overhead, ok := ctx.SparkConf.AllConf["spark.executor.memoryOverhead"]; ok {
					if overhead == "384m" || overhead == "384" {
						return []string{"Default memory overhead may be too low: " + overhead}
					}
				}
				return logPatterns(ctx, "Container killed", "memory overhead", "exceeding physical memory")
			},
		},
		{
			ID: "R012", Name: "Container Restart Loop", Description: "Container repeatedly restarting",
			Severity: models.SeverityHigh, Score: 78,
			Recommendation: "Check container logs for crash reason. Fix underlying issue before increasing restart limits.",
			Detect: func(ctx *models.AnalysisContext) []string {
				var ev []string
				if ctx.DriverRes.RestartCount > 2 {
					ev = append(ev, "Driver restart count: "+itoa(int(ctx.DriverRes.RestartCount)))
				}
				for _, r := range ctx.ExecutorRes {
					if r.RestartCount > 2 {
						ev = append(ev, r.PodName+" restart count: "+itoa(int(r.RestartCount)))
					}
				}
				ev = append(ev, containerWaiting(ctx, "CrashLoopBackOff")...)
				return ev
			},
		},
		{
			ID: "R013", Name: "Java Heap Space", Description: "Java heap space exhausted",
			Severity: models.SeverityCritical, Score: 87,
			Recommendation: "Increase JVM heap size via spark.executor.memory or spark.driver.memory.",
			Detect: func(ctx *models.AnalysisContext) []string {
				return logPatterns(ctx, "Java heap space", "java.lang.OutOfMemoryError: Java heap space")
			},
		},
		{
			ID: "R014", Name: "GC Overhead", Description: "GC overhead limit exceeded",
			Severity: models.SeverityHigh, Score: 76,
			Recommendation: "Increase memory, tune GC settings, or reduce data per partition.",
			Detect: func(ctx *models.AnalysisContext) []string {
				return logPatterns(ctx, "GC overhead limit exceeded", "OutOfMemoryError: GC overhead")
			},
		},
		{
			ID: "R015", Name: "OutOfDirectMemory", Description: "Direct buffer memory exhausted",
			Severity: models.SeverityHigh, Score: 74,
			Recommendation: "Increase spark.executor.memoryOverhead or reduce off-heap memory usage.",
			Detect: func(ctx *models.AnalysisContext) []string {
				return logPatterns(ctx, "OutOfDirectMemoryError", "Direct buffer memory")
			},
		},
		{
			ID: "R016", Name: "Permission Denied", Description: "File or resource access denied",
			Severity: models.SeverityHigh, Score: 70,
			Recommendation: "Check file permissions, RBAC roles, and service account permissions.",
			Detect: func(ctx *models.AnalysisContext) []string {
				return logPatterns(ctx, "Permission denied", "AccessDeniedException", "PermissionDenied")
			},
		},
		{
			ID: "R017", Name: "AWS Credential Failure", Description: "AWS credentials invalid or missing",
			Severity: models.SeverityCritical, Score: 83,
			Recommendation: "Verify IAM roles, IRSA configuration, or AWS credential secrets.",
			DocLinks:       []string{"https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html"},
			Detect: func(ctx *models.AnalysisContext) []string {
				return logPatterns(ctx, "Unable to load AWS credentials", "InvalidAccessKeyId", "SignatureDoesNotMatch", "AmazonClientException")
			},
		},
		{
			ID: "R018", Name: "S3 Authentication", Description: "S3 authentication or authorization failure",
			Severity: models.SeverityHigh, Score: 79,
			Recommendation: "Verify S3 bucket permissions, IAM policies, and endpoint configuration.",
			Detect: func(ctx *models.AnalysisContext) []string {
				return logPatterns(ctx, "403 Forbidden", "S3Exception", "Access Denied", "s3a://")
			},
		},
		{
			ID: "R019", Name: "Missing ConfigMap", Description: "Required ConfigMap not found",
			Severity: models.SeverityHigh, Score: 73,
			Recommendation: "Create the missing ConfigMap or update pod spec to reference correct name.",
			Detect: func(ctx *models.AnalysisContext) []string {
				return append(eventReasons(ctx, "FailedMount"),
					descPatterns(ctx, "configmap", "not found")...)
			},
		},
		{
			ID: "R020", Name: "Missing Secret", Description: "Required Secret not found",
			Severity: models.SeverityHigh, Score: 73,
			Recommendation: "Create the missing Secret or verify secret name in pod spec.",
			Detect: func(ctx *models.AnalysisContext) []string {
				return append(eventReasons(ctx, "FailedMount"),
					descPatterns(ctx, "secret", "not found")...)
			},
		},
		{
			ID: "R021", Name: "RBAC Denied", Description: "Kubernetes RBAC permission denied",
			Severity: models.SeverityHigh, Score: 71,
			Recommendation: "Grant required RBAC permissions to the Spark service account.",
			DocLinks:       []string{"https://kubernetes.io/docs/reference/access-authn-authz/rbac/"},
			Detect: func(ctx *models.AnalysisContext) []string {
				return logPatterns(ctx, "Forbidden", "is forbidden", "RBAC", "cannot create resource")
			},
		},
		{
			ID: "R022", Name: "ServiceAccount Missing", Description: "ServiceAccount not found",
			Severity: models.SeverityHigh, Score: 72,
			Recommendation: "Create the ServiceAccount or update SparkApplication spec.",
			Detect: func(ctx *models.AnalysisContext) []string {
				return descPatterns(ctx, "serviceaccount", "not found")
			},
		},
		{
			ID: "R023", Name: "Webhook Failure", Description: "Admission webhook rejected or failed",
			Severity: models.SeverityHigh, Score: 77,
			Recommendation: "Check mutating/validating webhook configurations and webhook server health.",
			Detect: func(ctx *models.AnalysisContext) []string {
				return eventReasons(ctx, "FailedCreate", "webhook")
			},
		},
		{
			ID: "R024", Name: "Admission Rejected", Description: "Pod admission rejected by policy",
			Severity: models.SeverityHigh, Score: 75,
			Recommendation: "Review PodSecurityPolicy, OPA/Gatekeeper policies, and resource quotas.",
			Detect: func(ctx *models.AnalysisContext) []string {
				return append(eventReasons(ctx, "FailedCreate"),
					logPatterns(ctx, "admission webhook", "denied the request")...)
			},
		},
		{
			ID: "R025", Name: "Insufficient CPU", Description: "Cluster lacks sufficient CPU resources",
			Severity: models.SeverityHigh, Score: 74,
			Recommendation: "Reduce CPU requests, add nodes, or enable cluster autoscaling.",
			Detect: func(ctx *models.AnalysisContext) []string {
				return append(eventReasons(ctx, "FailedScheduling"),
					descPatterns(ctx, "Insufficient cpu")...)
			},
		},
		{
			ID: "R026", Name: "Insufficient Memory", Description: "Cluster lacks sufficient memory resources",
			Severity: models.SeverityHigh, Score: 74,
			Recommendation: "Reduce memory requests, add nodes, or optimize Spark memory settings.",
			Detect: func(ctx *models.AnalysisContext) []string {
				return append(eventReasons(ctx, "FailedScheduling"),
					descPatterns(ctx, "Insufficient memory")...)
			},
		},
		{
			ID: "R027", Name: "Node Tainted", Description: "Pod cannot schedule due to node taints",
			Severity: models.SeverityMedium, Score: 60,
			Recommendation: "Add tolerations to SparkApplication spec or remove/adjust node taints.",
			DocLinks:       []string{"https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/"},
			Detect: func(ctx *models.AnalysisContext) []string {
				return descPatterns(ctx, "didn't tolerate node taint", "untolerated taint")
			},
		},
		{
			ID: "R028", Name: "ImagePullBackOff", Description: "Container image pull in backoff state",
			Severity: models.SeverityCritical, Score: 84,
			Recommendation: "Fix image reference or registry authentication. Wait for backoff to clear.",
			Detect: func(ctx *models.AnalysisContext) []string {
				return containerWaiting(ctx, "ImagePullBackOff")
			},
		},
		{
			ID: "R029", Name: "CrashLoopBackOff", Description: "Container in crash loop backoff",
			Severity: models.SeverityCritical, Score: 86,
			Recommendation: "Examine container logs and fix the crash cause. Check startup probes.",
			Detect: func(ctx *models.AnalysisContext) []string {
				return containerWaiting(ctx, "CrashLoopBackOff")
			},
		},
		{
			ID: "R030", Name: "Executor Timeout", Description: "Executor heartbeat timeout",
			Severity: models.SeverityHigh, Score: 77,
			Recommendation: "Increase spark.network.timeout and spark.executor.heartbeatInterval.",
			Detect: func(ctx *models.AnalysisContext) []string {
				return logPatterns(ctx, "Executor heartbeat timed out", "heartbeat timeout", "Timed out waiting for executor")
			},
		},
		{
			ID: "R031", Name: "Shuffle Service Failure", Description: "External shuffle service failure",
			Severity: models.SeverityMedium, Score: 58,
			Recommendation: "Verify shuffle service is running or disable external shuffle service.",
			Detect: func(ctx *models.AnalysisContext) []string {
				return logPatterns(ctx, "shuffle service", "ExternalShuffleService", "shuffle server")
			},
		},
		{
			ID: "R032", Name: "Dynamic Allocation Failure", Description: "Dynamic allocation unable to acquire executors",
			Severity: models.SeverityMedium, Score: 62,
			Recommendation: "Check cluster capacity, shuffle service, and dynamic allocation settings.",
			Detect: func(ctx *models.AnalysisContext) []string {
				if val, ok := ctx.SparkConf.AllConf["spark.dynamicAllocation.enabled"]; ok && val == "true" {
					ev := logPatterns(ctx, "Unable to acquire executor", "dynamic allocation", "No available executors")
					if len(ev) > 0 {
						return ev
					}
				}
				return nil
			},
		},
		{
			ID: "R033", Name: "Speculation Issue", Description: "Task speculation detected repeated failures",
			Severity: models.SeverityMedium, Score: 50,
			Recommendation: "Investigate straggler tasks. Tune spark.speculation settings.",
			Detect: func(ctx *models.AnalysisContext) []string {
				return logPatterns(ctx, "Speculative task", "speculation", "straggler")
			},
		},
		{
			ID: "R034", Name: "OOM Score High", Description: "Pod at risk of OOM kill due to high OOM score",
			Severity: models.SeverityMedium, Score: 55,
			Recommendation: "Increase memory limits or reduce memory usage to avoid OOM kill.",
			Detect: func(ctx *models.AnalysisContext) []string {
				return eventReasons(ctx, "OOMKilling", "SystemOOM")
			},
		},
		{
			ID: "R035", Name: "ContainerKilled", Description: "Container was killed (non-OOM)",
			Severity: models.SeverityHigh, Score: 72,
			Recommendation: "Check exit codes, liveness probes, and resource limits.",
			Detect: func(ctx *models.AnalysisContext) []string {
				return append(eventReasons(ctx, "Killing"),
					descPatterns(ctx, "Container killed", "Exit Code")...)
			},
		},
		{
			ID: "R036", Name: "SparkConf Invalid", Description: "Invalid Spark configuration detected",
			Severity: models.SeverityMedium, Score: 52,
			Recommendation: "Review Spark configuration for typos and invalid values.",
			Detect: func(ctx *models.AnalysisContext) []string {
				return logPatterns(ctx, "SparkConf error", "Invalid Spark config", "unknown configuration", "IllegalArgumentException")
			},
		},
		{
			ID: "R037", Name: "Spark Version Mismatch", Description: "Spark version mismatch between components",
			Severity: models.SeverityMedium, Score: 58,
			Recommendation: "Ensure driver and executor use the same Spark version and image.",
			Detect: func(ctx *models.AnalysisContext) []string {
				return logPatterns(ctx, "Spark version mismatch", "incompatible Spark version", "Version mismatch")
			},
		},
		{
			ID: "R038", Name: "Classpath Issue", Description: "Classpath or class loading error",
			Severity: models.SeverityHigh, Score: 73,
			Recommendation: "Verify JAR dependencies, --packages, and classpath configuration.",
			Detect: func(ctx *models.AnalysisContext) []string {
				return logPatterns(ctx, "ClassNotFoundException", "NoClassDefFoundError", "NoSuchMethodError", "classpath")
			},
		},
		{
			ID: "R039", Name: "Jar Missing", Description: "Application JAR file not found",
			Severity: models.SeverityCritical, Score: 85,
			Recommendation: "Verify mainApplicationFile path, S3/HDFS accessibility, and JAR upload.",
			Detect: func(ctx *models.AnalysisContext) []string {
				return logPatterns(ctx, "FileNotFoundException", "jar not found", "does not exist", "No such file or directory")
			},
		},
		{
			ID: "R040", Name: "Dependency Resolution Failure", Description: "Failed to resolve Maven/Ivy dependencies",
			Severity: models.SeverityHigh, Score: 70,
			Recommendation: "Check --packages coordinates, repository access, and network connectivity.",
			Detect: func(ctx *models.AnalysisContext) []string {
				return logPatterns(ctx, "ResolutionException", "Failed to resolve dependencies", "not found in central", "Ivy")
			},
		},
		{
			ID: "R041", Name: "Unsupported Hadoop Version", Description: "Hadoop version incompatibility",
			Severity: models.SeverityMedium, Score: 55,
			Recommendation: "Use Spark build compatible with your Hadoop version.",
			Detect: func(ctx *models.AnalysisContext) []string {
				return logPatterns(ctx, "Unsupported Hadoop", "Hadoop version", "Incompatible Hadoop")
			},
		},
		{
			ID: "R042", Name: "DeadlineExceeded", Description: "Operation exceeded deadline",
			Severity: models.SeverityHigh, Score: 68,
			Recommendation: "Increase timeout settings or optimize slow operations.",
			Detect: func(ctx *models.AnalysisContext) []string {
				return eventReasons(ctx, "DeadlineExceeded")
			},
		},
		{
			ID: "R043", Name: "BackOff", Description: "Container restart backoff active",
			Severity: models.SeverityMedium, Score: 65,
			Recommendation: "Fix underlying container failure to exit backoff state.",
			Detect: func(ctx *models.AnalysisContext) []string {
				return append(eventReasons(ctx, "BackOff"),
					containerWaiting(ctx, "BackOff")...)
			},
		},
		{
			ID: "R044", Name: "Evicted", Description: "Pod was evicted from node",
			Severity: models.SeverityHigh, Score: 78,
			Recommendation: "Address node resource pressure. Consider pod priority and resource limits.",
			DocLinks:       []string{"https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/"},
			Detect: func(ctx *models.AnalysisContext) []string {
				return eventReasons(ctx, "Evicted")
			},
		},
		{
			ID: "R045", Name: "FailedScheduling", Description: "Pod could not be scheduled",
			Severity: models.SeverityHigh, Score: 76,
			Recommendation: "Check node resources, affinity rules, taints, and PVC binding.",
			Detect: func(ctx *models.AnalysisContext) []string {
				return eventReasons(ctx, "FailedScheduling")
			},
		},
	}
}

func itoa(n int) string {
	return strconv.Itoa(n)
}
