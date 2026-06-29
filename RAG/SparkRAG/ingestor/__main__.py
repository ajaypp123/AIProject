import sys
import argparse
import logging

from config import IngestorConfig
from ingestor import Ingestor

def main():
    parser = argparse.ArgumentParser(description="Spark RAG Document Ingestor")
    parser.add_argument("--config", type=str, help="Path to configuration file")
    parser.add_argument("--dir", type=str, help="Directory to ingest")
    parser.add_argument("--log-level", type=str, default="INFO", help="Logging level")
    args = parser.parse_args()

    logging.basicConfig(
        level=getattr(logging, args.log_level.upper()),
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    )

    config = IngestorConfig(config_file=args.config)
    if args.dir:
        config.watch_dir = args.dir

    ingestor = Ingestor(config)

    logger = logging.getLogger(__name__)
    
    if not ingestor.health_check():
        logger.error("Vector database is not healthy. Exiting.")
        sys.exit(1)

    logger.info(f"Ingesting documents from {config.watch_dir}")
    stats = ingestor.ingest_directory(config.watch_dir)
    logger.info(f"Ingestion complete: {stats}")

if __name__ == "__main__":
    main()
