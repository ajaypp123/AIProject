import os
from typing import Optional
from pathlib import Path
import yaml


class EmbeddingConfig:
    def __init__(self):
        self.provider = os.getenv("EMBEDDING_PROVIDER", "huggingface")
        self.model = os.getenv(
            "EMBEDDING_MODEL", "sentence-transformers/all-MiniLM-L6-v2"
        )
        self.api_key = os.getenv("EMBEDDING_API_KEY", "")
        self.url = os.getenv("EMBEDDING_URL", "")
        self.batch_size = int(os.getenv("EMBEDDING_BATCH_SIZE", "32"))
        # If true, fall back to deterministic embeddings when local model is unavailable
        self.allow_fallback = (
            os.getenv("EMBEDDING_ALLOW_FALLBACK", "false").lower() == "true"
        )


class VectorDBConfig:
    def __init__(self):
        self.provider = os.getenv("VECTORDB_PROVIDER", "chroma-local")
        self.url = os.getenv("VECTORDB_URL", "http://localhost:8000")
        self.api_key = os.getenv("VECTORDB_API_KEY", "")
        self.collection = os.getenv("VECTORDB_COLLECTION", "spark-rag")
        self.batch_size = int(os.getenv("VECTORDB_BATCH_SIZE", "32"))


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
        self.watch_dir = os.getenv("INGESTOR_WATCH_DIR", "/data")
        self.checkpoint_file = os.getenv("INGESTOR_CHECKPOINT", "/tmp/.ingestor.json")
        self.skip_existing = os.getenv("SKIP_EXISTING", "true").lower() == "true"
        self.log_level = os.getenv("LOG_LEVEL", "INFO")

        # Resolve config_file: explicit -> env var -> default project development config
        config_path = config_file or os.getenv("INGESTOR_CONFIG_FILE")

        if config_path and os.path.exists(config_path):
            with open(config_path, "r") as f:
                config_data = yaml.safe_load(f)
                if config_data:
                    # Pass config_path so relative watch.dir entries resolve relative to the config file
                    self._load_from_dict(config_data, config_path=config_path)

    def _get_watch_dir(self):
        base = Path(__file__).resolve().parents[1]
        default_data_dir = base / "data"
        return os.getenv("INGESTOR_WATCH_DIR", "/data")

    def _load_from_dict(self, data: dict, config_path: Optional[str] = None):
        if "logging" in data:
            self.log_level = data["logging"].get("level", self.log_level)
            if isinstance(self.log_level, str):
                self.log_level = self.log_level.upper()
            else:
                self.log_level = "INFO"

        if "embedding" in data:
            self.embedding.provider = data["embedding"].get(
                "provider", self.embedding.provider
            )
            self.embedding.model = data["embedding"].get("model", self.embedding.model)
            self.embedding.api_key = data["embedding"].get(
                "api_key", self.embedding.api_key
            )
            self.embedding.url = data["embedding"].get("url", self.embedding.url)
            self.embedding.batch_size = data["embedding"].get(
                "batch_size", self.embedding.batch_size
            )
            self.embedding.allow_fallback = data["embedding"].get(
                "allow_fallback", self.embedding.allow_fallback
            )

        if "vectordb" in data:
            self.vectordb.provider = data["vectordb"].get(
                "provider", self.vectordb.provider
            )
            self.vectordb.url = data["vectordb"].get("url", self.vectordb.url)
            self.vectordb.collection = data["vectordb"].get(
                "collection", self.vectordb.collection
            )
            self.vectordb.api_key = data["vectordb"].get(
                "api_key", self.vectordb.api_key
            )
            self.vectordb.batch_size = data["vectordb"].get(
                "batch_size", self.vectordb.batch_size
            )

        if "chunking" in data:
            self.chunking.chunk_size = data["chunking"].get(
                "chunk_size", self.chunking.chunk_size
            )
            self.chunking.chunk_overlap = data["chunking"].get(
                "chunk_overlap", self.chunking.chunk_overlap
            )
            self.chunking.strategy = data["chunking"].get(
                "strategy", self.chunking.strategy
            )
            self.chunking.top_k = data["chunking"].get("top_k", self.chunking.top_k)
            self.chunking.score_threshold = data["chunking"].get(
                "score_threshold", self.chunking.score_threshold
            )

        if "ingestion" in data:
            self.watch_dir = data["ingestion"].get("watch_dir", self.watch_dir)
            self.checkpoint_file = data["ingestion"].get(
                "checkpoint", self.checkpoint_file
            )
            self.skip_existing = bool(
                data["ingestion"].get("skip_existing", self.skip_existing)
            )
