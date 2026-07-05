import logging
from typing import Any, Dict, List, Optional

from vector_client.cromadb import ChromaClientAdapter, ChromaLocalClientAdapter
from vector_client.quadrant import QdrantClientAdapter

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

        Response:
            Dict with id, document, metadata, distance
        """
        raise NotImplementedError


class QdrantClient(VectorDBClient):
    def __init__(self, url: str, collection: str, api_key: Optional[str] = None):
        super().__init__(url, collection, api_key)
        self._client = QdrantClientAdapter(url, collection, api_key)

    def store_chunks(self, chunks: List[Dict]) -> bool:
        return self._client.store_chunks(chunks)

    def delete_document(self, doc_id: str) -> bool:
        return self._client.delete_document(doc_id)

    def health_check(self) -> bool:
        return self._client.health_check()

    def search(self, query_text: str, k: int = 3, where: dict = None) -> Dict[str, Any]:
        return self._client.search(query_text, k, where)


class ChromaClient(VectorDBClient):
    def __init__(self, url: str, collection: str, api_key: Optional[str] = None):
        super().__init__(url, collection, api_key)
        self._client = ChromaClientAdapter(url, collection, api_key)

    def store_chunks(self, chunks: List[Dict]) -> bool:
        return self._client.store_chunks(chunks)

    def delete_document(self, doc_id: str) -> bool:
        return self._client.delete_document(doc_id)

    def health_check(self) -> bool:
        return self._client.health_check()

    def search(self, query_text: str, k: int = 3, where: dict = None) -> Dict[str, Any]:
        return self._client.search(query_text, k, where)


class ChromaLocalClient(VectorDBClient):
    def __init__(self, url: str, collection: str, api_key: Optional[str] = None):
        super().__init__(url, collection, api_key)
        self._client = ChromaLocalClientAdapter(url, collection, api_key)

    def store_chunks(self, chunks: List[Dict]) -> bool:
        return self._client.store_chunks(chunks)

    def delete_document(self, doc_id: str) -> bool:
        return self._client.delete_document(doc_id)

    def health_check(self) -> bool:
        return self._client.health_check()

    def search(self, query_text: str, k: int = 3, where: dict = None) -> Dict[str, Any]:
        return self._client.search(query_text, k, where)


def get_vectordb_client(provider: str, **kwargs) -> VectorDBClient:
    url = kwargs.get("url", "")
    collection = kwargs.get("collection", "spark-rag")
    api_key = kwargs.get("api_key")

    if provider == "qdrant":
        return QdrantClient(url, collection, api_key)
    if provider == "chroma":
        return ChromaClient(url, collection, api_key)
    if provider == "chroma-local":
        return ChromaLocalClient(url, collection, api_key)

    raise ValueError(f"Unknown vector DB provider: {provider}")
