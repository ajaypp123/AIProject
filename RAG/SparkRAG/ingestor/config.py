import os
from typing import Optional
from pathlib import Path
#from dotenv import load_dotenv
import yaml

#load_dotenv()

class EmbeddingConfig:
    def __init__(self):
        self.provider = os.getenv("EMBEDDING_PROVIDER", "huggingface")
        self.model = os.getenv("EMBEDDING_MODEL", "sentence-transformers/all-MiniLM-L6-v2")
        self.api_key = os.getenv("EMBEDDING_API_KEY", "")
        self.url = os.getenv("EMBEDDING_URL", "")
        self.batch_size = int(os.getenv("EMBEDDING_BATCH_SIZE", "32"))
        # If true, fall back to deterministic embeddings when local model is unavailable
        self.allow_fallback = os.getenv("EMBEDDING_ALLOW_FALLBACK", "false").lower() == "true"

class VectorDBConfig:
    def __init__(self):
        self.provider = os.getenv("VECTORDB_PROVIDER", "chroma-local")
        self.url = os.getenv("VECTORDB_URL", "http://localhost:8000")
        self.api_key = os.getenv("VECTORDB_API_KEY", "")
        self.collection = os.getenv("VECTORDB_COLLECTION", "spark-rag")

class ChunkingConfig:
    def __init__(self):
        self.chunk_size = int(os.getenv("CHUNK_SIZE", "512"))
        self.chunk_overlap = int(os.getenv("CHUNK_OVERLAP", "50"))
        self.separator = os.getenv("CHUNK_SEPARATOR", "\n\n")
        self.strategy = os.getenv("CHUNKING_STRATEGY", "similarity")
        self.top_k = int(os.getenv("CHUNKING_TOP_K", "5"))
        self.score_threshold = float(os.getenv("CHUNKING_SCORE_THRESHOLD", "0.5"))

class IngestorConfig:
    def __init__(self, config_file: Optional[str] = None):
        self.embedding = EmbeddingConfig()
        self.vectordb = VectorDBConfig()
        self.chunking = ChunkingConfig()
        # Default watch directory: package-level data directory if present
        base = Path(__file__).resolve().parents[1]
        default_data_dir = base / 'data'
        self.watch_dir = os.getenv("INGESTOR_WATCH_DIR", str(default_data_dir) if default_data_dir.exists() else "./data")
        self.checkpoint_file = os.getenv("INGESTOR_CHECKPOINT", ".ingestor.json")
        self.skip_existing = os.getenv("SKIP_EXISTING", "true").lower() == "true"
        self.log_level = os.getenv("LOG_LEVEL", "INFO")

        # Resolve config_file: explicit -> env var -> default project development config
        config_path = config_file or os.getenv("INGESTOR_CONFIG_FILE")
        if not config_path:
            # look for config/config.development.yaml relative to project root
            base = Path(__file__).resolve().parents[1]
            candidate = base / 'config' / 'config.development.yaml'
            if candidate.exists():
                config_path = str(candidate)

        if config_path and os.path.exists(config_path):
            with open(config_path, 'r') as f:
                config_data = yaml.safe_load(f)
                if config_data:
                    # Pass config_path so relative watch.dir entries resolve relative to the config file
                    self._load_from_dict(config_data, config_path=config_path)

    def _load_from_dict(self, data: dict, config_path: Optional[str] = None):
        if 'logging' in data:
            self.log_level = data['logging'].get('level', self.log_level)
            if isinstance(self.log_level, str):
                self.log_level = self.log_level.upper()
            else:
                self.log_level = "INFO"

        if 'embedding' in data:
            self.embedding.provider = data['embedding'].get('provider', self.embedding.provider)
            self.embedding.model = data['embedding'].get('model', self.embedding.model)
            self.embedding.api_key = data['embedding'].get('api_key', self.embedding.api_key)
            self.embedding.url = data['embedding'].get('url', self.embedding.url)
            self.embedding.batch_size = data['embedding'].get('batch_size', self.embedding.batch_size)
            self.embedding.allow_fallback = data['embedding'].get('allow_fallback', self.embedding.allow_fallback)

        if 'vectordb' in data:
            self.vectordb.provider = data['vectordb'].get('provider', self.vectordb.provider)
            self.vectordb.url = data['vectordb'].get('url', self.vectordb.url)
            self.vectordb.collection = data['vectordb'].get('collection', self.vectordb.collection)
            self.vectordb.api_key = data['vectordb'].get('api_key', self.vectordb.api_key)

        if 'chunking' in data:
            self.chunking.chunk_size = data['chunking'].get('chunk_size', self.chunking.chunk_size)
            self.chunking.chunk_overlap = data['chunking'].get('chunk_overlap', self.chunking.chunk_overlap)
            self.chunking.strategy = data['chunking'].get('strategy', self.chunking.strategy)
            self.chunking.top_k = data['chunking'].get('top_k', self.chunking.top_k)
            self.chunking.score_threshold = data['chunking'].get('score_threshold', self.chunking.score_threshold)

        if 'watch' in data:
            wd = data['watch'].get('dir', self.watch_dir)
            # Resolve relative watch dir relative to the config file if provided
            if config_path and wd and not os.path.isabs(wd):
                base_dir = os.path.dirname(config_path)
                wd = os.path.normpath(os.path.join(base_dir, wd))
            self.watch_dir = wd
        if 'checkpoint' in data:
            self.checkpoint_file = data['checkpoint'].get('file', self.checkpoint_file)
        if 'skip_existing' in data:
            self.skip_existing = bool(data.get('skip_existing', self.skip_existing))
