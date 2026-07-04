import logging
import os
from pathlib import Path
import requests
import chromadb
from typing import List, Dict, Optional, Any

logger = logging.getLogger(__name__)

class VectorDBClient:

    def __init__(self, url, collection, api_key):
        self.url = url
        self.api_key = api_key
        self.collection = collection

    def store_chunks(self, chunks: List[Dict]) -> bool:
        raise NotImplementedError

    def delete_document(self, doc_id: str) -> bool:
        raise NotImplementedError

    def health_check(self) -> bool:
        raise NotImplementedError

    def search(self, query_text: str, k: int = 3, where: dict = None) -> Dict:
        """
        Query the collection.
        - query_text: string to search
        - k: number of results to return
        - where: optional metadata filter dict, e.g. {"source": "beowulf.txt"}

        Responce:
            Dict with id, document, metadta, distance
        """
        raise NotImplementedError

class QdrantClient(VectorDBClient):
    def __init__(self, url: str, collection: str, api_key: Optional[str] = None):
        self.url = url.rstrip('/')
        self.collection = collection
        self.api_key = api_key
        self.headers = {}
        if api_key:
            self.headers["api-key"] = api_key

    def store_chunks(self, chunks: List[Dict]) -> bool:
        try:
            points = []
            for chunk in chunks:
                point = {
                    "id": hash(chunk["id"]) & 0x7fffffff,
                    "vector": chunk["embedding"],
                    "payload": {
                        "document": chunk.get("document_id"),
                        "content": chunk.get("content"),
                        "source": chunk.get("source"),
                        "document_type": chunk.get("document_type"),
                        "chunk_id": chunk.get("id"),
                    }
                }
                points.append(point)

            response = requests.put(
                f"{self.url}/collections/{self.collection}/points?wait=true",
                json={"points": points},
                headers=self.headers,
                timeout=60
            )
            response.raise_for_status()
            logger.info(f"Stored {len(chunks)} chunks to Qdrant")
            return True
        except Exception as e:
            logger.error(f"Failed to store chunks to Qdrant: {e}")
            return False

    def delete_document(self, doc_id: str) -> bool:
        try:
            response = requests.post(
                f"{self.url}/collections/{self.collection}/points/delete",
                json={
                    "filter": {
                        "must": [
                            {
                                "key": "document",
                                "match": {
                                    "value": doc_id
                                }
                            }
                        ]
                    }
                },
                headers=self.headers,
                timeout=30
            )
            response.raise_for_status()
            logger.info(f"Deleted document {doc_id} from Qdrant")
            return True
        except Exception as e:
            logger.error(f"Failed to delete document from Qdrant: {e}")
            return False

    def health_check(self) -> bool:
        try:
            response = requests.get(f"{self.url}/health", headers=self.headers, timeout=10)
            return response.status_code == 200
        except Exception as e:
            logger.error(f"Qdrant health check failed: {e}")
            return False

    def search(self, query_text: str, k: int = 3, where: dict = None) -> Dict[str, Any]:
        try:
            response = requests.post(
                f"{self.url}/collections/{self.collection}/points/scroll",
                json={
                    "limit": 100,
                    "with_payload": True,
                    "with_vector": False,
                },
                headers=self.headers,
                timeout=30,
            )
            response.raise_for_status()
            points = response.json().get("result", {}).get("points", [])

            filtered_results = []
            for point in points:
                payload = point.get("payload", {}) or {}
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
                filtered_results.append({
                    "id": point.get("id"),
                    "document": document,
                    "metadata": metadata
                })

            filtered_results = filtered_results[:k]
            return {"results": filtered_results}
        except Exception as e:
            logger.error(f"Qdrant search failed: {e}")
            return {"results": []}

    @staticmethod
    def _matches_where(metadata: Dict[str, Any], where: Dict[str, Any]) -> bool:
        for key, expected in where.items():
            if metadata.get(key) != expected:
                return False
        return True

class ChromaClient(VectorDBClient):
    def __init__(self, url: str, collection: str, api_key: Optional[str] = None):
        self.url = url.rstrip('/')
        self.collection = collection
        self.api_key = api_key
        self.headers = {"Content-Type": "application/json"}
        if api_key:
            self.headers["Authorization"] = f"Bearer {api_key}"

    def store_chunks(self, chunks: List[Dict]) -> bool:
        try:
            ids = []
            embeddings = []
            documents = []
            metadatas = []

            for chunk in chunks:
                ids.append(chunk["id"])
                embeddings.append(chunk.get("embedding"))
                documents.append(chunk.get("content", ""))
                metadatas.append({
                    "document": chunk.get("document_id"),
                    "source": chunk.get("source"),
                    "document_type": chunk.get("document_type"),
                })

            response = requests.post(
                f"{self.url}/api/v1/collections/{self.collection}/add",
                json={
                    "ids": ids,
                    "embeddings": embeddings,
                    "documents": documents,
                    "metadatas": metadatas,
                },
                headers=self.headers,
                timeout=60,
            )
            response.raise_for_status()
            logger.info(f"Stored {len(chunks)} chunks to Chroma at {self.url}")
            return True
        except Exception as e:
            logger.error(f"Failed to store chunks to Chroma: {e}")
            return False

    def delete_document(self, doc_id: str) -> bool:
        try:
            response = requests.post(
                f"{self.url}/api/v1/collections/{self.collection}/delete",
                json={
                    "where": {
                        "document": {"$eq": doc_id}
                    }
                },
                headers=self.headers,
                timeout=30,
            )
            response.raise_for_status()
            logger.info(f"Deleted document {doc_id} from Chroma at {self.url}")
            return True
        except Exception as e:
            logger.error(f"Failed to delete document from Chroma: {e}")
            return False

    def health_check(self) -> bool:
        try:
            response = requests.get(f"{self.url}/api/v1/heartbeat", headers=self.headers, timeout=10)
            return response.status_code == 200
        except Exception as e:
            logger.error(f"Chroma health check failed: {e}")
            return False

    def search(self, query_text: str, k: int = 3, where: dict = None) -> Dict[str, Any]:
        try:
            payload = {
                "limit": 100,
                "include": ["documents", "metadatas"],
            }
            if where:
                payload["where"] = where
            response = requests.post(
                f"{self.url}/api/v1/collections/{self.collection}/get",
                json=payload,
                headers=self.headers,
                timeout=30,
            )
            response.raise_for_status()
            data = response.json()

            documents = data.get("documents", []) or []
            metadatas = data.get("metadatas", []) or []
            ids = data.get("ids", []) or []

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
            logger.error(f"Chroma search failed: {e}")
            return {"results": []}

    @staticmethod
    def _matches_where(metadata: Dict[str, Any], where: Dict[str, Any]) -> bool:
        for key, expected in where.items():
            if metadata.get(key) != expected:
                return False
        return True

class ChromaLocalClient(VectorDBClient):
    def __init__(self, url: str, collection: str, api_key: Optional[str] = None):
        self._store = []
        # --------- (A) Choose persistence location ---------
        croma_db_location = os.getenv("CROMA_DB_LOCATION", url)
        DB_DIR = Path(croma_db_location)
        DB_DIR.mkdir(exist_ok=True)
        # loads existing DB if present
        client = chromadb.PersistentClient(path=str(DB_DIR))
        self.collection = self._create_collection(collection, client)

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

def get_vectordb_client(provider: str, **kwargs) -> VectorDBClient:
    url = kwargs.get("url", "")
    collection = kwargs.get("collection", "spark-rag")
    api_key = kwargs.get("api_key")

    if provider == "qdrant":
        return QdrantClient(url, collection, api_key)
    elif provider == "chroma":
        return ChromaClient(url, collection, api_key)
    elif provider == "chroma-local":
        return ChromaLocalClient(url, collection, api_key)
    else:
        raise ValueError(f"Unknown vector DB provider: {provider}")
