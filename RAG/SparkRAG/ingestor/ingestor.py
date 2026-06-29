import logging
import json
import os
import hashlib
from pathlib import Path
from typing import Dict, List, Optional
from datetime import datetime
from uuid import uuid4

from config import IngestorConfig
from document_loader import DocumentLoader, load_documents_from_directory
from chunking import Chunker
from embeddings import get_embedding_provider
from vectordb import get_vectordb_client

logger = logging.getLogger(__name__)

class Ingestor:
    def __init__(self, config: IngestorConfig):
        self.config = config
        self.chunker = Chunker(
            chunk_size=config.chunking.chunk_size,
            chunk_overlap=config.chunking.chunk_overlap
        )
        self.embedding_provider = get_embedding_provider(
            config.embedding.provider,
            base_url=config.embedding.api_key if config.embedding.provider == "ollama" else None,
            model=config.embedding.model,
            api_key=config.embedding.api_key if config.embedding.provider == "openai" else None
        )
        self.vectordb_client = get_vectordb_client(
            config.vectordb.provider,
            url=config.vectordb.url,
            collection=config.vectordb.collection,
            api_key=config.vectordb.api_key
        )
        self.checkpoint = self._load_checkpoint()

    def _load_checkpoint(self) -> Dict:
        if os.path.exists(self.config.checkpoint_file):
            try:
                with open(self.config.checkpoint_file, 'r') as f:
                    return json.load(f)
            except Exception as e:
                logger.error(f"Failed to load checkpoint: {e}")
        return {"documents": {}}

    def _save_checkpoint(self):
        try:
            with open(self.config.checkpoint_file, 'w') as f:
                json.dump(self.checkpoint, f, indent=2)
        except Exception as e:
            logger.error(f"Failed to save checkpoint: {e}")

    def _compute_file_hash(self, file_path: str) -> str:
        try:
            with open(file_path, 'rb') as f:
                return hashlib.sha256(f.read()).hexdigest()
        except Exception as e:
            logger.error(f"Failed to compute hash for {file_path}: {e}")
            return ""

    def ingest_document(self, file_path: str, metadata: Optional[Dict] = None) -> bool:
        file_path = str(file_path)
        file_hash = self._compute_file_hash(file_path)
        
        if not file_hash:
            return False
        
        if file_path in self.checkpoint["documents"]:
            if self.checkpoint["documents"][file_path].get("hash") == file_hash:
                logger.info(f"Skipping {file_path} (already indexed)")
                return True
        
        loader = DocumentLoader()
        doc = loader.load(file_path)
        
        if not doc:
            logger.warning(f"Failed to load document: {file_path}")
            return False
        
        chunks = self.chunker.chunk_text(doc.content)
        
        chunk_data = []
        for i, chunk_content in enumerate(chunks):
            chunk_id = str(uuid4())
            embedding = self.embedding_provider.embed(chunk_content)
            
            if embedding is None:
                logger.error(f"Failed to embed chunk {i} from {file_path}")
                return False
            
            chunk_obj = {
                "id": chunk_id,
                "document_id": doc.id,
                "content": chunk_content,
                "embedding": embedding,
                "source": doc.metadata.get("source", file_path),
                "document_type": doc.metadata.get("document_type", "unknown"),
            }
            chunk_data.append(chunk_obj)
        
        if not self.vectordb_client.store_chunks(chunk_data):
            logger.error(f"Failed to store chunks for {file_path}")
            return False
        
        self.checkpoint["documents"][file_path] = {
            "hash": file_hash,
            "chunks": len(chunk_data),
            "last_ingested": datetime.now().isoformat(),
            "document_id": doc.id,
        }
        self._save_checkpoint()
        
        logger.info(f"Ingested {file_path} with {len(chunk_data)} chunks")
        return True

    def ingest_directory(self, directory: str) -> Dict:
        stats = {
            "total_files": 0,
            "successful": 0,
            "failed": 0,
            "skipped": 0,
        }
        
        docs = load_documents_from_directory(directory)
        
        for doc in docs:
            stats["total_files"] += 1
            if self.ingest_document(doc.id):
                stats["successful"] += 1
            else:
                stats["failed"] += 1
        
        return stats

    def health_check(self) -> bool:
        return self.vectordb_client.health_check()

    def delete_document(self, doc_id: str) -> bool:
        if self.vectordb_client.delete_document(doc_id):
            if doc_id in self.checkpoint["documents"]:
                del self.checkpoint["documents"][doc_id]
                self._save_checkpoint()
            return True
        return False

def main():
    import logging.config
    
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    )
    
    config = IngestorConfig()
    
    logger.info("Starting Spark RAG Ingestor")
    logger.info(f"Vector DB: {config.vectordb.provider} at {config.vectordb.url}")
    logger.info(f"Embedding: {config.embedding.provider} ({config.embedding.model})")
    
    ingestor = Ingestor(config)
    
    if not ingestor.health_check():
        logger.error("Vector database is not healthy. Exiting.")
        return
    
    logger.info(f"Ingesting documents from {config.watch_dir}")
    stats = ingestor.ingest_directory(config.watch_dir)
    
    logger.info(f"Ingestion complete: {stats}")

if __name__ == "__main__":
    main()
