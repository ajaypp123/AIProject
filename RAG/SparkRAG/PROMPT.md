# Enterprise Spark Knowledge RAG Platform

You are a Principal Software Engineer with expertise in Go, Python, Enterprise RAG systems, Distributed Systems, Apache Spark, Docker, REST APIs and AI platforms.

Generate a **production-quality, enterprise-grade RAG platform**.

The project must compile successfully and every file must be fully implemented.

Do not leave TODOs.

---

# PROJECT NAME

spark-rag-platform

---

# GOAL

Build an Enterprise Spark Knowledge Assistant.

The platform should answer questions about

* Apache Spark
* Hadoop
* Iceberg
* Delta Lake
* Hive
* Kubernetes
* Presto
* Spark SQL
* Spark Configuration
* Spark Performance
* Spark Troubleshooting
* Previous Incident Documents
* Internal Documentation

The project must support local execution using Docker Compose.

No Kubernetes deployment is required.

---

# HIGH LEVEL ARCHITECTURE

```
                 React Chat UI
                        │
                        │ REST
                        ▼
               spark-rag-api (Go)
                        │
          ┌─────────────┼─────────────┐
          │             │             │
          ▼             ▼             ▼
      Retriever     Vector DB      LLM
          │
          ▼
     Qdrant / Chroma
          ▲
          │
 spark-ingestor (Python)
          │
          ▼
 Documents / PDFs / Markdown / CSV / Jira Export
```

---

# SERVICES

Generate the following services.

## 1

spark-rag-api

Language

Go

Responsibilities

REST API

Search

Chat

Configuration

Retriever

Prompt Builder

LLM Gateway

Citation Builder

Health

Metrics

CLI

---

## 2

spark-ingestor

Language

Python

Responsibilities

Watch folders

Load documents

Chunk

Generate embeddings

Store vectors

Incremental indexing

Metadata extraction

---

## 3

Vector Database

Must support

Qdrant

ChromaDB

Provider selected using configuration.

---

## 4

spark-chat-ui

React

Responsibilities

Simple Chat UI

Conversation history

Markdown rendering

Source citations

Dark mode

Streaming responses

---

# SUPPORTED DOCUMENT TYPES

Markdown

PDF

TXT

CSV

JIRA Export JSON

JIRA Export CSV

HTML

Directory Recursively

ZIP archives

Git repositories

---

# DOCUMENT INGESTION

Implement

Recursive document discovery

Incremental indexing

File checksum

Version tracking

Metadata extraction

Automatic reindexing

Document deletion detection

Supported metadata

Filename

Title

Version

Source

Tags

Language

Last Modified

Document Type

---

# CHUNKING

Support

Recursive chunking

Markdown-aware chunking

Heading-aware chunking

Sentence chunking

Configurable

Chunk Size

Chunk Overlap

---

# EMBEDDINGS

Support

SentenceTransformers

BAAI/bge-base-en-v1.5

BAAI/bge-large-en-v1.5

all-MiniLM-L6-v2

Nomic Embed

OpenAI Embeddings

Provider selected using configuration.

---

# VECTOR DATABASE

Support

Qdrant

ChromaDB

Selection entirely configuration driven.

Implement repository abstraction.

No vector DB specific logic outside repository layer.

---

# RETRIEVAL PIPELINE

Question

↓

Embedding

↓

Similarity Search

↓

Metadata Filter

↓

Hybrid Search

↓

MMR

↓

Reranker

↓

Top K

↓

Prompt Builder

↓

LLM

↓

Answer

Implement retrieval strategies

Similarity

MMR

Hybrid

Configurable.

---

# RERANKER

Support

CrossEncoder

BGE Reranker

Can be disabled.

---

# LLM PROVIDERS

Support

OpenAI

Azure OpenAI

Claude

Gemini

Ollama

OpenRouter

Everything configurable.

---

# CONFIGURATION

Support

config.yaml

.env

Environment variables

CLI overrides

No hardcoded values.

Example

server

port

vectordb

provider

url

api_key

collection

embedding

provider

model

llm

provider

url

model

api_key

retriever

top_k

rerank

chunk_size

chunk_overlap

logging

level

---

# REST APIs

POST /api/chat

POST /api/search

POST /api/ingest

POST /api/reindex

POST /api/delete

GET /api/documents

GET /api/providers

GET /api/collections

GET /health

GET /metrics

---

# CLI

Implement

spark-rag ingest

spark-rag search

spark-rag ask

spark-rag providers

spark-rag collections

spark-rag stats

spark-rag delete

spark-rag rebuild

spark-rag config

Example

spark-rag ask "Explain Adaptive Query Execution"

---

# CHAT FEATURES

Conversation memory

Streaming

Markdown

Source citations

Related documents

Suggested follow-up questions

Confidence score

---

# CITATIONS

Every answer must include

Source

Document

Section

Page (PDF)

Line if available

Document URL

---

# SEARCH FILTERS

Support

Tags

Version

Language

Document Type

Date

Source

---

# DOCKER

Provide Dockerfiles for

spark-rag-api

spark-ingestor

spark-chat-ui

Use multi-stage builds.

Run as non-root.

Healthcheck enabled.

---

# DOCKER COMPOSE

Generate docker-compose.yml.

Running

docker compose up

should start

spark-rag-api

spark-ingestor

qdrant

chromadb (optional profile)

ollama (optional profile)

react ui

Persistent Docker volumes.

---

# MAKEFILE

Targets

build

run

test

lint

fmt

docker

clean

compose-up

compose-down

---

# BUILD SCRIPT

scripts/build.sh

Runs

fmt

lint

tests

build Go

build Python package

docker builds

Print versions

---

# TESTING

Go

Unit Tests

API Tests

Repository Tests

Retriever Tests

Configuration Tests

Python

Chunking Tests

Embedding Tests

Parser Tests

Incremental Index Tests

Mock LLM

Mock Vector DB

Minimum 80% coverage.

---

# SAMPLE DATA

Provide

Example Spark Documentation

Example Hadoop Documentation

Example Markdown

Example PDF

Example CSV

Example Jira Export

Example search results

---

# README

Include

Architecture Diagram

Installation

Configuration

Docker Compose

CLI Usage

REST API Examples

Adding Documents

Switching Vector Databases

Switching LLM Providers

Adding New Embedding Models

Development Guide

Project Structure

Screenshots placeholders

Roadmap

Troubleshooting

---

# CODE QUALITY

Use clean architecture.

Dependency injection.

Repository pattern.

Interfaces.

Context everywhere.

Structured logging.

Graceful shutdown.

Retry logic.

Error wrapping.

No duplicated code.

---

# FUTURE EXTENSIBILITY

Design so that later we can add

VS Code extension

Slack Bot

Discord Bot

MCP Server

GitHub indexing

Confluence indexing

SharePoint indexing

without changing core retrieval logic.

---

# FINAL REQUIREMENTS

Generate the complete project.

Every source file must be fully implemented.

No TODOs.

The project must compile.

Docker Compose must work with a single command.

The architecture should resemble a real production internal knowledge platform rather than a simple RAG demo.
