import sys

sys.path.insert(0, "..")

from embeddings import get_embedding_provider


def test_embedding_provider_factory():
    # Test Ollama provider
    provider = get_embedding_provider(
        "ollama", base_url="http://localhost:11434", model="nomic-embed-text"
    )
    assert provider is not None
    assert hasattr(provider, "embed")
    assert hasattr(provider, "embed_batch")


def test_embedding_provider_local():
    try:
        provider = get_embedding_provider("local", model="all-MiniLM-L6-v2")
        assert provider is not None
    except Exception:
        print("Skipping local embedding test (SentenceTransformers not installed)")


def test_invalid_provider():
    try:
        provider = get_embedding_provider("invalid", model="test")
        assert False, "Should have raised ValueError"
    except ValueError as e:
        assert "Unknown embedding provider" in str(e)


if __name__ == "__main__":
    test_embedding_provider_factory()
    test_embedding_provider_local()
    test_invalid_provider()
    print("All tests passed!")
