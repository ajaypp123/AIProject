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

logger = logging.getLogger(__name__)


class EmbeddingProvider:
    """Base interface for embedding providers."""
    def __init__(self):
        pass

    def provision(self, model: str, api_key: str, batch_size: int,
                     base_url: str):
        self.model = model
        self.api_key = api_key
        self.batch_size = batch_size
        self.base_url = base_url
        logger.info(f"Initialized OllamaEmbedding with model {self.model}")

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
            raise Exception(f"Ollama embedding failed: {e}")

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
    # model: str = "text-embedding-3-small"

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
            raise Exception(f"Ollama embedding failed: {e}")

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
            raise Exception(f"OpenAI batch embedding failed: {e}")


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

    def provision(self, model: str, api_key: str, batch_size: int,
                     base_url: str):
        # model: str = "sentence-transformers/all-MiniLM-L6-v2"
        self.model = model
        self.api_key = api_key
        self.batch_size = batch_size
        self.base_url = base_url
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
            raise Exception(msg)

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
            raise Exception(f"SentenceTransformer embedding failed: {e}")

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
            raise Exception(f"SentenceTransformer batch embedding failed: {e}")


class HuggingFaceEmbedding(EmbeddingProvider):
    """
    HuggingFace Inference API embeddings (cloud-based).

    Uses HuggingFace Inference API for on-demand embeddings.
    Supports any model from HuggingFace hub.

    Requires HF_TOKEN environment variable or token parameter.
    """

    # model: str = "sentence-transformers/all-MiniLM-L6-v2"
    # self.base_url = f"https://api-inference.huggingface.co/models/{model}"

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
                self.base_url,
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
            raise Exception(f"HuggingFace embedding failed: {e}")

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
                self.base_url,
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
    providerObj = None

    if provider == "ollama":
        providerObj = OllamaEmbedding()

    elif provider == "openai":
        providerObj = OpenAIEmbedding()

    elif provider in ["sentence-transformers", "local"]:
        providerObj = SentenceTransformerEmbedding()

    elif provider == "huggingface":
        providerObj = HuggingFaceEmbedding()

    else:
        raise ValueError(
            f"Unknown embedding provider: {provider}\n"
            f"Supported providers: ollama, openai, sentence-transformers, huggingface"
        )

    providerObj.provision(
            model=kwargs.get("model", "nomic-embed-text"),
            api_key=kwargs.get("api_key", ""),
            batch_size=kwargs.get("batch_size", 32),
            base_url=kwargs.get("base_url", "http://localhost:11434")
        )
    return providerObj
