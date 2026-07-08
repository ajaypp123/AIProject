
import logging
import json
import os
import logging
import hashlib
from pathlib import Path
from typing import Dict
from datetime import datetime
from uuid import uuid4

from config import IngestorConfig
from chunking import Chunker
from embeddings import get_embedding_provider
from vectordb import get_vectordb_client

from document_loader import DocumentLoader

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
            base_url=getattr(config.embedding, 'url', None),
            model=getattr(config.embedding, 'model', None),
            api_key=getattr(config.embedding, 'api_key', None),
            batch_size=getattr(config.embedding, 'batch_size', None)
        )
        self.vectordb_client = get_vectordb_client(
            config.vectordb.provider,
            url=config.vectordb.url,
            collection=config.vectordb.collection,
            api_key=config.vectordb.api_key,
            batch_size=getattr(config.vectordb, 'batch_size', None)
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
            raise Exception(f"Failed to compute hash for {file_path}: {e}")

    def ingest_document(self, file_path: str) -> bool:
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
        logging.info(f"Checking all document in provided path {directory}.")
        documents = []
        directory_path = Path(directory)

        if not directory_path.is_dir():
            logger.warning(f"Directory not found: {directory}")
            return stats

        for file_path in directory_path.rglob('*'):
            stats["total_files"] += 1
            if not file_path.is_file(): continue

            sufix = file_path.suffix.lower()
            if sufix not in DocumentLoader.SUPPORTED_TYPES:
                logging.warning(f"Skipping, {file_path}")
                stats["skipped"] += 1
                continue

            logging.info(f"Please wait, ingesting document {file_path}")
            if file_path and self.ingest_document(file_path):
                logging.debug(f"Successfully ingested document {file_path}")
                stats["successful"] += 1
            else:
                logging.warning(f"Failed ingesting document {file_path}")
                stats["failed"] += 1

        return stats

    def health_check(self) -> bool:
        return self.vectordb_client.health_check()

    def delete_document(self, doc_id: str) -> bool:
        # Accept either a file path (checkpoint key) or a document id stored in checkpoint
        # If a file path is provided, resolve the document_id from the checkpoint
        document_id = doc_id
        file_key = None
        if doc_id in self.checkpoint.get("documents", {}):
            # doc_id is a file path key
            file_key = doc_id
            document_id = self.checkpoint["documents"][doc_id].get("document_id")

        if not document_id:
            logger.warning(f"Document id not found for {doc_id}")
            return False

        if self.vectordb_client.delete_document(document_id):
            if file_key:
                del self.checkpoint["documents"][file_key]
                self._save_checkpoint()
            else:
                # remove any checkpoint entries matching this document_id
                to_delete = [k for k, v in self.checkpoint.get("documents", {}).items() if v.get("document_id") == document_id]
                for k in to_delete:
                    del self.checkpoint["documents"][k]
                if to_delete:
                    self._save_checkpoint()
            return True
        return False

    def version():
        return "1.0.0"
