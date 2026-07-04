# ingest_and_query.py
import os
import glob
from pathlib import Path
import chromadb
from chromadb.utils import embedding_functions

# --------- (A) Choose persistence location ---------
DB_DIR = Path("chroma_db")
DB_DIR.mkdir(exist_ok=True)

# Create a persistent client (data survives restarts)
# Note: PersistentClient is available in some chromadb versions.
# If your version expects chromadb.Client(...), adapt accordingly.
client = chromadb.PersistentClient(path=str(DB_DIR))  # loads existing DB if present

# --------- (B) Choose an embedding function ---------
# Use SentenceTransformers model for embeddings
embedding_fn = embedding_functions.SentenceTransformersEmbeddingFunction(
    model_name="all-MiniLM-L6-v2"
)

# --------- (C) Get or create collection ---------
collection = client.get_or_create_collection(
    name="demo_texts",
    embedding_function=embedding_fn,
    metadata={"hnsw:space": "cosine"}  # cosine is common for semantic search
)

# --------- (D) Simple chunker ---------
def chunk_text(text: str, max_chars: int = 1500, overlap: int = 200):
    """Split long text into overlapping chunks to improve recall."""
    chunks = []
    start = 0
    n = len(text)
    while start < n:
        end = min(start + max_chars, n)
        chunk = text[start:end]
        chunks.append(chunk.strip())
        # move start forward, but keep 'overlap' characters overlapping
        start = end - overlap if (end - overlap) > start else end
    return [c for c in chunks if c]

# --------- (E) Read .txt files and prepare records ---------
DOCS_DIR = Path("data")
paths = sorted(glob.glob(str(DOCS_DIR / "*.txt")))

documents = []
metadatas = []
ids = []

for p in paths:
    file_id_base = Path(p).stem  # e.g., "beowulf"
    with open(p, "r", encoding="utf-8") as f:
        raw = f.read()

    # If text is long, chunk it; otherwise keep as a single chunk
    chunks = chunk_text(raw, max_chars=1500, overlap=200) if len(raw) > 1800 else [raw]

    for idx, ch in enumerate(chunks):
        uid = f"{file_id_base}__{idx:03d}"  # idempotent ID per file chunk
        documents.append(ch)
        metadatas.append({
            "source": os.path.basename(p),
            "chunk": idx
        })
        ids.append(uid)

if not ids:
    print("🚨 No .txt files found in ./data")
else:
    # --------- (F) Upsert into Chroma (idempotent) ---------
    # Upsert will add new records or replace existing records with same IDs.
    collection.upsert(
        ids=ids,
        documents=documents,
        metadatas=metadatas
    )
    print(f"✅ Ingested {len(ids)} records from {len(paths)} file(s).")

# --------- (G) Run a few demo queries ---------
def search(query_text: str, k: int = 4, where: dict = None):
    """
    Query the collection.
    - query_text: string to search
    - k: number of results to return
    - where: optional metadata filter dict, e.g. {"source": "beowulf.txt"}
    """
    res = collection.query(
        query_texts=[query_text],
        n_results=k,
        where=where  # optional metadata filter dict
    )

    print(f"\n🔎 Query: {query_text}")
    # robustly handle result structure
    ids_res = res.get("ids", [[]])[0]
    docs_res = res.get("documents", [[]])[0]
    metas_res = res.get("metadatas", [[]])[0]
    dists_res = res.get("distances", [[]])[0] if res.get("distances") else [None] * len(ids_res)

    for i in range(len(ids_res)):
        id_ = ids_res[i]
        meta = metas_res[i] if i < len(metas_res) else {}
        doc = docs_res[i] if i < len(docs_res) else ""
        dist = dists_res[i] if i < len(dists_res) else None
        dist_str = f"{dist:.4f}" if isinstance(dist, (int, float)) else "N/A"
        snippet = doc[:180].replace("\n", " ")
        print(f" • id={id_}  dist={dist_str}  source={meta.get('source')}\n   {snippet}...")

# Example queries
if ids:
    search("Describe how Beowulf defeats the monster's mother.", k=3)
    search("Who was Grendel, and why did he attack Heorot?", k=3)
    search("Why does Macbeth decide to kill Duncan?", k=3)