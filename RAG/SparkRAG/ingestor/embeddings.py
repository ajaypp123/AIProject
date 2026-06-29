import logging
import requests
import json
from typing import List, Optional

logger = logging.getLogger(__name__)

class EmbeddingProvider:
    def embed(self, text: str) -> Optional[List[float]]:
        raise NotImplementedError
    
    def embed_batch(self, texts: List[str]) -> Optional[List[List[float]]]:
        raise NotImplementedError

class OllamaEmbedding(EmbeddingProvider):
    def __init__(self, base_url: str, model: str):
        self.base_url = base_url
        self.model = model
    
    def embed(self, text: str) -> Optional[List[float]]:
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
        embeddings = []
        for text in texts:
            emb = self.embed(text)
            if emb is None:
                return None
            embeddings.append(emb)
        return embeddings

class OpenAIEmbedding(EmbeddingProvider):
    def __init__(self, api_key: str, model: str = "text-embedding-3-small"):
        self.api_key = api_key
        self.model = model
    
    def embed(self, text: str) -> Optional[List[float]]:
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
        try:
            response = requests.post(
                "https://api.openai.com/v1/embeddings",
                json={"input": texts, "model": self.model},
                headers={"Authorization": f"Bearer {self.api_key}"},
                timeout=60
            )
            response.raise_for_status()
            data = response.json()
            embeddings = [item["embedding"] for item in sorted(data["data"], key=lambda x: x["index"])]
            return embeddings
        except Exception as e:
            logger.error(f"OpenAI batch embedding failed: {e}")
            return None

class LocalEmbedding(EmbeddingProvider):
    def __init__(self, model: str = "all-MiniLM-L6-v2"):
        try:
            from sentence_transformers import SentenceTransformer
            self.model = SentenceTransformer(model)
            self.embedding_dim = self.model.get_sentence_embedding_dimension()
        except Exception as e:
            logger.error(f"Failed to load local embedding model: {e}")
            self.model = None
    
    def embed(self, text: str) -> Optional[List[float]]:
        if self.model is None:
            return None
        try:
            embedding = self.model.encode(text)
            return embedding.tolist()
        except Exception as e:
            logger.error(f"Local embedding failed: {e}")
            return None
    
    def embed_batch(self, texts: List[str]) -> Optional[List[List[float]]]:
        if self.model is None:
            return None
        try:
            embeddings = self.model.encode(texts)
            return embeddings.tolist()
        except Exception as e:
            logger.error(f"Local batch embedding failed: {e}")
            return None

def get_embedding_provider(provider: str, **kwargs) -> EmbeddingProvider:
    if provider == "ollama":
        return OllamaEmbedding(kwargs.get("base_url", "http://localhost:11434"), kwargs.get("model", "nomic-embed-text"))
    elif provider == "openai":
        return OpenAIEmbedding(kwargs.get("api_key", ""), kwargs.get("model", "text-embedding-3-small"))
    elif provider == "local":
        return LocalEmbedding(kwargs.get("model", "all-MiniLM-L6-v2"))
    else:
        raise ValueError(f"Unknown embedding provider: {provider}")
