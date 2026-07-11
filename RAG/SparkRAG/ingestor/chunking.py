import re
from typing import List


class Chunker:
    def __init__(self, chunk_size: int = 512, chunk_overlap: int = 50):
        self.chunk_size = chunk_size
        self.chunk_overlap = chunk_overlap

    def chunk_text(self, text: str) -> List[str]:
        if not text or len(text) == 0:
            return []

        chunks = []

        paragraphs = text.split("\n\n")

        current_chunk = ""
        for para in paragraphs:
            if len(current_chunk) + len(para) + 2 <= self.chunk_size:
                current_chunk += para + "\n\n"
            else:
                if current_chunk:
                    chunks.append(current_chunk.strip())

                if len(para) > self.chunk_size:
                    sub_chunks = self._chunk_long_paragraph(para)
                    chunks.extend(sub_chunks)
                    current_chunk = ""
                else:
                    current_chunk = para + "\n\n"

        if current_chunk.strip():
            chunks.append(current_chunk.strip())

        overlapped_chunks = []
        for i, chunk in enumerate(chunks):
            if i > 0:
                prev_chunk = chunks[i - 1]
                overlap_text = (
                    prev_chunk[-self.chunk_overlap :]
                    if len(prev_chunk) > self.chunk_overlap
                    else prev_chunk
                )
                chunk = overlap_text + "\n" + chunk

            overlapped_chunks.append(chunk)

        return overlapped_chunks

    def _chunk_long_paragraph(self, text: str) -> List[str]:
        # First try to split by sentence boundaries
        sentences = re.split(r"(?<=[.!?])\s+", text)
        chunks = []

        # If there are no sentence boundaries (or a single long sentence),
        # fall back to word-based splitting to ensure chunk_size is respected.
        if len(sentences) <= 1:
            words = text.split()
            current = ""
            for w in words:
                if len(current) + len(w) + 1 <= self.chunk_size:
                    current = (current + " " + w).strip()
                else:
                    if current:
                        chunks.append(current)
                    current = w
            if current:
                chunks.append(current)
            return chunks

        current_chunk = ""
        for sentence in sentences:
            if len(sentence) > self.chunk_size:
                # split long sentence by words
                words = sentence.split()
                current = ""
                for w in words:
                    if len(current) + len(w) + 1 <= self.chunk_size:
                        current = (current + " " + w).strip()
                    else:
                        if current:
                            if len(current) > 0:
                                chunks.append(current)
                        current = w
                if current:
                    chunks.append(current)
                continue

            if len(current_chunk) + len(sentence) <= self.chunk_size:
                current_chunk += sentence + " "
            else:
                if current_chunk:
                    chunks.append(current_chunk.strip())
                current_chunk = sentence + " "

        if current_chunk:
            chunks.append(current_chunk.strip())

        return chunks

    def chunk_markdown(self, text: str) -> List[str]:
        chunks = []
        current_chunk = ""
        current_heading = ""

        lines = text.split("\n")

        for line in lines:
            heading_match = re.match(r"^(#{1,6})\s+(.+)$", line)

            if heading_match:
                if current_chunk.strip():
                    chunks.append(current_chunk.strip())

                level = len(heading_match.group(1))
                heading_text = heading_match.group(2)
                current_heading = heading_text
                current_chunk = f"{'#' * level} {heading_text}\n"
            else:
                if len(current_chunk) + len(line) <= self.chunk_size:
                    current_chunk += line + "\n"
                else:
                    if current_chunk.strip():
                        chunks.append(current_chunk.strip())
                    current_chunk = line + "\n"

        if current_chunk.strip():
            chunks.append(current_chunk.strip())

        return chunks
