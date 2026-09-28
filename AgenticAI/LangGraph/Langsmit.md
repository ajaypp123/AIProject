> ## Documentation Index
> Fetch the complete documentation index at: https://notes.kodekloud.com/llms.txt
> Use this file to discover all available pages before exploring further.

# LangSmith Observability Introduction

> Introduction to LangSmith observability showing how execution traces in LangGraph and LangChain record prompts, tool calls, state changes, errors, and performance for debugging

When you deploy AI agents in production, observability becomes critical. Without visibility into internal execution—what nodes ran, which prompts were used, what inputs were processed, and how state changed—debugging is nearly impossible. LangSmith provides observability tools built specifically for large language model (LLM) applications to make these internals visible and actionable.

LangSmith is an observability platform created by the [LangChain](https://learn.kodekloud.com/user/courses/langchain) team for monitoring, debugging, and improving systems built with frameworks like [LangChain](https://learn.kodekloud.com/user/courses/langchain) and [LangGraph](https://learn.kodekloud.com/user/courses/langgraph). Its core concept is the execution trace: every run of your application is recorded in full. Traces capture model inputs and outputs, prompt versions, tool calls, errors, latency, and workflow state changes so you can inspect exactly how an agent behaved.

<Frame>
  <img src="https://mintcdn.com/kodekloud-c4ac6d9a/s58V3Wk57W0ne4D5/images/LangGraph/Long-Term-Memory-and-Stateful-Persistence/LangSmith-Observability-Introduction/langsmith-overview-execution-trace-factors.jpg?fit=max&auto=format&n=s58V3Wk57W0ne4D5&q=85&s=a39e88d59b725ec58548decc08c3be9a" alt="The image is an overview of LangSmith, a platform for applications built with LangChain and LangGraph, featuring a central &#x22;Execution Trace&#x22; surrounded by factors like Prompt Versions, Inputs, Outputs, Errors, Latency, and State Transitions." width="1920" height="1080" data-path="images/LangGraph/Long-Term-Memory-and-Stateful-Persistence/LangSmith-Observability-Introduction/langsmith-overview-execution-trace-factors.jpg" />
</Frame>

You can trace each step of a workflow to find where errors occurred, examine which prompts and tools contributed to a result, and analyze execution latency. Practically, LangSmith acts like mission control for LLM applications, giving developers complete visibility into agent internals.

Every time an agent runs, multiple components interact: models receive inputs, prompts are constructed and versioned, tools may be called, graph state can change, and the model generates outputs. Complex agent systems may perform thousands of decisions and interactions in a single run. Without observability, this all looks like a black box—you only see the final response.

<Frame>
  <img src="https://mintcdn.com/kodekloud-c4ac6d9a/A72Yow2kiLd9Lv77/images/LangGraph/Long-Term-Memory-and-Stateful-Persistence/LangSmith-Observability-Introduction/langsmith-overview-robot-data-charts.jpg?fit=max&auto=format&n=A72Yow2kiLd9Lv77&q=85&s=6a74a482855506302d0f3970ba7aa929" alt="The image titled &#x22;LangSmith – Overview&#x22; shows a robot with data charts on screens, accompanied by a text bubble explaining that LangSmith records traces whenever an agent runs." width="1920" height="1080" data-path="images/LangGraph/Long-Term-Memory-and-Stateful-Persistence/LangSmith-Observability-Introduction/langsmith-overview-robot-data-charts.jpg" />
</Frame>

LangSmith records an execution trace whenever your agent runs. A trace is a detailed timeline describing:

* Model inputs and outputs
* Tool calls and external API interactions
* Node execution sequence and state transitions
* Latency measurements and timing per step
* Errors and full tracebacks
* Prompt versions used during the run

Inspecting traces lets you see the exact prompt that was sent, which tool was invoked, how state evolved, and where failures or slowdowns occurred—turning guesswork into diagnosis.

Robby used to just deliver packages. Now his clipboard tracks every step — which doors he knocked on, what he said, and where he had issues. LangSmith is that clipboard, but for your graph.

[LangGraph](https://learn.kodekloud.com/user/courses/langgraph) integrates natively with LangSmith, so enabling tracing usually requires only a few lines of configuration. Once enabled, detailed traces are automatically sent to LangSmith for every LangGraph run.

<Frame>
  <img src="https://mintcdn.com/kodekloud-c4ac6d9a/s58V3Wk57W0ne4D5/images/LangGraph/Long-Term-Memory-and-Stateful-Persistence/LangSmith-Observability-Introduction/langgraph-integration-coding-laptop.jpg?fit=max&auto=format&n=s58V3Wk57W0ne4D5&q=85&s=6b29d97b9af6ebeb15e6c9c8440a36a0" alt="The image shows a concept of &#x22;LangGraph Integration&#x22; with a person sitting at a desk, coding on a laptop, and integration lines connecting &#x22;LangGraph&#x22; to &#x22;LangSmith.&#x22; It suggests enabling tracing in code." width="1920" height="1080" data-path="images/LangGraph/Long-Term-Memory-and-Stateful-Persistence/LangSmith-Observability-Introduction/langgraph-integration-coding-laptop.jpg" />
</Frame>

Traces include per-step details (inputs, outputs, tool calls, node transitions, latency) and can be enriched with custom metadata such as session IDs, user IDs, or experiment tags. This metadata helps correlate user activity and configuration with system behavior—critical for robust debugging and production monitoring.

Getting started with LangSmith tracing is simple. Add the tracing initialization early in your application:

```python theme={null}
from langgraph.tracing.langsmith import enable_tracing

# Enable LangSmith tracing for all LangGraph runs
enable_tracing()
```

When tracing is enabled, LangGraph opens a trace session for each run and captures events like node execution, model calls, tool invocations, state updates, and timings. All captured data is sent to LangSmith and stored as an execution trace for later inspection.

<Callout icon="lightbulb" color="#1CB2FE">
  Initialize tracing before your graph executes. Tracing begins only after `enable_tracing()` is called, so place it at the top of your script, app startup, or notebook cell that runs prior to graph execution.
</Callout>

LangGraph’s tracing is non-invasive: you do not need to modify node functions or change workflow logic. The framework records execution details without altering runtime behavior.

<Frame>
  <img src="https://mintcdn.com/kodekloud-c4ac6d9a/s58V3Wk57W0ne4D5/images/LangGraph/Long-Term-Memory-and-Stateful-Persistence/LangSmith-Observability-Introduction/langsmith-automatic-tracking-features-outline.jpg?fit=max&auto=format&n=s58V3Wk57W0ne4D5&q=85&s=507c9b0b5d842b4d88ccddc8aceaf354" alt="The image outlines what LangSmith automatically tracks, highlighting features like automatic tracking, non-invasive operation, and important placement. Each feature is accompanied by a brief description." width="1920" height="1080" data-path="images/LangGraph/Long-Term-Memory-and-Stateful-Persistence/LangSmith-Observability-Introduction/langsmith-automatic-tracking-features-outline.jpg" />
</Frame>

What LangSmith records (at a glance)

| Layer             | What is captured                          | Why it matters                                |
| ----------------- | ----------------------------------------- | --------------------------------------------- |
| Execution path    | Sequence of nodes executed in the graph   | See exactly which branches ran and why        |
| Model/Chain calls | Model inputs, outputs, prompt versions    | Reproduce responses and audit prompts         |
| Tool usage        | Tool names, arguments, results, API calls | Debug external integrations and failures      |
| Agent decisions   | Intermediate reasoning steps              | Understand decision points and logic flow     |
| Performance       | Latency and per-step timings              | Identify bottlenecks and optimize performance |
| Errors            | Full tracebacks and failure context       | Faster root-cause analysis                    |

In local testing, prints and logs can sometimes be enough. But production workflows typically have branching logic, external calls, stateful memory, and unpredictable user inputs. Two runs with similar inputs may take different execution paths; traces let you compare runs side-by-side.

<Frame>
  <img src="https://mintcdn.com/kodekloud-c4ac6d9a/A72Yow2kiLd9Lv77/images/LangGraph/Long-Term-Memory-and-Stateful-Persistence/LangSmith-Observability-Introduction/testing-approach-vs-production-challenges.jpg?fit=max&auto=format&n=A72Yow2kiLd9Lv77&q=85&s=e65d3c7be23bc0cf707036e299289dc2" alt="The image contrasts a simple and insufficient testing approach with the complex and demanding realities of production, highlighting the challenges of debugging in a production environment due to multiple conditional paths, real user input, and memory updates." width="1920" height="1080" data-path="images/LangGraph/Long-Term-Memory-and-Stateful-Persistence/LangSmith-Observability-Introduction/testing-approach-vs-production-challenges.jpg" />
</Frame>

Key investigation tools in LangSmith

| Feature                 | How it helps                                                       |
| ----------------------- | ------------------------------------------------------------------ |
| Visual node traces      | See the path the graph took visually and quickly identify branches |
| Step-by-step breakdowns | Inspect inputs/outputs for each node and tool call                 |
| Prompt version history  | Audit which prompt template or version was used for a given run    |
| Session replay          | Replay previous runs to understand behavior over time              |
| Detailed error context  | View full tracebacks and surrounding state to accelerate fixes     |

<Frame>
  <img src="https://mintcdn.com/kodekloud-c4ac6d9a/s58V3Wk57W0ne4D5/images/LangGraph/Long-Term-Memory-and-Stateful-Persistence/LangSmith-Observability-Introduction/langsmith-observability-features-list.jpg?fit=max&auto=format&n=s58V3Wk57W0ne4D5&q=85&s=d691a0ab2536bddf91270b5a1fc409ab" alt="The image lists key observability features in LangSmith, including visual node traces, step-by-step breakdowns, prompt version history, session replay, and error context. Each feature includes a short description of its function." width="1920" height="1080" data-path="images/LangGraph/Long-Term-Memory-and-Stateful-Persistence/LangSmith-Observability-Introduction/langsmith-observability-features-list.jpg" />
</Frame>

Best practices for production observability

* Enable tracing for development and staging to detect regressions early.
* Use session tags (user IDs, agent names, experiment IDs) to filter and group traces.
* Attach custom metadata (prompt version, model version, memory size) to correlate configuration with behavior.
* Review traces for subtle issues (unexpected outputs, memory mis-recall), not just explicit errors.
* Configure dashboards and alerts to surface critical signals like latency spikes or recurring failures.

<Frame>
  <img src="https://mintcdn.com/kodekloud-c4ac6d9a/A72Yow2kiLd9Lv77/images/LangGraph/Long-Term-Memory-and-Stateful-Persistence/LangSmith-Observability-Introduction/tracing-best-practices-software-environments.jpg?fit=max&auto=format&n=A72Yow2kiLd9Lv77&q=85&s=1bffd8dcb00f67f502db7b86836ff878" alt="The image presents four best practices for tracing in software environments, including enabling tracing in development and staging, using session tags, logging custom metadata, and reviewing traces for unexpected outputs." width="1920" height="1080" data-path="images/LangGraph/Long-Term-Memory-and-Stateful-Persistence/LangSmith-Observability-Introduction/tracing-best-practices-software-environments.jpg" />
</Frame>

LangSmith gives you a visual trace of LangGraph execution: which nodes ran, what each step did, and how long each took. If a run fails, inspect the full error context; if it’s slow, identify the slow step. This shifts debugging from speculative to diagnostic.

<Frame>
  <img src="https://mintcdn.com/kodekloud-c4ac6d9a/A72Yow2kiLd9Lv77/images/LangGraph/Long-Term-Memory-and-Stateful-Persistence/LangSmith-Observability-Introduction/langsmith-trace-execution-flow-diagram.jpg?fit=max&auto=format&n=A72Yow2kiLd9Lv77&q=85&s=9db106d5ef98f961e36b9e198e8fbf13" alt="The image illustrates a &#x22;LangSmith Trace&#x22; analysis with a visual execution flow diagram, highlighting error analysis and performance bottlenecks. It shows a step-by-step process with time durations and emphasizes the importance of transforming debugging from guesswork into diagnosis." width="1920" height="1080" data-path="images/LangGraph/Long-Term-Memory-and-Stateful-Persistence/LangSmith-Observability-Introduction/langsmith-trace-execution-flow-diagram.jpg" />
</Frame>

Troubleshooting and validation use cases

* Answering tough questions: When two similar inputs yield different outputs, traces reveal which nodes executed, the exact prompts sent, and any memory or state accessed.
* Validating changes: After modifying prompts, reducers, or memory logic, use traces to confirm the new behavior before releasing to users.
* Performance tuning: Identify slow tools or model calls and prioritize optimizations based on per-step timings.

<Frame>
  <img src="https://mintcdn.com/kodekloud-c4ac6d9a/A72Yow2kiLd9Lv77/images/LangGraph/Long-Term-Memory-and-Stateful-Persistence/LangSmith-Observability-Introduction/observability-use-cases-questions-changes.jpg?fit=max&auto=format&n=A72Yow2kiLd9Lv77&q=85&s=bdf64ba9307aaf6cefed39e289382bb7" alt="The image lists use cases for observability, categorized into &#x22;Answering Tough Questions&#x22; and &#x22;Validating New Changes,&#x22; with specific scenarios for each category." width="1920" height="1080" data-path="images/LangGraph/Long-Term-Memory-and-Stateful-Persistence/LangSmith-Observability-Introduction/observability-use-cases-questions-changes.jpg" />
</Frame>

Operationalizing observability

* Enrich traces with metadata (user IDs, input types, version tags) for easier filtering and correlation.
* Add monitoring, dashboards, and alerts to detect important signals automatically (latency spikes, frequent retries, long execution chains).
* After updates (prompts, reducers, memory), inspect sample traces across key flows to ensure expected behavior.

<Frame>
  <img src="https://mintcdn.com/kodekloud-c4ac6d9a/A72Yow2kiLd9Lv77/images/LangGraph/Long-Term-Memory-and-Stateful-Persistence/LangSmith-Observability-Introduction/operational-observability-practices-diagram.jpg?fit=max&auto=format&n=A72Yow2kiLd9Lv77&q=85&s=8c654fd7a0c17db2c4290ce554968dde" alt="The image outlines &#x22;Operational Observability Practices&#x22; with sections on enriching traces with metadata, configuring monitoring and alerts, and reviewing key flows post-update, each with specific steps and diagrams." width="1920" height="1080" data-path="images/LangGraph/Long-Term-Memory-and-Stateful-Persistence/LangSmith-Observability-Introduction/operational-observability-practices-diagram.jpg" />
</Frame>

LangGraph enables powerful, stateful agent workflows—and that power brings complexity. As graphs grow, tracing becomes essential to understand actual execution, debug failures, and validate changes. Observability tools like LangSmith help teams build AI systems that are more robust, explainable, and maintainable.

<Callout icon="warning" color="#FF6B6B">
  Be mindful of sensitive data. Traces can capture user inputs and API responses—sanitize or redact personally identifiable information (PII) and secrets before storing or sharing traces.
</Callout>

For engineers building production agents, LangSmith becomes an essential part of the monitoring and debugging toolkit—turning opaque agent runs into inspectable, repeatable traces.

Links and references

* [LangGraph course on KodeKloud](https://learn.kodekloud.com/user/courses/langgraph)
* [LangChain course on KodeKloud](https://learn.kodekloud.com/user/courses/langchain)
* [LangChain documentation](https://langchain.readthedocs.io/)

<CardGroup>
  <Card title="Watch Video" icon="video" cta="Learn more" href="https://learn.kodekloud.com/user/courses/langgraph/module/e0cd494a-00a7-4c52-88e9-b3932b03ff9f/lesson/32b29c2b-a9e8-427d-94d5-cded50b77f6a" />

  <Card title="Practice Lab" icon="flask-conical" cta="Learn more" href="https://learn.kodekloud.com/user/courses/langgraph/module/e0cd494a-00a7-4c52-88e9-b3932b03ff9f/lesson/f8153413-b11a-4e16-9ff0-bfe795bfe05e" />
</CardGroup>