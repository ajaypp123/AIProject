from typing import TypedDict

from langgraph.graph import StateGraph, START, END
from langgraph.checkpoint.memory import InMemorySaver
from langgraph.types import interrupt, Command


# ---------------------------------------------------------
# 1. Define graph state
# ---------------------------------------------------------

class AgentState(TypedDict):
    user_request: str
    proposed_action: str
    approved: bool
    result: str


# ---------------------------------------------------------
# 2. Analyze user request
# ---------------------------------------------------------

def analyze_request(state: AgentState):

    print("\n[Agent] Analyzing request...")

    request = state["user_request"]

    # In a real application this would be an LLM.
    # Here we keep it deterministic for learning HITL.

    if "restart" in request.lower():
        action = "Restart the Presto worker pod presto-worker-123"

    else:
        action = "No supported action identified"

    print(f"[Agent] Proposed action: {action}")

    return {
        "proposed_action": action
    }


# ---------------------------------------------------------
# 3. Human approval
# ---------------------------------------------------------

def human_approval(state: AgentState):

    print("\n[Agent] Waiting for human approval...")

    decision = interrupt({
        "type": "approval",
        "message": "Do you approve this operation?",
        "action": state["proposed_action"]
    })

    print(f"[Human decision] {decision}")

    return {
        "approved": decision
    }


# ---------------------------------------------------------
# 4. Decide what to do after human response
# ---------------------------------------------------------

def route_after_approval(state: AgentState):

    if state["approved"]:
        return "execute"

    return "reject"


# ---------------------------------------------------------
# 5. Execute approved operation
# ---------------------------------------------------------

def execute_action(state: AgentState):

    action = state["proposed_action"]

    print("\n[Agent] Executing action...")
    print(f"[Agent] {action}")

    # In a real application:
    #
    # subprocess.run([
    #     "kubectl",
    #     "delete",
    #     "pod",
    #     "presto-worker-123",
    #     "-n",
    #     "ezpresto"
    # ])

    result = f"Successfully executed: {action}"

    return {
        "result": result
    }


# ---------------------------------------------------------
# 6. Handle rejection
# ---------------------------------------------------------

def reject_action(state: AgentState):

    print("\n[Agent] Human rejected the operation.")

    return {
        "result": "Operation rejected by human."
    }


# ---------------------------------------------------------
# 7. Build graph
# ---------------------------------------------------------

builder = StateGraph(AgentState)

builder.add_node(
    "analyze",
    analyze_request
)

builder.add_node(
    "human_approval",
    human_approval
)

builder.add_node(
    "execute",
    execute_action
)

builder.add_node(
    "reject",
    reject_action
)


# ---------------------------------------------------------
# 8. Define graph flow
# ---------------------------------------------------------

builder.add_edge(
    START,
    "analyze"
)

builder.add_edge(
    "analyze",
    "human_approval"
)

builder.add_conditional_edges(
    "human_approval",
    route_after_approval,
    {
        "execute": "execute",
        "reject": "reject"
    }
)

builder.add_edge(
    "execute",
    END
)

builder.add_edge(
    "reject",
    END
)


# ---------------------------------------------------------
# 9. Add checkpointer
# ---------------------------------------------------------

checkpointer = InMemorySaver()

graph = builder.compile(
    checkpointer=checkpointer
)


# ---------------------------------------------------------
# 10. Create thread
# ---------------------------------------------------------

config = {
    "configurable": {
        "thread_id": "k8s-operation-001"
    }
}


# ---------------------------------------------------------
# 11. Start graph
# ---------------------------------------------------------

print("\nStarting agent...\n")

result = graph.invoke(
    {
        "user_request": "Restart the Presto worker",
        "proposed_action": "",
        "approved": False,
        "result": ""
    },
    config=config
)


print("\nGraph paused for human approval.")
print("Interrupt information:")
print(result)


# ---------------------------------------------------------
# 12. Resume after human approval
# ---------------------------------------------------------

print("\nResuming graph with human approval...\n")

result = graph.invoke(
    Command(resume=True),
    config=config
)


print("\nFinal result:")
print(result)