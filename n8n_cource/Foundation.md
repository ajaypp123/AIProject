
# What is n8n

n8n is a free, open-source workflow automation platform. It’s designed to be a universal connector so you can visually compose workflows that pass data between apps, services, and AI models without heavy custom code.

# Nodes: the building blocks

Everything in n8n is a node. Nodes come in two primary categories:

| Node Type | Role | Example |
| - | - | - |
| Trigger | Starts workflows when an event occurs | Webhook, Email trigger, Chat trigger |
| Action | Performs work and returns output | Send email, Call external API, Run LLM inference |

# Workflow

<Frame>
  <img src="https://mintcdn.com/kodekloud-c4ac6d9a/zSiqgjMftx8xoTrJ/images/n8n-Zero-to-Hero/n8n-Foundational/Foundations-of-n8n/n8n-workflow-architecture-ai-agent.jpg?fit=max&auto=format&n=zSiqgjMftx8xoTrJ&q=85&s=fbf3f2d6dc43020119c40aaac49fcc49" alt="The image depicts a workflow architecture for n8n, showing an AI agent integrating with API tokens, LLMs, memory, tools, and nodes for event triggering and progression." width="1920" height="1080" data-path="images/n8n-Zero-to-Hero/n8n-Foundational/Foundations-of-n8n/n8n-workflow-architecture-ai-agent.jpg" />
</Frame>

# Branching and parallel paths

A single trigger can branch into multiple independent paths.

* Branch A: post to Slack
* Branch B: send email

# Memory for AI agents

AI agents typically use two memory styles:

| Memory Type | Purpose | Example / Technologies |
| - | - | - |
| Context memory (short-term) | Keeps immediate chat history and prompt state for coherent conversations | Pass chat messages in the prompt to the LLM |
| Long-term searchable memory (RAG) | Stores documents as embeddings and retrieves relevant context | Pinecone, Weaviate — Retrieval-Augmented Generation (RAG) |

# Inputs and outputs for nodes

n8n can display the same underlying data in multiple formats:

* Schema: a structured definition for fields.
* Table: spreadsheet-like rows and columns.
* JSON: the most common structured representation for programmatic access.
* Binary: files such as images, PDFs, and other blob data.

# Expressions and dynamic fields

Node fields accept:
* Static values (fixed strings), or
* Expressions that reference earlier nodes or runtime data.

Examples:
* Bracket notation: `{{ $json["chatInput"] }}`
* Dot notation: `{{ $json.user.name }}`

### Data types used in n8n

Common data types you’ll encounter:

| Type | Description | Example |
| - | - | - |
| String | Text values | `hello world` |
| Number | Numeric values (ints and floats) | `42`, `3.14` |
| Boolean | True/false | `true` or `false` |
| Array | Ordered lists | `["James","Tony","Samantha"]` |
| Object | Structured key/value data | See example below |

### Community nodes

n8n has a vibrant community that builds connectors (community nodes). If an integration you need is not included by default, there’s a good chance someone in the community has implemented it.
