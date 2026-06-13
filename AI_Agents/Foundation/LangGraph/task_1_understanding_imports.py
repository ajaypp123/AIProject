#!/usr/bin/env python3
"""Task 1: Understanding Imports - Setting up LangGraph basics"""

import os

# ╔════════════════════════════════════════╗
# ║     LangGraph Import Structure         ║
# ╠════════════════════════════════════════╣
# ║                                        ║
# ║  langgraph.graph                       ║
# ║      ├── StateGraph (Class)           ║
# ║      │    └─ Creates workflow graphs  ║
# ║      └── END (Constant)               ║
# ║           └─ Marks graph termination  ║
# ║                                        ║
# ║  typing.TypedDict                      ║
# ║      └── For defining State structure ║
# ║           └─ Type-safe state schema   ║
# ║                                        ║
# ╚════════════════════════════════════════╝

print("📚 Task 1: Understanding Imports\n")

# TODO 1: Import StateGraph and END from langgraph.graph
# Hint: from langgraph.graph import StateGraph, END
from langgraph.graph import ___, ___  # Replace with StateGraph, END

# TODO 2: Import TypedDict from typing
# Hint: from typing import TypedDict
from ___ import TypedDict  # Replace with typing

# TODO 3: Define State with messages list
# Hint: messages should be type list
class State(TypedDict):
    messages: ___  # Replace with list
    next_step: str

# Test that imports work
print("Testing imports...")
try:
    # This will only work when TODOs are completed
    test_graph = StateGraph(State)
    print("✅ StateGraph imported successfully!")
    print(f"✅ State has fields: {list(State.__annotations__.keys())}")
    print(f"✅ END constant is available: {END}")
except:
    print("❌ Complete the TODOs to make imports work!")

print("\n" + "=" * 60)
print("💡 KEY CONCEPTS:")
print("- StateGraph: Creates stateful workflow graphs")
print("- END: Marks the final node in a graph")
print("- TypedDict: Defines the structure of our state")
print("- State: Data that flows through all nodes")
print("=" * 60)

os.makedirs("/root/markers", exist_ok=True)
with open("/root/markers/task1_imports_complete.txt", "w") as f:
    f.write("TASK1_COMPLETE")

print("\n✅ Task 1 completed!")
