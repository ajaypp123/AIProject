# main.py
import csv
import json
from pathlib import Path
from typing import List, Dict, Optional

def parse_csv_for_rag(csv_file_path: str, output_file_path: Optional[str] = None) -> List[Dict]:
    """
    Parse a CSV file into a list of documents suitable for RAG ingestion.
    Each document contains:
      - id: a generated id (doc_{row_index})
      - text: a concatenated text representation of non-empty fields
      - metadata: includes source, row_number, and the original CSV fields
    """
    documents: List[Dict] = []

    try:
        with open(csv_file_path, 'r', encoding='utf-8') as csvfile:
            reader = csv.DictReader(csvfile)
            fieldnames = reader.fieldnames or []
            print(f"Found columns: {fieldnames}")

            for idx, row in enumerate(reader):
                # Build a readable text representation from non-empty fields
                text_parts = []
                for key, value in row.items():
                    if value is not None:
                        value_str = value.strip()
                        if value_str:
                            text_parts.append(f"{key}: {value_str}")

                text = " | ".join(text_parts)

                # Copy row fields into metadata (preserve as strings)
                row_metadata = {k: (v if v is not None else "") for k, v in row.items()}

                document = {
                    "id": f"doc_{idx}",
                    "text": text,
                    "metadata": {
                        "source": str(csv_file_path),
                        "row_number": idx,
                        **row_metadata
                    }
                }

                documents.append(document)

    except FileNotFoundError:
        print(f"Error: {csv_file_path} not found!")
        return []
    except Exception as e:
        print(f"Error reading {csv_file_path}: {e}")
        return []

    print(f"Processed {len(documents)} documents")

    # Optionally save to JSON file
    if output_file_path:
        try:
            with open(output_file_path, 'w', encoding='utf-8') as outfile:
                json.dump(documents, outfile, indent=2, ensure_ascii=False)
            print(f"Saved documents to {output_file_path}")
        except Exception as e:
            print(f"Error writing {output_file_path}: {e}")

    return documents


def main():
    # Input and output file paths (adjust as needed)
    csv_file = "sample_data.csv"
    output_file = "rag_documents.json"

    if not Path(csv_file).exists():
        print(f"Error: {csv_file} not found!")
        return

    documents = parse_csv_for_rag(csv_file, output_file)

    if documents:
        # Print a sample document and the total number of chunks
        print("\nSample document:")
        print(json.dumps(documents[0], indent=2, ensure_ascii=False))
        total = len(documents)
        print(f"\nTotal Chunks: {total}")


if __name__ == "__main__":
    main()