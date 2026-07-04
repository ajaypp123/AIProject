# main.py
import hashlib
from pathlib import Path
from typing import List, Dict


class TextDocumentParser:
    """Parse text files for RAG system ingestion."""

    def __init__(self, chunk_size: int = 1000, chunk_overlap: int = 200):
        """
        Args:
            chunk_size: approximate maximum number of characters per chunk.
            chunk_overlap: number of characters to overlap between consecutive chunks.
        """
        self.chunk_size = chunk_size
        self.chunk_overlap = chunk_overlap

    def parse_file(self, file_path: str) -> Dict:
        """
        Read a UTF-8 text file and return its content and metadata.

        Returns:
            A dict with 'content' (the file text) and 'metadata' (filename, path, size, extension,
            document_id, char_count, and word_count).
        """
        path = Path(file_path)
        with path.open("r", encoding="utf-8") as f:
            content = f.read()

        metadata = {
            "filename": path.name,
            "file_path": str(path.resolve()),
            "file_size": path.stat().st_size,
            "file_extension": path.suffix,
            "document_id": self._generate_doc_id(str(path.resolve())),
            "char_count": len(content),
            "word_count": len(content.split()),
        }

        return {"content": content, "metadata": metadata}

    def _generate_doc_id(self, file_path: str) -> str:
        """Generate a stable unique document ID based on the absolute file path."""
        return hashlib.md5(file_path.encode("utf-8")).hexdigest()

    def chunk_text(self, text: str) -> List[Dict]:
        """
        Split text into overlapping chunks for RAG processing.

        Algorithm:
        - Start at position 0.
        - Set an initial end = start + chunk_size.
        - If the end is not at the end of text, search backwards from end for a sentence boundary
          (one of '.', '!', '?', or a newline). If found within a reasonable backtrack window,
          break there (include the punctuation/newline in the chunk where appropriate).
        - If no sentence boundary is found, search backwards for a whitespace (word boundary).
        - Extract the chunk, strip whitespace, append to list if non-empty.
        - Advance start to (end - chunk_overlap) to keep an overlap between chunks.
        - Guard against zero-length progress (when no suitable break is found) by forcing forward
          progress up to chunk_size to avoid infinite loops.
        """
        if not text:
            return []

        chunks: List[Dict] = []
        start = 0
        chunk_id = 0
        length = len(text)
        sentence_ends = {".", "!", "?", "\n"}

        while start < length:
            end = min(start + self.chunk_size, length)

            # When we're not at the end, prefer a sentence boundary inside a lookback window
            if end < length:
                best_break = end
                lookback_sentence = max(start, end - 100)  # search up to 100 chars back for sentence end
                for i in range(end - 1, lookback_sentence - 1, -1):
                    if text[i] in sentence_ends:
                        # include the sentence terminator in the chunk
                        best_break = i + 1
                        break

                # If no sentence break found, search for a whitespace/word boundary within a smaller window
                if best_break == end:
                    lookback_word = max(start, end - 50)  # search up to 50 chars back for whitespace
                    for i in range(end - 1, lookback_word - 1, -1):
                        if text[i].isspace():
                            best_break = i
                            break

                end = best_break

                # Ensure we make forward progress; if no break was found and end equals start, advance by chunk_size
                if end <= start:
                    end = min(start + self.chunk_size, length)
                    if end <= start:
                        # nothing more to extract
                        break

            # Extract and store the chunk
            chunk_text = text[start:end].strip()
            if chunk_text:
                chunks.append(
                    {
                        "chunk_id": chunk_id,
                        "text": chunk_text,
                        "start_char": start,
                        "end_char": end,
                        "chunk_length": len(chunk_text),
                    }
                )
                chunk_id += 1

            # Advance start with overlap
            if end >= length:
                break
            start = max(0, end - self.chunk_overlap)

        return chunks

    def process_document(self, file_path: str) -> List[Dict]:
        """
        Complete pipeline: parse file, chunk content, and attach document metadata to each chunk.

        Returns:
            A list of chunk dicts, each augmented with 'document_metadata'.
        """
        doc_data = self.parse_file(file_path)
        chunks = self.chunk_text(doc_data["content"])

        for chunk in chunks:
            chunk["document_metadata"] = doc_data["metadata"]

        return chunks


if __name__ == "__main__":
    # Example usage
    parser = TextDocumentParser(chunk_size=500, chunk_overlap=100)

    sample_text = """Introduction to RAG Systems
Retrieval-Augmented Generation (RAG) is a powerful technique that augments a model's responses
with external documents by retrieving relevant content and conditioning generation on those results.

The Squirrel and the Wi-Fi Router
Once upon a time, in a quiet suburban neighborhood, there lived a squirrel named Nibbles who loved gadgets...
(Imagine a longer story here to demo chunking.)

How the chunking helps RAG (summary)
Preferencing sentence boundaries makes chunk contents more meaningful for semantic embeddings.
Falling back to word boundaries prevents cutting tokens mid-word and reduces noisy embeddings.
Controlled overlap preserves local context across chunk boundaries, improving relevance during retrieval.
Attaching document_metadata to each chunk enables tracing back results to original documents for provenance, filtering, or display.
"""

    sample_path = "sample_doc.txt"
    with open(sample_path, "w", encoding="utf-8") as f:
        f.write(sample_text)

    chunks = parser.process_document(sample_path)

    print(f"Document: {chunks[0]['document_metadata']['filename']}")
    print(f"Total chunks: {len(chunks)}")

    for chunk in chunks[:5]:
        print(f"\nChunk {chunk['chunk_id']}:")
        print(f"Length: {chunk['chunk_length']} chars")
        print(f"Text (preview): {chunk['text'][:150]}...")
