#!/usr/bin/env python3
"""Task 4: Complete LangGraph Flow - Multi-step workflow"""

import os
import time
from typing import TypedDict
from langgraph.graph import StateGraph, END

# ╔════════════════════════════════════╗
# ║  Complete Workflow Pipeline        ║
# ╚════════════════════════════════════╝
#
#     [START]
#        │
#        ▼
#  ┌─────────────┐     State Updates:
#  │   outline   │ --> {outline: "..."}
#  │  Create an  │     Adds outline
#  │   outline   │     to state
#  └──────┬──────┘
#         │ add_edge
#         ▼
#  ┌─────────────┐
#  │    draft    │ --> {draft: "..."}
#  │  Write the  │     Adds draft
#  │   content   │     to state
#  └──────┬──────┘
#         │ add_edge
#         ▼
#  ┌─────────────┐
#  │   review    │ --> {final: "..."}
#  │  Polish &   │     Adds final
#  │  finalize   │     version
#  └──────┬──────┘
#         │
#         ▼
#      [END]
#
# IMPORTANT: Each node builds on the previous!
# - State accumulates data
# - Each node adds its contribution
# - Final state has ALL the data

print("🎯 Task 4: Complete LangGraph Flow\n")

# More complex state for our workflow
class State(TypedDict):
    topic: str
    outline: str
    draft: str
    final: str

# Node 1: Create outline
def outline_node(state: State):
    """Creates an outline for the topic"""
    print("  🔄 Creating outline...")
    time.sleep(2)  # Visualize processing time
    outline = f"Outline for '{state['topic']}':\n1. Introduction\n2. Main Points\n3. Conclusion"
    return {"outline": outline}

# Node 2: Create draft
def draft_node(state: State):
    """Creates a draft based on the outline"""
    print("  🔄 Writing draft...")
    time.sleep(2)  # Visualize processing time
    draft = f"Draft: Expanding on the outline for '{state['topic']}'..."
    return {"draft": draft}

# TODO 1: Complete the review_node function
# Hint: Create final version and return {"final": ...}
def review_node(state: State):
    """Reviews and finalizes the content"""
    print("  🔄 Reviewing and finalizing...")
    time.sleep(2)  # Visualize processing time
    final = f"Final: Reviewed and polished content about '{state['topic']}'. Ready to publish!"
    return {"___": final}  # Replace ___ with "final"

print("Building multi-step workflow:\n")

# Build the complete workflow
workflow = StateGraph(State)

# TODO 2: Add all three nodes to the graph
# Hint: Use add_node for each node
workflow.add_node("outline", outline_node)
workflow.add_node("draft", draft_node)
workflow.add_node("___", ___)  # Replace ___ with "review", review_node

# TODO 3: Connect all nodes in sequence
# Hint: outline → draft → review → END
workflow.set_entry_point("outline")
workflow.add_edge("outline", "draft")
workflow.add_edge("draft", "___")  # Replace ___ with "review"
workflow.add_edge("review", END)

# Compile and run
app = workflow.compile()
print("Graph compiled! Running workflow...\n")

# Execute the complete flow
result = app.invoke({
    "topic": "LangGraph Basics",
    "outline": "",
    "draft": "",
    "final": ""
})

print("\n" + "=" * 60)
print("WORKFLOW RESULTS:")
print(f"Topic: {result['topic']}")
print(f"Outline: {result['outline'][:50]}...")
print(f"Draft: {result['draft'][:50]}...")
print(f"Final: {result['final']}")
print("=" * 60)

print("\n💡 KEY CONCEPTS:")
print("- Multi-node workflows process data in stages")
print("- State accumulates data from each node")
print("- Each node focuses on one transformation")
print("- Edges define the execution order")

os.makedirs("/root/markers", exist_ok=True)
with open("/root/markers/task4_flow_complete.txt", "w") as f:
    f.write("TASK4_COMPLETE")

print("\n✅ Task 4 completed!")
