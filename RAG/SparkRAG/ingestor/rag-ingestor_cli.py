#!/usr/bin/env python

"""
version:
rag-ingestor_cli --version

store:
rag-ingestor_cli store --conf SparkRAG/config/config.development.yaml

delete:
rag-ingestor_cli delete SparkRAG/data/Spark.pdf --conf SparkRAG/config/config.development.yaml
"""

import argparse
import json
import logging
import sys

from config import IngestorConfig
from ingestor import Ingestor

logger = logging.getLogger(__name__)


class IngestorAppCli:
    def build_parser(self) -> argparse.ArgumentParser:
        common_parser = argparse.ArgumentParser(add_help=False)
        common_parser.add_argument(
            "--conf",
            dest="config_file",
            default=None,
            help="Optional path to a YAML config file",
        )

        parser = argparse.ArgumentParser(
            prog=sys.argv[0],
            description="Spark RAG ingestor CLI",
            formatter_class=argparse.ArgumentDefaultsHelpFormatter,
            parents=[common_parser],
        )
        parser.add_argument(
            "--version",
            action="version",
            version="%(prog)s 1.0.0",
        )

        subparsers = parser.add_subparsers(dest="command", help="Available commands")
        subparsers.required = False

        health_parser = subparsers.add_parser(
            "health",
            help="Check whether the configured vector database is healthy",
            parents=[common_parser],
        )
        health_parser.set_defaults(handler=self._handle_health)

        store_parser = subparsers.add_parser(
            "store",
            help="Chunk, embed, and store documents into the vector database",
            parents=[common_parser],
        )
        store_parser.add_argument(
            "path",
            nargs="?",
            default=None,
            help="Directory or file to store. Defaults to the configured watch_dir.",
        )
        store_parser.set_defaults(handler=self._handle_store)

        delete_parser = subparsers.add_parser(
            "delete",
            help="Delete stored documents from the vector database",
            parents=[common_parser],
        )
        delete_parser.add_argument(
            "target",
            nargs="?",
            default=None,
            help="File path or document id to delete. ('all' to delete all ids)",
        )
        delete_parser.set_defaults(handler=self._handle_delete)

        return parser

    def run(self, argv=None) -> int:
        parser = self.build_parser()
        args = parser.parse_args(argv)

        if not getattr(args, "command", None):
            parser.print_help()
            return 0

        config = IngestorConfig(args.config_file)

        logging.basicConfig(
            level=config.log_level,
            format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
        )

        logger.info(f"Using config file: {args.config_file}")
        logger.info("Starting Spark RAG Ingestor")
        logger.info(f"Vector DB: {config.vectordb.provider} at {config.vectordb.url}")
        logger.info(
            f"Embedding: {config.embedding.provider} ({config.embedding.model})"
        )

        ingestor = Ingestor(config)
        return args.handler(ingestor, config, args)

    def _handle_store(self, ingestor: Ingestor, config: IngestorConfig, args) -> int:
        target = args.path or config.watch_dir
        logger.info(f"Storing documents from {target}")
        stats = ingestor.ingest_directory(target)
        print(json.dumps(stats, indent=2))
        logger.info(f"Store complete: {stats}")
        return 0 if stats["failed"] == 0 else 1

    def _handle_delete(self, ingestor: Ingestor, config: IngestorConfig, args) -> int:
        if not args.target:
            print(
                json.dumps(
                    {
                        "deleted": False,
                        "error": "A target file path or document id is required",
                    },
                    indent=2,
                )
            )
            return 2

        success = ingestor.delete_document(args.target)
        print(json.dumps({"deleted": success, "target": args.target}, indent=2))
        return 0 if success else 1

    def _handle_health(self, ingestor: Ingestor, config: IngestorConfig, args) -> int:
        is_healthy = ingestor.health_check()
        print(json.dumps({"health_check": is_healthy}, indent=2))
        return 0 if is_healthy else 1

    def _handle_version(self, args) -> int:
        print(json.dumps({"version": Ingestor.version()}))
        return 0


def main(argv=None) -> int:
    return IngestorAppCli().run(argv)


if __name__ == "__main__":
    sys.exit(main())
