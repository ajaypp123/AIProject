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

import pandas as pd

# Normalize/pretty print both lists with rank
kw_df = pd.DataFrame(kw_hits, columns=["Title", "TFIDF_Score"])
kw_df["KW_Rank"] = range(1, len(kw_df) + 1)

sem_df = pd.DataFrame(sem_hits, columns=["Title", "CosineSim"])
sem_df["SEM_Rank"] = range(1, len(sem_df) + 1)

# Merge on title to show both ranks together
comparison = pd.merge(sem_df, kw_df, on="Title", how="outer")

# Sort by semantic rank to highlight the "meaning" ordering
comparison_sorted = comparison.sort_values(by="SEM_Rank", na_position="last")

comparison_sorted[["Title", "SEM_Rank", "CosineSim", "KW_Rank", "TFIDF_Score"]]