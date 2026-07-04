# pdf_parser.py
from typing import List, Dict, Any
from pathlib import Path
import json
import PyPDF2
from langchain_text_splitters import RecursiveCharacterTextSplitter


class PDFParser:
    """Parse PDF files and prepare them for RAG ingestion."""

    def __init__(self, chunk_size: int = 1000, chunk_overlap: int = 200):
        """
        Initialize the PDF parser.

        Args:
            chunk_size: Maximum size of each text chunk in characters.
            chunk_overlap: Number of characters to overlap between chunks.
        """
        self.chunk_size = chunk_size
        self.chunk_overlap = chunk_overlap
        # separators chosen to prefer paragraph/sentence boundaries, then fallback to spaces
        self.text_splitter = RecursiveCharacterTextSplitter(
            chunk_size=chunk_size,
            chunk_overlap=chunk_overlap,
            length_function=len,
            separators=["\n\n", "\n", ". ", " ", ""],
        )

    def extract_text_from_pdf(self, pdf_path: str) -> str:
        """
        Extract all selectable text from a PDF file.

        Args:
            pdf_path: Path to the PDF file.

        Returns:
            Extracted text as a single string.

        Raises:
            FileNotFoundError: If the PDF file doesn't exist.
            Exception: If there's an error reading the PDF.
        """
        path = Path(pdf_path)
        if not path.exists():
            raise FileNotFoundError(f"PDF file not found: {pdf_path}")

        text_parts: List[str] = []
        try:
            with open(pdf_path, "rb") as f:
                reader = PyPDF2.PdfReader(f)
                num_pages = len(reader.pages)
                for i in range(num_pages):
                    page = reader.pages[i]
                    # extract_text() may return None for some pages; guard it
                    page_text = page.extract_text()
                    if page_text:
                        text_parts.append(page_text)
        except Exception as e:
            raise Exception(f"Error reading PDF: {e}")

        # Join pages with double newline to preserve some structure
        return "\n\n".join(text_parts).strip()

    def chunk_text(self, text: str) -> List[str]:
        """
        Split text into chunks suitable for RAG ingestion.

        Args:
            text: The text to chunk.

        Returns:
            List of text chunks.
        """
        if not text:
            return []
        return self.text_splitter.split_text(text)

    def parse_pdf(self, pdf_path: str) -> List[Dict[str, Any]]:
        """
        Parse a PDF and return chunked text with metadata.

        Args:
            pdf_path: Path to the PDF file.

        Returns:
            List of dictionaries containing chunk text and metadata.
        """
        text = self.extract_text_from_pdf(pdf_path)
        chunks = self.chunk_text(text)

        results: List[Dict[str, Any]] = []
        total_chunks = len(chunks)
        for idx, chunk in enumerate(chunks):
            results.append({
                "chunk_id": idx,
                "text": chunk,
                "source": str(Path(pdf_path).resolve()),
                "chunk_size": len(chunk),
                "total_chunks": total_chunks,
            })
        return results

    def parse_pdf_to_file(self, pdf_path: str, output_path: str) -> None:
        """
        Parse a PDF and save the chunked results to a text file for inspection.

        Args:
            pdf_path: Path to the PDF file.
            output_path: Path to write the output text file.
        """
        results = self.parse_pdf(pdf_path)

        with open(output_path, "w", encoding="utf-8") as f:
            f.write("PDF Parser Results\n")
            f.write(f"Source: {pdf_path}\n")
            f.write(f"Total Chunks: {len(results)}\n")
            f.write("-" * 80 + "\n\n")

            for result in results:
                f.write(f"Chunk {result['chunk_id'] + 1}/{result['total_chunks']}\n")
                f.write(f"Size: {result['chunk_size']} characters\n")
                f.write("-" * 80 + "\n")
                f.write(result["text"].strip() + "\n\n")
                f.write("=" * 80 + "\n\n")


def main():
    # Path to your sample PDF
    pdf_path = "simple.pdf"  # <-- replace with your PDF path

    parser = PDFParser(chunk_size=1000, chunk_overlap=200)

    print("Example 1: Parsing PDF to structured data")
    print("-" * 80)

    try:
        results = parser.parse_pdf(pdf_path)
        print(f"Successfully parsed: {pdf_path}")
        print(f"Total chunks created: {len(results)}")
        if results:
            print("\nFirst chunk preview:")
            print(results[0]["text"][:200] + "...")
        # Save structured chunks to JSON for RAG ingestion
        with open("chunks.json", "w", encoding="utf-8") as f:
            json.dump(results, f, indent=2, ensure_ascii=False)
        print("\nChunks saved to: chunks.json")
    except FileNotFoundError:
        print(f"Error: PDF file '{pdf_path}' not found. Please update the path variable.")
    except Exception as e:
        print(f"Error processing PDF: {e}")

    print("\n" + "=" * 80 + "\n")

    print("Example 2: Parsing PDF to text file")
    print("-" * 80)
    try:
        parser.parse_pdf_to_file(pdf_path, "output.txt")
        print("Chunks saved to: output.txt")
    except FileNotFoundError:
        print(f"Error: PDF file '{pdf_path}' not found.")
    except Exception as e:
        print(f"Error: {e}")

    print("\n" + "=" * 80 + "\n")

    print("Example 3: Custom chunking parameters")
    print("-" * 80)
    # Smaller chunks (more pieces)
    small_parser = PDFParser(chunk_size=500, chunk_overlap=100)
    try:
        small_results = small_parser.parse_pdf(pdf_path)
        print(f"Chunks with size=500: {len(small_results)}")
    except Exception:
        print("Skipping small chunk example due to missing or unreadable PDF file")

    # Larger chunks (more context per chunk)
    large_parser = PDFParser(chunk_size=2000, chunk_overlap=400)
    try:
        large_results = large_parser.parse_pdf(pdf_path)
        print(f"Chunks with size=2000: {len(large_results)}")
    except Exception:
        print("Skipping large chunk example due to missing or unreadable PDF file")


if __name__ == "__main__":
    main()
