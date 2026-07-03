"""
Embedding providers for document vectorization.

Supports multiple embedding backends:
- Local: SentenceTransformers and HuggingFace models
- Ollama: Local inference via Ollama API
- OpenAI: Cloud-based embeddings
"""

import logging
from typing import List, Optional

import requests
import hashlib
import random

logger = logging.getLogger(__name__)


class EmbeddingProvider:
    """Base interface for embedding providers."""

    def embed(self, text: str) -> Optional[List[float]]:
        """
        Generate embedding for a single text.

        Args:
            text: Text to embed

        Returns:
            List of floats representing the embedding, or None on error
        """
        raise NotImplementedError

    def embed_batch(self, texts: List[str]) -> Optional[List[List[float]]]:
        """
        Generate embeddings for multiple texts.

        Args:
            texts: List of texts to embed

        Returns:
            List of embeddings, or None on error
        """
        raise NotImplementedError


class OllamaEmbedding(EmbeddingProvider):
    """
    Ollama-based local embeddings via HTTP API.

    Requires Ollama service running and model pulled:
        ollama pull nomic-embed-text
        ollama pull all-minilm
    """

    def __init__(self, base_url: str, model: str):
        """
        Initialize Ollama embedding provider.

        Args:
            base_url: Ollama API base URL (e.g., http://localhost:11434)
            model: Model name to use (e.g., nomic-embed-text)
        """
        self.base_url = base_url
        self.model = model
        logger.info(f"Initialized OllamaEmbedding with model {model}")

    def embed(self, text: str) -> Optional[List[float]]:
        """
        Generate embedding using Ollama API.

        Args:
            text: Text to embed

        Returns:
            Embedding vector or None on error
        """
        try:
            response = requests.post(
                f"{self.base_url}/api/embeddings",
                json={"model": self.model, "prompt": text},
                timeout=60
            )
            response.raise_for_status()
            data = response.json()
            return data.get("embedding")
        except Exception as e:
            logger.error(f"Ollama embedding failed: {e}")
            return None

    def embed_batch(self, texts: List[str]) -> Optional[List[List[float]]]:
        """
        Generate embeddings for multiple texts sequentially.

        Args:
            texts: List of texts to embed

        Returns:
            List of embeddings or None on error
        """
        embeddings = []
        for text in texts:
            emb = self.embed(text)
            if emb is None:
                return None
            embeddings.append(emb)
        return embeddings


class OpenAIEmbedding(EmbeddingProvider):
    """
    OpenAI cloud-based embeddings.

    Requires OPENAI_API_KEY environment variable or api_key parameter.
    Supports multiple models:
        - text-embedding-3-small (affordable)
        - text-embedding-3-large (highest quality)
        - text-embedding-ada-002 (legacy)
    """

    def __init__(self, api_key: str, model: str = "text-embedding-3-small"):
        """
        Initialize OpenAI embedding provider.

        Args:
            api_key: OpenAI API key
            model: Model to use (default: text-embedding-3-small)
        """
        self.api_key = api_key
        self.model = model
        logger.info(f"Initialized OpenAIEmbedding with model {model}")

    def embed(self, text: str) -> Optional[List[float]]:
        """
        Generate embedding using OpenAI API.

        Args:
            text: Text to embed

        Returns:
            Embedding vector or None on error
        """
        try:
            response = requests.post(
                "https://api.openai.com/v1/embeddings",
                json={"input": text, "model": self.model},
                headers={"Authorization": f"Bearer {self.api_key}"},
                timeout=60
            )
            response.raise_for_status()
            data = response.json()
            return data["data"][0]["embedding"]
        except Exception as e:
            logger.error(f"OpenAI embedding failed: {e}")
            return None

    def embed_batch(self, texts: List[str]) -> Optional[List[List[float]]]:
        """
        Generate embeddings for multiple texts in one API call.

        More efficient than single embeddings for batch processing.

        Args:
            texts: List of texts to embed

        Returns:
            List of embeddings or None on error
        """
        try:
            response = requests.post(
                "https://api.openai.com/v1/embeddings",
                json={"input": texts, "model": self.model},
                headers={"Authorization": f"Bearer {self.api_key}"},
                timeout=60
            )
            response.raise_for_status()
            data = response.json()
            # Sort by index to maintain order
            embeddings = [item["embedding"] for item in sorted(data["data"], key=lambda x: x["index"])]
            return embeddings
        except Exception as e:
            logger.error(f"OpenAI batch embedding failed: {e}")
            return None


class SentenceTransformerEmbedding(EmbeddingProvider):
    """
    Local embeddings using HuggingFace SentenceTransformers.

    Supports popular models:
        - sentence-transformers/all-MiniLM-L6-v2 (fast, good quality)
        - sentence-transformers/all-mpnet-base-v2 (high quality)
        - BAAI/bge-base-en-v1.5 (excellent quality)
        - BAAI/bge-large-en-v1.5 (best quality, larger)
        - sentence-transformers/all-distilroberta-v1 (very fast)

    No API calls needed - runs entirely locally.
    Requires sentence-transformers library.
    """

    def __init__(self, model: str = "sentence-transformers/all-MiniLM-L6-v2"):
        """
        Initialize SentenceTransformer embedding provider.

        Args:
            model: HuggingFace model name or path

        Raises:
            ImportError: If sentence-transformers not installed
            OSError: If model cannot be downloaded
        """
        try:
            from sentence_transformers import SentenceTransformer
            logger.info(f"Loading SentenceTransformer model: {model}")
            self.model = SentenceTransformer(model)
            self.embedding_dim = self.model.get_embedding_dimension()
            logger.info(f"Model loaded. Embedding dimension: {self.embedding_dim}")
            self._available = True
        except Exception as e:
            # Surface actionable instructions for common dependency issues
            msg = (
                f"SentenceTransformer model initialization failed: {e}.\n"
                "To fix: ensure 'sentence-transformers' and a compatible 'huggingface-hub' are installed.\n"
                "Try: pip install -U sentence-transformers 'huggingface-hub'\n"
                "If using a virtualenv, run the command in that environment."
            )
            logger.error(msg)
            self._available = False
            self.model = None
            self.embedding_dim = 384

    def embed(self, text: str) -> Optional[List[float]]:
        """
        Generate embedding using SentenceTransformer.

        Args:
            text: Text to embed

        Returns:
            Embedding vector or None on error
        """
        if not self._available:
            logger.error("SentenceTransformer not available. Enable 'embedding.allow_fallback' or install dependencies to use real embeddings.")
            return None
        try:
            embedding = self.model.encode(text)
            return embedding.tolist()
        except Exception as e:
            logger.error(f"SentenceTransformer embedding failed: {e}")
            return None

    def embed_batch(self, texts: List[str]) -> Optional[List[List[float]]]:
        """
        Generate embeddings for multiple texts efficiently.

        Processes texts in batches for better GPU utilization.

        Args:
            texts: List of texts to embed

        Returns:
            List of embeddings or None on error
        """
        if not self._available:
            logger.error("SentenceTransformer not available. Enable 'embedding.allow_fallback' or install dependencies to use real embeddings.")
            return None
        try:
            embeddings = self.model.encode(texts, batch_size=32, show_progress_bar=False)
            return embeddings.tolist()
        except Exception as e:
            logger.error(f"SentenceTransformer batch embedding failed: {e}")
            return None


def _deterministic_embedding(text: str, dim: int = 384) -> List[float]:
    """Generate a deterministic pseudo-embedding for `text`.

    This is a lightweight fallback used when a real embedding model isn't available.
    It is deterministic (same input -> same vector) but not semantically meaningful.
    """
    h = hashlib.sha256(text.encode('utf-8')).hexdigest()
    seed = int(h[:16], 16)
    rnd = random.Random(seed)
    return [rnd.uniform(-1.0, 1.0) for _ in range(dim)]


class HuggingFaceEmbedding(EmbeddingProvider):
    """
    HuggingFace Inference API embeddings (cloud-based).

    Uses HuggingFace Inference API for on-demand embeddings.
    Supports any model from HuggingFace hub.

    Requires HF_TOKEN environment variable or token parameter.
    """

    def __init__(self, api_key: str, model: str = "sentence-transformers/all-MiniLM-L6-v2"):
        """
        Initialize HuggingFace Inference API embedding provider.

        Args:
            api_key: HuggingFace API token
            model: Model name from HuggingFace hub
        """
        self.api_key = api_key
        self.model = model
        self.api_url = f"https://api-inference.huggingface.co/models/{model}"
        logger.info(f"Initialized HuggingFaceEmbedding with model {model}")

    def embed(self, text: str) -> Optional[List[float]]:
        """
        Generate embedding using HuggingFace Inference API.

        Args:
            text: Text to embed

        Returns:
            Embedding vector or None on error
        """
        try:
            response = requests.post(
                self.api_url,
                headers={"Authorization": f"Bearer {self.api_key}"},
                json={"inputs": text},
                timeout=60
            )
            response.raise_for_status()
            data = response.json()
            # HF returns nested list for single input
            if isinstance(data, list) and len(data) > 0:
                return data[0] if isinstance(data[0], list) else data
            return data
        except Exception as e:
            logger.error(f"HuggingFace embedding failed: {e}")
            return None

    def embed_batch(self, texts: List[str]) -> Optional[List[List[float]]]:
        """
        Generate embeddings for multiple texts using HuggingFace API.

        Args:
            texts: List of texts to embed

        Returns:
            List of embeddings or None on error
        """
        try:
            response = requests.post(
                self.api_url,
                headers={"Authorization": f"Bearer {self.api_key}"},
                json={"inputs": texts},
                timeout=60
            )
            response.raise_for_status()
            return response.json()
        except Exception as e:
            logger.error(f"HuggingFace batch embedding failed: {e}")
            return None


def get_embedding_provider(provider: str, **kwargs) -> EmbeddingProvider:
    """
    Factory function to create embedding provider instances.

    Supported providers:
        - "ollama": Local Ollama service
        - "openai": OpenAI cloud API
        - "sentence-transformers": Local HuggingFace models
        - "huggingface": HuggingFace Inference API

    Args:
        provider: Provider type
        **kwargs: Provider-specific arguments:
            - ollama: base_url, model
            - openai: api_key, model
            - sentence-transformers: model
            - huggingface: api_key, model

    Returns:
        Initialized embedding provider

    Raises:
        ValueError: If provider is unknown

    Examples:
        >>> # Local Ollama
        >>> provider = get_embedding_provider("ollama",
        ...     base_url="http://localhost:11434",
        ...     model="nomic-embed-text")

        >>> # Local HuggingFace models
        >>> provider = get_embedding_provider("sentence-transformers",
        ...     model="BAAI/bge-base-en-v1.5")

        >>> # Cloud OpenAI
        >>> provider = get_embedding_provider("openai",
        ...     api_key="sk-...",
        ...     model="text-embedding-3-small")
    """
    provider = provider.lower().strip()

    if provider == "ollama":
        return OllamaEmbedding(
            base_url=kwargs.get("base_url", "http://localhost:11434"),
            model=kwargs.get("model", "nomic-embed-text")
        )

    elif provider == "openai":
        return OpenAIEmbedding(
            api_key=kwargs.get("api_key", ""),
            model=kwargs.get("model", "text-embedding-3-small")
        )

    elif provider in ["sentence-transformers", "local"]:
        ste = SentenceTransformerEmbedding(
            model=kwargs.get("model", "sentence-transformers/all-MiniLM-L6-v2")
        )
        allow_fallback = kwargs.get("allow_fallback", False)
        if ste._available:
            return ste
        if allow_fallback:
            logger.warning("SentenceTransformer unavailable; using deterministic fallback embeddings as configured")
            class FallbackEmbedding(EmbeddingProvider):
                def __init__(self, dim=384):
                    self.dim = dim
                def embed(self, text: str):
                    return _deterministic_embedding(text, self.dim)
                def embed_batch(self, texts: List[str]):
                    return [ _deterministic_embedding(t, self.dim) for t in texts ]
            return FallbackEmbedding(dim=ste.embedding_dim)
        # Not available and no fallback allowed -> surface error
        raise RuntimeError("SentenceTransformer model unavailable and fallback not allowed")

    elif provider == "huggingface":
        return HuggingFaceEmbedding(
            api_key=kwargs.get("api_key", ""),
            model=kwargs.get("model", "sentence-transformers/all-MiniLM-L6-v2")
        )

    else:
        raise ValueError(
            f"Unknown embedding provider: {provider}\n"
            f"Supported providers: ollama, openai, sentence-transformers, huggingface"
        )
