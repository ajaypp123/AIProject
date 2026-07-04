from sentence_transformers import SentenceTransformer
from sklearn.metrics.pairwise import cosine_similarity
import numpy as np

docs = [
    "A beginner's guide to engine troubleshooting: check fuel lines, spark, and air intake before replacing parts.",
    "Piston wear can cause loss of compression; regular maintenance and proper lubrication extend engine life.",
    "Valve timing issues often masquerade as rough idle—inspect the timing belt and camshaft alignment.",
    "Motor diagnostics for intermittent power loss: scan error codes, inspect sensors, and test the ignition coil.",
    "Understanding airflow: clogged filters, intake leaks, and MAF sensor failures reduce performance.",
    "Basic oil system checks: pressure light warnings, pump failures, and choosing the right viscosity."
]

titles = [
    "Engine Troubleshooting 101",
    "Piston Wear & Compression",
    "Valve Timing Problems",
    "Motor Diagnostics Checklist",
    "Airflow & Intake Issues",
    "Oil System Basics"
]

# Load a small, fast sentence-transformer model
model = SentenceTransformer("sentence-transformers/all-MiniLM-L6-v2")

# Embed documents and the query
doc_embeddings = model.encode(docs, normalize_embeddings=True)
query_embedding = model.encode([query_text], normalize_embeddings=True)

# Cosine similarity scores
sims = cosine_similarity(query_embedding, doc_embeddings).flatten()

# Rank by semantic similarity (descending)
sem_hits_idx = np.argsort(-sims)
sem_hits = [(titles[i], float(sims[i])) for i in sem_hits_idx]

# Show top 5 semantic hits
sem_hits[:5]