import logging
import requests
import json
from typing import List, Dict, Optional

logger = logging.getLogger(__name__)

class VectorDBClient:
    def store_chunks(self, chunks: List[Dict]) -> bool:
        raise NotImplementedError
    
    def delete_document(self, doc_id: str) -> bool:
        raise NotImplementedError
    
    def health_check(self) -> bool:
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

class ChromaClient(VectorDBClient):
    def __init__(self, url: str, collection: str, api_key: Optional[str] = None):
        self.url = url.rstrip('/')
        self.collection = collection
        self.api_key = api_key
        self.headers = {}
        if api_key:
            self.headers["Authorization"] = f"Bearer {api_key}"
    
    def store_chunks(self, chunks: List[Dict]) -> bool:
        try:
            ids = []
            embeddings = []
            metadatas = []
            documents = []
            
            for chunk in chunks:
                ids.append(chunk["id"])
                embeddings.append(chunk["embedding"])
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
                    "metadatas": metadatas
                },
                headers=self.headers,
                timeout=60
            )
            response.raise_for_status()
            logger.info(f"Stored {len(chunks)} chunks to Chroma")
            return True
        except Exception as e:
            logger.error(f"Failed to store chunks to Chroma: {e}")
            return False
    
    def delete_document(self, doc_id: str) -> bool:
        try:
            response = requests.delete(
                f"{self.url}/api/v1/collections/{self.collection}/delete",
                json={
                    "where": {
                        "document": {
                            "$eq": doc_id
                        }
                    }
                },
                headers=self.headers,
                timeout=30
            )
            response.raise_for_status()
            logger.info(f"Deleted document {doc_id} from Chroma")
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

def get_vectordb_client(provider: str, **kwargs) -> VectorDBClient:
    url = kwargs.get("url", "")
    collection = kwargs.get("collection", "spark-rag")
    api_key = kwargs.get("api_key")
    
    if provider == "qdrant":
        return QdrantClient(url, collection, api_key)
    elif provider == "chroma":
        return ChromaClient(url, collection, api_key)
    else:
        raise ValueError(f"Unknown vector DB provider: {provider}")
