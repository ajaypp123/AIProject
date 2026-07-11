import sys

sys.path.insert(0, "..")

from chunking import Chunker


def test_chunk_text_basic():
    chunker = Chunker(chunk_size=100, chunk_overlap=10)
    text = "This is a test paragraph.\n\nThis is another paragraph.\n\nAnd a third one."

    chunks = chunker.chunk_text(text)
    assert len(chunks) > 0
    assert all(isinstance(chunk, str) for chunk in chunks)


def test_chunk_text_empty():
    chunker = Chunker()
    chunks = chunker.chunk_text("")
    assert len(chunks) == 0


def test_chunk_markdown():
    chunker = Chunker(chunk_size=200, chunk_overlap=20)
    text = """# Title
    
## Section 1
Content for section 1.

## Section 2
Content for section 2."""

    chunks = chunker.chunk_markdown(text)
    assert len(chunks) > 0
    assert any("Title" in chunk for chunk in chunks)


def test_chunk_long_text():
    chunker = Chunker(chunk_size=50, chunk_overlap=5)
    text = "word " * 100

    chunks = chunker.chunk_text(text)
    assert len(chunks) > 0
    for chunk in chunks:
        assert len(chunk) <= 60  # Allow some buffer


if __name__ == "__main__":
    test_chunk_text_basic()
    test_chunk_text_empty()
    test_chunk_markdown()
    test_chunk_long_text()
    print("All tests passed!")
