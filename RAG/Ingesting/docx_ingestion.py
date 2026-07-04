from docx import Document
from pathlib import Path
from typing import List, Dict, Optional
import hashlib
import datetime


class DocxParser:
    def __init__(self, chunk_size: int = 1000, chunk_overlap: int = 200):
        """
        Args:
            chunk_size: maximum number of characters per chunk
            chunk_overlap: number of characters to overlap between chunks
        """
        self.chunk_size = chunk_size
        self.chunk_overlap = chunk_overlap

    def parse_docx(self, file_path: str) -> Dict:
        """
        Parse a DOCX file and extract content with metadata.

        Returns:
            dict: {
                'content': str,
                'paragraphs': List[str],
                'metadata': Dict[str, Optional[str]]
            }
        """
        path = Path(file_path)
        doc = Document(file_path)

        # Extract paragraphs (skip empty paragraphs)
        paragraphs: List[str] = []
        for para in doc.paragraphs:
            text = para.text.strip()
            if text:
                paragraphs.append(text)

        # Join paragraphs with double newline to preserve paragraph boundaries
        content = "\n\n".join(paragraphs)

        # Core properties may be None
        core_props = doc.core_properties
        created = None
        modified = None
        try:
            if core_props.created:
                created = core_props.created.isoformat()
        except Exception:
            # Some versions or properties may not be datetime; convert to str as fallback
            created = str(core_props.created) if core_props.created else None

        try:
            if core_props.modified:
                modified = core_props.modified.isoformat()
        except Exception:
            modified = str(core_props.modified) if core_props.modified else None

        metadata = {
            "filename": path.name,
            "file_path": str(path.absolute()),
            "title": core_props.title or "Untitled",
            "author": core_props.author or "Unknown",
            "created": created,
            "modified": modified,
        }

        return {"content": content, "paragraphs": paragraphs, "metadata": metadata}

    def chunk_text(self, text: str) -> List[Dict]:
        """
        Split text into overlapping chunks.

        The method prioritizes:
          1. Paragraph boundaries ('\n\n')
          2. Sentence endings ('.', '!', '?')
          3. Word boundaries (whitespace)

        Returns:
            List of chunk dictionaries with metadata:
            {
                'chunk_id': int,
                'text': str,
                'start_char': int,
                'end_char': int,
                'chunk_length': int
            }
        """
        chunks: List[Dict] = []
        start = 0
        chunk_id = 0
        text_length = len(text)

        sentence_end_chars = {".", "!", "?"}

        while start < text_length:
            # Default end (cap at text_length)
            end = min(start + self.chunk_size, text_length)

            if end < text_length:
                # Search for paragraph break within the overlap zone
                search_start = max(start, end - self.chunk_overlap)
                found = False

                # Look for paragraph break '\n\n'
                for i in range(end, search_start - 1, -1):
                    if text[i : i + 2] == "\n\n":
                        end = i + 2  # include the paragraph break
                        found = True
                        break

                # Look for sentence ending if no paragraph break found
                if not found:
                    sentence_search_start = max(start, end - max(100, self.chunk_overlap // 2))
                    for i in range(end - 1, sentence_search_start - 1, -1):
                        if text[i] in sentence_end_chars:
                            end = i + 1  # include the sentence-ending punctuation
                            found = True
                            break

                # Finally, fallback to word boundary (whitespace)
                if not found:
                    for i in range(end - 1, search_start - 1, -1):
                        if text[i].isspace():
                            end = i
                            found = True
                            break

                # If still not found, keep the original end (hard cut)
            # Extract the chunk text
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

            # Advance start; if not at end, move back by overlap
            if end >= text_length:
                break
            start = end - self.chunk_overlap if end < text_length else end

        return chunks

    def process_document(self, file_path: str) -> List[Dict]:
        """
        Complete pipeline: parse DOCX and create chunks, attaching document metadata to each chunk.

        Returns:
            List[Dict]: chunks with attached 'document_metadata' and 'document_id'
        """
        # Parse the document
        doc_data = self.parse_docx(file_path)

        # Chunk the full content
        chunks = self.chunk_text(doc_data["content"])

        # Add document metadata to each chunk
        document_id = self._generate_doc_id(str(file_path))
        for chunk in chunks:
            chunk["document_metadata"] = doc_data["metadata"]
            chunk["document_id"] = document_id

        return chunks

    def _generate_doc_id(self, file_path: str) -> str:
        """Generate a deterministic document ID (MD5 of the file path)."""
        return hashlib.md5(file_path.encode("utf-8")).hexdigest()


if __name__ == "__main__":
    # Ensure you have a file named 'Sample.docx' in the same directory!
    parser = DocxParser(chunk_size=1000, chunk_overlap=200)

    # Process the document and get our chunks
    chunks = parser.process_document("Sample.docx")

    # Print the results for inspection
    for chunk in chunks:
        print(f"---- Chunk {chunk['chunk_id']} (Length: {chunk['chunk_length']}) ----")
        print(chunk["text"])
        print(f"Source: {chunk['document_metadata']['title']} (ID: {chunk['document_id']})")
        print("________________________________________________________________\n")
