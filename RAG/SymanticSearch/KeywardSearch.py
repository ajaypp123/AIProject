import tempfile
# Whoosh imports
from whoosh import index
from whoosh.fields import Schema, TEXT, ID
from whoosh.qparser import QueryParser

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

# Create temp index
index_dir = tempfile.mkdtemp()
schema = Schema(title=ID(stored=True), content=TEXT(stored=True))
ix = index.create_in(index_dir, schema)

# Add documents
writer = ix.writer()
for t, d in zip(titles, docs):
    writer.add_document(title=t, content=d)
writer.commit()

# Run a keyword query (lexical retrieval)
query_text = "engine troubleshooting"

with ix.searcher() as searcher:
    parser = QueryParser("content", ix.schema)
    q = parser.parse(query_text)
    results = searcher.search(q, limit=10)
    kw_hits = [(r['title'], float(r.score)) for r in results]

# (Optional) cleanup the temporary index directory when done:
# import shutil
kw_hits
