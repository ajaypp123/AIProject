import os
import sys

sys.path.insert(0, os.path.dirname(os.path.dirname(__file__)))

from vectordb import ChromaLocalClient, get_vectordb_client

def test_chroma_local_client_stores_searches_and_deletes_chunks():
    client = ChromaLocalClient(collection="test-local")

    assert client.health_check() is True

    chunks = [
        {
            "id": "chunk-1",
            "embedding": [0.1, 0.2, 0.3],
            "document_id": "doc-1",
            "content": "hello world",
            "source": "sample.md",
            "document_type": "markdown",
        }
    ]

    assert client.store_chunks(chunks) is True
    assert client.collection.count() == 1

    search_result = client.search("hello", k=3)
    assert search_result["results"]
    assert any("hello" in item.get("document", "") for item in search_result["results"])

    assert client.delete_document("doc-1") is True
    assert client.collection.count() == 0


def test_factory_returns_local_client_for_chroma_local_provider():
    client = get_vectordb_client("chroma-local", collection="test-local")
    assert isinstance(client, ChromaLocalClient)
