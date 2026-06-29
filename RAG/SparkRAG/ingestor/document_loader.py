import logging
import os
from pathlib import Path
from typing import List, Dict, Optional
import json
import csv
from datetime import datetime

logger = logging.getLogger(__name__)

class Document:
    def __init__(self, id: str, content: str, metadata: Dict):
        self.id = id
        self.content = content
        self.metadata = metadata

class DocumentLoader:
    SUPPORTED_TYPES = {'.pdf', '.txt', '.md', '.csv', '.json', '.html'}

    def load(self, file_path: str) -> Optional[Document]:
        file_path = Path(file_path)
        
        if not file_path.exists():
            logger.warning(f"File not found: {file_path}")
            return None
        
        suffix = file_path.suffix.lower()
        
        try:
            if suffix == '.pdf':
                return self._load_pdf(file_path)
            elif suffix == '.txt':
                return self._load_text(file_path)
            elif suffix == '.md':
                return self._load_markdown(file_path)
            elif suffix == '.csv':
                return self._load_csv(file_path)
            elif suffix == '.json':
                return self._load_json(file_path)
            elif suffix == '.html':
                return self._load_html(file_path)
            else:
                logger.warning(f"Unsupported file type: {suffix}")
                return None
        except Exception as e:
            logger.error(f"Error loading {file_path}: {e}")
            return None

    def _load_pdf(self, file_path: Path) -> Optional[Document]:
        try:
            from pypdf import PdfReader
            reader = PdfReader(str(file_path))
            content = ""
            for page_num, page in enumerate(reader.pages):
                content += f"\n--- Page {page_num + 1} ---\n"
                content += page.extract_text()
            
            metadata = {
                "source": str(file_path),
                "document_type": "pdf",
                "last_modified": datetime.fromtimestamp(file_path.stat().st_mtime).isoformat(),
                "filename": file_path.name,
            }
            
            return Document(
                id=file_path.stem,
                content=content,
                metadata=metadata
            )
        except Exception as e:
            logger.error(f"PDF loading failed: {e}")
            return None

    def _load_text(self, file_path: Path) -> Optional[Document]:
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()
            
            metadata = {
                "source": str(file_path),
                "document_type": "text",
                "last_modified": datetime.fromtimestamp(file_path.stat().st_mtime).isoformat(),
                "filename": file_path.name,
            }
            
            return Document(
                id=file_path.stem,
                content=content,
                metadata=metadata
            )
        except Exception as e:
            logger.error(f"Text loading failed: {e}")
            return None

    def _load_markdown(self, file_path: Path) -> Optional[Document]:
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()
            
            metadata = {
                "source": str(file_path),
                "document_type": "markdown",
                "last_modified": datetime.fromtimestamp(file_path.stat().st_mtime).isoformat(),
                "filename": file_path.name,
            }
            
            return Document(
                id=file_path.stem,
                content=content,
                metadata=metadata
            )
        except Exception as e:
            logger.error(f"Markdown loading failed: {e}")
            return None

    def _load_csv(self, file_path: Path) -> Optional[Document]:
        try:
            rows = []
            with open(file_path, 'r', encoding='utf-8') as f:
                reader = csv.DictReader(f)
                for row in reader:
                    rows.append(row)
            
            content = "CSV Data:\n"
            if rows:
                content += "\nColumns: " + ", ".join(rows[0].keys()) + "\n"
                for i, row in enumerate(rows[:100]):
                    content += f"\nRow {i+1}:\n"
                    for key, value in row.items():
                        content += f"  {key}: {value}\n"
            
            metadata = {
                "source": str(file_path),
                "document_type": "csv",
                "last_modified": datetime.fromtimestamp(file_path.stat().st_mtime).isoformat(),
                "filename": file_path.name,
                "row_count": len(rows),
            }
            
            return Document(
                id=file_path.stem,
                content=content,
                metadata=metadata
            )
        except Exception as e:
            logger.error(f"CSV loading failed: {e}")
            return None

    def _load_json(self, file_path: Path) -> Optional[Document]:
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                data = json.load(f)
            
            content = json.dumps(data, indent=2)
            
            metadata = {
                "source": str(file_path),
                "document_type": "json",
                "last_modified": datetime.fromtimestamp(file_path.stat().st_mtime).isoformat(),
                "filename": file_path.name,
            }
            
            return Document(
                id=file_path.stem,
                content=content,
                metadata=metadata
            )
        except Exception as e:
            logger.error(f"JSON loading failed: {e}")
            return None

    def _load_html(self, file_path: Path) -> Optional[Document]:
        try:
            from html.parser import HTMLParser
            
            class TextExtractor(HTMLParser):
                def __init__(self):
                    super().__init__()
                    self.text = []
                
                def handle_data(self, data):
                    text = data.strip()
                    if text:
                        self.text.append(text)
            
            with open(file_path, 'r', encoding='utf-8') as f:
                html = f.read()
            
            extractor = TextExtractor()
            extractor.feed(html)
            content = "\n".join(extractor.text)
            
            metadata = {
                "source": str(file_path),
                "document_type": "html",
                "last_modified": datetime.fromtimestamp(file_path.stat().st_mtime).isoformat(),
                "filename": file_path.name,
            }
            
            return Document(
                id=file_path.stem,
                content=content,
                metadata=metadata
            )
        except Exception as e:
            logger.error(f"HTML loading failed: {e}")
            return None

def load_documents_from_directory(directory: str) -> List[Document]:
    documents = []
    directory_path = Path(directory)
    
    if not directory_path.is_dir():
        logger.warning(f"Directory not found: {directory}")
        return documents
    
    loader = DocumentLoader()
    
    for file_path in directory_path.rglob('*'):
        if file_path.is_file() and file_path.suffix.lower() in DocumentLoader.SUPPORTED_TYPES:
            doc = loader.load(str(file_path))
            if doc:
                documents.append(doc)
                logger.info(f"Loaded: {file_path.name}")
    
    return documents
