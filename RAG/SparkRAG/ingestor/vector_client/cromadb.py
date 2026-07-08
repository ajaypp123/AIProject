import logging
import os
from pathlib import Path
from typing import Any, Dict, List, Optional
from urllib.parse import urlparse

import chromadb

logger = logging.getLogger(__name__)

class ChromaLocalClientAdapterInterface:

    def _create_collection(self, collection, client):
        return client.get_or_create_collection(
            name=collection,
            #embedding_function=embedding_fn,
            metadata={"hnsw:space": "cosine"}  # cosine is common for semantic search
        )

    def store_chunks(self, chunks: List[Dict]) -> bool:
        try:
            ids = [chunk["id"] for chunk in chunks]
            embeddings = [chunk.get("embedding") for chunk in chunks]
            documents = [chunk.get("content", "") for chunk in chunks]
            metadatas = [{
                "document": chunk.get("document_id"),
                "source": chunk.get("source"),
                "document_type": chunk.get("document_type"),
            } for chunk in chunks]

            self.collection.upsert(
                ids=ids,
                embeddings=embeddings,
                documents=documents,
                metadatas=metadatas,
            )
            logger.info(f"Stored {len(chunks)} chunks locally to ChromaLocalClient")
            return True
        except Exception as e:
            logger.error(f"Failed to store chunks locally: {e}")
            return False

    def delete_document(self, doc_id: str) -> bool:
        try:
            self.collection.delete(where={"document": {"$eq": doc_id}})
            logger.info(f"Deleted document {doc_id} from ChromaLocalClient")
            return True
        except Exception as e:
            logger.error(f"Failed to delete document locally: {e}")
            return False

    def health_check(self) -> bool:
        return True

    def search(self, query_text: str, k: int = 3, where: dict = None) -> Dict[str, Any]:
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
                filtered_results.append({
                    "id": ids[idx] if idx < len(ids) else None,
                    "document": document,
                    "metadata": metadata
                })

            filtered_results = filtered_results[:k]
            return {"results": filtered_results}
        except Exception as e:
            logger.error(f"Chroma local search failed: {e}")
            return {"results": []}

    @staticmethod
    def _matches_where(metadata: Dict[str, Any], where: Dict[str, Any]) -> bool:
        for key, expected in where.items():
            if metadata.get(key) != expected:
                return False
        return True


class ChromaClientAdapter(ChromaLocalClientAdapterInterface):
    def __init__(self, url: str, collection: str, api_key: Optional[str] = None, batch_size = 100):
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


class ChromaLocalClientAdapter(ChromaLocalClientAdapterInterface):
    def __init__(self, url: str, collection: str, api_key: Optional[str] = None, batch_size: int = 100):
        croma_db_location = os.getenv("CROMA_DB_LOCATION", url)
        DB_DIR = Path(croma_db_location)
        DB_DIR.mkdir(exist_ok=True)
        # loads existing DB if present
        client = chromadb.PersistentClient(path=str(DB_DIR))
        self.collection = self._create_collection(collection, client)
