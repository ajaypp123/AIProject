#!/usr/bin/env python
"""
FastAPI wrapper for SentenceTransformer embeddings.

Wraps the local SentenceTransformer model and exposes it via HTTP API.
This allows the Go API to use local embeddings without needing to implement
the ML model in Go.

Usage:
    python embedding_service.py --model sentence-transformers/all-MiniLM-L6-v2 --port 8000

API Endpoints:
    POST /embed - Generate single embedding
    POST /embed-batch - Generate multiple embeddings
    GET /health - Health check
"""

import argparse
import logging
from typing import List

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import uvicorn
from sentence_transformers import SentenceTransformer

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(title="Sentence Transformer Embedding Service")

# Global model instance
model = None


class EmbedRequest(BaseModel):
    """Single text embedding request."""
    text: str


class EmbedBatchRequest(BaseModel):
    """Batch text embedding request."""
    texts: List[str]


class EmbedResponse(BaseModel):
    """Single text embedding response."""
    embedding: List[float]


class EmbedBatchResponse(BaseModel):
    """Batch text embedding response."""
    embeddings: List[List[float]]


@app.on_event("startup")
async def startup_event():
    """Load model on startup."""
    global model
    try:
        logger.info(f"Loading SentenceTransformer model: {args.model}")
        model = SentenceTransformer(args.model)
        embedding_dim = model.get_embedding_dimension()
        logger.info(f"Model loaded successfully. Embedding dimension: {embedding_dim}")
    except Exception as e:
        logger.error(f"Failed to load model: {e}")
        raise


@app.get("/health")
async def health_check():
    """Health check endpoint."""
    if model is None:
        raise HTTPException(status_code=503, detail="Model not loaded")
    return {"status": "healthy", "model": args.model}


@app.post("/embed", response_model=EmbedResponse)
async def embed(request: EmbedRequest):
    """Generate embedding for single text."""
    if model is None:
        raise HTTPException(status_code=503, detail="Model not loaded")

    try:
        embedding = model.encode(request.text)
        return EmbedResponse(embedding=embedding.tolist())
    except Exception as e:
        logger.error(f"Embedding failed: {e}")
        raise HTTPException(status_code=500, detail=f"Embedding failed: {e}")


@app.post("/embed-batch", response_model=EmbedBatchResponse)
async def embed_batch(request: EmbedBatchRequest):
    """Generate embeddings for multiple texts."""
    if model is None:
        raise HTTPException(status_code=503, detail="Model not loaded")

    if not request.texts:
        raise HTTPException(status_code=400, detail="texts list is empty")

    try:
        embeddings = model.encode(request.texts, batch_size=32, show_progress_bar=False)
        return EmbedBatchResponse(embeddings=embeddings.tolist())
    except Exception as e:
        logger.error(f"Batch embedding failed: {e}")
        raise HTTPException(status_code=500, detail=f"Batch embedding failed: {e}")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Sentence Transformer Embedding Service")
    parser.add_argument(
        "--model",
        type=str,
        default="sentence-transformers/all-MiniLM-L6-v2",
        help="HuggingFace model name"
    )
    parser.add_argument(
        "--port",
        type=int,
        default=8000,
        help="Port to run service on"
    )
    parser.add_argument(
        "--host",
        type=str,
        default="0.0.0.0",
        help="Host to bind to"
    )

    args = parser.parse_args()

    logger.info(f"Starting Sentence Transformer Embedding Service on {args.host}:{args.port}")
    uvicorn.run(app, host=args.host, port=args.port)
