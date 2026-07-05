import logging
from typing import Any, Dict, List, Optional
from urllib.parse import urlparse

import chromadb

logger = logging.getLogger(__name__)


class ChromaClientAdapter:
    def __init__(self, url: str, collection: str, api_key: Optional[str] = None):
        self.url = url.rstrip("/") if url else "http://localhost:8000"
        self.collection_name = collection
        self.api_key = api_key
        parsed = urlparse(self.url)
        host = parsed.hostname or "localhost"
        port = parsed.port or 8000
        self.client = chromadb.HttpClient(host=host, port=port, ssl=False)
        self.collection = self.client.get_or_create_collection(
            name=collection,
            metadata={"hnsw:space": "cosine"},
        )

    def store_chunks(self, chunks: List[Dict[str, Any]]) -> bool:
        try:
            ids = [chunk["id"] for chunk in chunks]
            embeddings = [chunk.get("embedding") for chunk in chunks]
            documents = [chunk.get("content", "") for chunk in chunks]
            metadatas = [
                {
                    "document": chunk.get("document_id"),
                    "source": chunk.get("source"),
                    "document_type": chunk.get("document_type"),
                }
                for chunk in chunks
            ]

            self.collection.add(
                ids=ids,
                embeddings=embeddings,
                documents=documents,
                metadatas=metadatas,
            )
            logger.info("Stored %s chunks to Chroma", len(chunks))
            return True
        except Exception as exc:  # pragma: no cover - defensive path
            logger.error("Failed to store chunks to Chroma: %s", exc)
            return False

    def delete_document(self, doc_id: str) -> bool:
        try:
            self.collection.delete(where={"document": {"$eq": doc_id}})
            logger.info("Deleted document %s from Chroma", doc_id)
            return True
        except Exception as exc:  # pragma: no cover - defensive path
            logger.error("Failed to delete document from Chroma: %s", exc)
            return False

    def health_check(self) -> bool:
        try:
            return self.client.heartbeat() is not None
        except Exception as exc:  # pragma: no cover - defensive path
            logger.error("Chroma health check failed: %s", exc)
            return False

    def search(self, query_text: str, k: int = 3, where: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        try:
            results = self.collection.get(include=["documents", "metadatas"])
            documents = results.get("documents", []) or []
            metadatas = results.get("metadatas", []) or []
            ids = results.get("ids", []) or []

            filtered_results = []
            for idx, document in enumerate(documents):
                metadata = metadatas[idx] if idx < len(metadatas) else {}
                if where and not self._matches_where(metadata, where):
                    continue
                if query_text and query_text.lower() not in document.lower():
                    continue
                filtered_results.append(
                    {"id": ids[idx] if idx < len(ids) else None, "document": document, "metadata": metadata}
                )

            filtered_results = filtered_results[:k]
            return {"results": filtered_results}
        except Exception as exc:  # pragma: no cover - defensive path
            logger.error("Chroma search failed: %s", exc)
            return {"results": []}

    @staticmethod
    def _matches_where(metadata: Dict[str, Any], where: Dict[str, Any]) -> bool:
        for key, expected in where.items():
            if metadata.get(key) != expected:
                return False
        return True


class ChromaLocalClientAdapter:
    def __init__(self, url: str, collection: str, api_key: Optional[str] = None):
        self.url = url.rstrip("/") if url else "."
        self.collection_name = collection
        self.api_key = api_key
        database_dir = self.url or "."
        client = chromadb.PersistentClient(path=database_dir)
        self.collection = client.get_or_create_collection(
            name=collection,
            metadata={"hnsw:space": "cosine"},
        )

    def store_chunks(self, chunks: List[Dict[str, Any]]) -> bool:
        try:
            self.collection.upsert(
                ids=[chunk["id"] for chunk in chunks],
                embeddings=[chunk.get("embedding") for chunk in chunks],
                documents=[chunk.get("content", "") for chunk in chunks],
                metadatas=[
                    {
                        "document": chunk.get("document_id"),
                        "source": chunk.get("source"),
                        "document_type": chunk.get("document_type"),
                    }
                    for chunk in chunks
                ],
            )
            logger.info("Stored %s chunks locally to Chroma", len(chunks))
            return True
        except Exception as exc:  # pragma: no cover - defensive path
            logger.error("Failed to store chunks locally: %s", exc)
            return False

    def delete_document(self, doc_id: str) -> bool:
        try:
            self.collection.delete(where={"document": {"$eq": doc_id}})
            logger.info("Deleted document %s from Chroma local client", doc_id)
            return True
        except Exception as exc:  # pragma: no cover - defensive path
            logger.error("Failed to delete document locally: %s", exc)
            return False

    def health_check(self) -> bool:
        return True

    def search(self, query_text: str, k: int = 3, where: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        try:
            results = self.collection.get(include=["documents", "metadatas"])
            documents = results.get("documents", []) or []
            metadatas = results.get("metadatas", []) or []
            ids = results.get("ids", []) or []

            filtered_results = []
            for idx, document in enumerate(documents):
                metadata = metadatas[idx] if idx < len(metadatas) else {}
                if where and not self._matches_where(metadata, where):
                    continue
                if query_text and query_text.lower() not in document.lower():
                    continue
                filtered_results.append(
                    {"id": ids[idx] if idx < len(ids) else None, "document": document, "metadata": metadata}
                )

            filtered_results = filtered_results[:k]
            return {"results": filtered_results}
        except Exception as exc:  # pragma: no cover - defensive path
            logger.error("Chroma local search failed: %s", exc)
            return {"results": []}

    @staticmethod
    def _matches_where(metadata: Dict[str, Any], where: Dict[str, Any]) -> bool:
        for key, expected in where.items():
            if metadata.get(key) != expected:
                return False
        return True
