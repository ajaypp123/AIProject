import os
import sys

sys.path.insert(0, os.path.dirname(os.path.dirname(__file__)))

from vectordb import ChromaLocalClient, get_vectordb_client

def test_factory_returns_local_client_for_chroma_local_provider():
    client = get_vectordb_client("chroma-local", collection="test-local")
    assert isinstance(client, ChromaLocalClient)
