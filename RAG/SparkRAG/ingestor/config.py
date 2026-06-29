import os
from typing import Optional
from dotenv import load_dotenv
import yaml

load_dotenv()

class EmbeddingConfig:
    def __init__(self):
        self.provider = os.getenv("EMBEDDING_PROVIDER", "ollama")
        self.model = os.getenv("EMBEDDING_MODEL", "nomic-embed-text")
        self.api_key = os.getenv("EMBEDDING_API_KEY", "")
        self.batch_size = int(os.getenv("EMBEDDING_BATCH_SIZE", "32"))

class VectorDBConfig:
    def __init__(self):
        self.provider = os.getenv("VECTORDB_PROVIDER", "qdrant")
        self.url = os.getenv("VECTORDB_URL", "http://localhost:6333")
        self.api_key = os.getenv("VECTORDB_API_KEY", "")
        self.collection = os.getenv("VECTORDB_COLLECTION", "spark-rag")

class ChunkingConfig:
    def __init__(self):
        self.chunk_size = int(os.getenv("CHUNK_SIZE", "512"))
        self.chunk_overlap = int(os.getenv("CHUNK_OVERLAP", "50"))
        self.separator = os.getenv("CHUNK_SEPARATOR", "\n\n")

class IngestorConfig:
    def __init__(self, config_file: Optional[str] = None):
        self.embedding = EmbeddingConfig()
        self.vectordb = VectorDBConfig()
        self.chunking = ChunkingConfig()
        self.watch_dir = os.getenv("INGESTOR_WATCH_DIR", "./data")
        self.checkpoint_file = os.getenv("INGESTOR_CHECKPOINT", ".ingestor.json")
        self.skip_existing = os.getenv("SKIP_EXISTING", "true").lower() == "true"
        self.log_level = os.getenv("LOG_LEVEL", "INFO")

        if config_file and os.path.exists(config_file):
            with open(config_file, 'r') as f:
                config_data = yaml.safe_load(f)
                if config_data:
                    self._load_from_dict(config_data)

    def _load_from_dict(self, data: dict):
        if 'embedding' in data:
            self.embedding.provider = data['embedding'].get('provider', self.embedding.provider)
            self.embedding.model = data['embedding'].get('model', self.embedding.model)

        if 'vectordb' in data:
            self.vectordb.provider = data['vectordb'].get('provider', self.vectordb.provider)
            self.vectordb.url = data['vectordb'].get('url', self.vectordb.url)
            self.vectordb.collection = data['vectordb'].get('collection', self.vectordb.collection)

        if 'chunking' in data:
            self.chunking.chunk_size = data['chunking'].get('chunk_size', self.chunking.chunk_size)
            self.chunking.chunk_overlap = data['chunking'].get('chunk_overlap', self.chunking.chunk_overlap)
