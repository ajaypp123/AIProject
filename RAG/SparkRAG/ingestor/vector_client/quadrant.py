import logging
from typing import Any, Dict, List, Optional

from qdrant_client import QdrantClient
from qdrant_client.http.models import Distance, FieldCondition, Filter, MatchValue, PointStruct, VectorParams

logger = logging.getLogger(__name__)


class QdrantClientAdapter:
    def __init__(self, url: str, collection: str, api_key: Optional[str] = None):
        self.url = url.rstrip("/") if url else "http://localhost:6333"
        self.collection_name = collection
        self.api_key = api_key
        self.client = QdrantClient(url=self.url, api_key=api_key, timeout=30)

    def store_chunks(self, chunks: List[Dict[str, Any]]) -> bool:
        try:
            if not chunks:
                return True

            first_embedding = chunks[0].get("embedding") or []
            if not self._collection_exists():
                self.client.create_collection(
                    collection_name=self.collection_name,
                    vectors_config=VectorParams(size=len(first_embedding), distance=Distance.COSINE),
                )

            points = []
            for chunk in chunks:
                points.append(
                    PointStruct(
                        id=hash(chunk["id"]) & 0x7FFFFFFF,
                        vector=chunk.get("embedding") or [],
                        payload={
                            "document": chunk.get("document_id"),
                            "content": chunk.get("content"),
                            "source": chunk.get("source"),
                            "document_type": chunk.get("document_type"),
                            "chunk_id": chunk.get("id"),
                        },
                    )
                )

            self.client.upsert(collection_name=self.collection_name, points=points, wait=True)
            logger.info("Stored %s chunks to Qdrant", len(chunks))
            return True
        except Exception as exc:  # pragma: no cover - defensive path
            logger.error("Failed to store chunks to Qdrant: %s", exc)
            return False

    def delete_document(self, doc_id: str) -> bool:
        try:
            self.client.delete(
                collection_name=self.collection_name,
                points_selector=Filter(
                    must=[FieldCondition(key="document", match=MatchValue(value=doc_id))]
                ),
                wait=True,
            )
            logger.info("Deleted document %s from Qdrant", doc_id)
            return True
        except Exception as exc:  # pragma: no cover - defensive path
            logger.error("Failed to delete document from Qdrant: %s", exc)
            return False

    def health_check(self) -> bool:
        try:
            self.client.get_collections()
            return True
        except Exception as exc:  # pragma: no cover - defensive path
            logger.error("Qdrant health check failed: %s", exc)
            return False

    def search(self, query_text: str, k: int = 3, where: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        try:
            points, _ = self.client.scroll(
                collection_name=self.collection_name,
                limit=100,
                with_payload=True,
                with_vectors=False,
            )

            filtered_results = []
            for point in points:
                payload = point.payload or {}
                document = payload.get("content", "") or ""
                metadata = {
                    "document": payload.get("document"),
                    "source": payload.get("source"),
                    "document_type": payload.get("document_type"),
                    "chunk_id": payload.get("chunk_id"),
                }
                if where and not self._matches_where(metadata, where):
                    continue
                if query_text and query_text.lower() not in document.lower():
                    continue
                filtered_results.append({"id": point.id, "document": document, "metadata": metadata})

            filtered_results = filtered_results[:k]
            return {"results": filtered_results}
        except Exception as exc:  # pragma: no cover - defensive path
            logger.error("Qdrant search failed: %s", exc)
            return {"results": []}

    def _collection_exists(self) -> bool:
        try:
            return self.client.collection_exists(self.collection_name)
        except Exception:
            return False

    @staticmethod
    def _matches_where(metadata: Dict[str, Any], where: Dict[str, Any]) -> bool:
        for key, expected in where.items():
            if metadata.get(key) != expected:
                return False
        return True
