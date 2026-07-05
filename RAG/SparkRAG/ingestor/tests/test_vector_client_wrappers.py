import os
import sys

sys.path.insert(0, os.path.dirname(os.path.dirname(__file__)))

from vectordb import get_vectordb_client
from vector_client.cromadb import ChromaClientAdapter
from vector_client.quadrant import QdrantClientAdapter


def test_factory_returns_client_adapter_classes():
    chroma = get_vectordb_client("chroma", url="http://localhost:8000", collection="test")
    qdrant = get_vectordb_client("qdrant", url="http://localhost:6333", collection="test")

    assert isinstance(chroma, ChromaClientAdapter)
    assert isinstance(qdrant, QdrantClientAdapter)
