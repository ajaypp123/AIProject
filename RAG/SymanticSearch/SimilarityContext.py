from rank_bm25 import BM25Okapi
from sentence_transformers import SentenceTransformer, util
import numpy as np

docs = [
    "Enable two-factor authentication (2FA) in your account settings to add an extra security step.",
    "HbA1c measures long-term glucose; talk to your physician about tests for glycated hemoglobin.",
    "Our PTO policy covers paid time off for vacations and sick leave.",
    "How to fix engine misfires caused by bad spark plugs.",
    "Kubernetes Ingress configuration for path-based routing.",
    "Configure MFA with authenticator apps.",
    "Doctor appointment scheduling policy."
]

queries = ["How do I set up 2FA?", "What does HbA1c mean?", "sick leave policy?"]

# Prepare BM25
tokenized_corpus = [d.lower().split() for d in docs]
bm25 = BM25Okapi(tokenized_corpus)

# Choose a model (fast general-purpose)
model = SentenceTransformer('all-MiniLM-L6-v2')
doc_emb = model.encode(docs, convert_to_tensor=True, normalize_embeddings=True)

def show_results(query, k=3, alpha=0.8):
    """
    Show BM25 top-k, semantic top-k, and hybrid top-k for `query`.
    alpha: semantic weight in [0,1] for hybrid. hybrid = alpha*semantic + (1-alpha)*bm25
    """
    print(f"\nQUERY: {query}\n" + "="*60)
    # BM25 scores
    bm25_scores = np.array(bm25.get_scores(query.lower().split()))
    top_bm25 = np.argsort(-bm25_scores)[:k]
    print("BM25 top-k:")
    for i in top_bm25:
        print(f"  [{i}] {bm25_scores[i]:.3f}  {docs[i]}")

    # Semantic (SentenceTransformer) scores (cosine similarity)
    q_emb = model.encode([query], convert_to_tensor=True, normalize_embeddings=True)
    cos = util.cos_sim(q_emb, doc_emb)[0].cpu().numpy()
    top_st = np.argsort(-cos)[:k]
    print("\nSemantic (SentenceTransformer) top-k:")
    for i in top_st:
        print(f"  [{i}] {cos[i]:.3f}  {docs[i]}")

    # Normalize both scores into [0,1] to combine them
    bm25_min, bm25_max = bm25_scores.min(), bm25_scores.max()
    bm25_norm = (bm25_scores - bm25_min) / (bm25_max - bm25_min + 1e-9)

    st_min, st_max = cos.min(), cos.max()
    st_norm = (cos - st_min) / (st_max - st_min + 1e-9)

    # Hybrid: semantic-weighted combination
    hybrid = alpha * st_norm + (1.0 - alpha) * bm25_norm
    top_h = np.argsort(-hybrid)[:k]
    print(f"\nHybrid (BM25 + Semantic) top-k (alpha={alpha}):")
    for i in top_h:
        print(f"  [{i}] {hybrid[i]:.3f}  {docs[i]}")

for q in queries:
    show_results(q, k=3, alpha=0.8)
