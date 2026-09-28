#!/usr/bin/env python3
"""
Lab 06: Summarizing Chatbot
Build a chatbot that manages conversation history and summarizes when context is full.
"""

from typing import TypedDict, List, Annotated
from langgraph.graph import StateGraph, END
from langgraph.graph.message import add_messages
from langchain_openai import ChatOpenAI
from langchain_core.messages import HumanMessage, AIMessage, SystemMessage
import tiktoken

# 1. Define the Graph State Schema with Message History
class GraphState(TypedDict):
    """
    Extends basic state to include conversation history with reducer.
    Messages are automatically managed with add_messages reducer.
    """
    messages: Annotated[List, add_messages]
    summary: str
    token_count: int

# 1: Create an initial state for execution
initial_state = {
    "messages": [],  # Replace ___ with: []
    "summary": "",   # Replace ___ with: ""
    "token_count": 0 # Replace ___ with: 0
}

print("State schema defined with conversation history.")


# 2. Token Counting Function
def count_tokens(messages: List) -> int:
    """
    Counts the approximate number of tokens in messages.
    """
    encoding = tiktoken.encoding_for_model("gpt-3.5-turbo")
    total_tokens = 0

    for message in messages:
        content = message.content if hasattr(message, 'content') else str(message)
        total_tokens += len(encoding.encode(content))

    return total_tokens


# 3. Summarization Node
def summarize_node(state: GraphState) -> GraphState:
    """
    Summarizes conversation history when token count exceeds threshold.
    Keeps recent messages and replaces old ones with a summary.
    """
    print("\n--- Executing Node: SUMMARIZE ---")
    messages = state.get("messages", [])
    max_tokens = 4000  # Context window limit
    threshold = int(max_tokens * 0.8)  # Summarize at 80% capacity
    keep_recent = 5  # Keep last 5 messages

    current_tokens = count_tokens(messages)
    print(f"Current token count: {current_tokens}/{max_tokens}")

    if current_tokens < threshold:
        print("Token count within limits. No summarization needed.")
        return {"token_count": current_tokens}

    print(f"Token count ({current_tokens}) exceeds threshold ({threshold}). Summarizing...")

    # 2: Separate old and recent messages
    # Hint: Keep last 'keep_recent' messages separate from old messages
    old_messages = messages[:-keep_recent] if len(messages) > keep_recent else []  # Replace ___ with: messages[:-keep_recent] if len(messages) > keep_recent else []
    recent_messages = messages[-keep_recent:] if len(messages) > keep_recent else messages  # Replace ___ with: messages[-keep_recent:] if len(messages) > keep_recent else messages

    if not old_messages:
        return {"token_count": current_tokens}

    # Create summary of old messages
    llm = ChatOpenAI(model="gpt-5.4-nano", temperature=0)
    old_content = "\n".join([f"{msg.__class__.__name__}: {msg.content}" for msg in old_messages])

    summary_prompt = f"""Summarize the following conversation history concisely:

{old_content}

Summary:"""

    response = llm.invoke(summary_prompt)
    summary = response.content

    # 3: Create new message list with summary + recent messages
    summary_message = SystemMessage(content=f"Previous conversation summary: {summary}")
    new_messages = [summary_message] + recent_messages  # Replace ___ with: [summary_message] + recent_messages

    new_token_count = count_tokens(new_messages)
    print(f"After summarization: {new_token_count} tokens")

    return {
        "messages": new_messages,
        "summary": summary,
        "token_count": new_token_count
    }


# 4. Chat Node
def chat_node(state: GraphState) -> GraphState:
    """
    Generates a response to the user's message.
    Includes conversation history for context.
    """
    print("\n--- Executing Node: CHAT ---")
    messages = state.get("messages", [])

    llm = ChatOpenAI(model="gpt-5.4-nano", temperature=0.7)

    # 4: Add system message if not present
    if not messages or not isinstance(messages[0], SystemMessage):
        system_msg = SystemMessage(content="You are a helpful assistant. Reference previous conversation when relevant.")  # Replace ___ with: "You are a helpful assistant. Reference previous conversation when relevant."
        messages = [system_msg] + messages

    # Generate response
    response = llm.invoke(messages)

    # Add AI response to messages
    new_messages = messages + [response]

    # Update token count
    new_token_count = count_tokens(new_messages)

    print(f"Generated response. New token count: {new_token_count}")

    return {
        "messages": new_messages,
        "token_count": new_token_count
    }


# 5. Conditional Edge: Check if summarization is needed
def should_summarize(state: GraphState) -> str:
    """
    Determines if summarization is needed before chatting.
    """
    token_count = state.get("token_count", 0)
    threshold = 3200  # 80% of 4000

    # 5: Return "summarize" if over threshold, else return "chat"
    if token_count > threshold:
        return "summarize"  # Replace ___ with: "summarize"
    return "chat"  # Replace ___ with: "chat"


# 6. Build the Graph
workflow = StateGraph(GraphState)

# Add nodes
workflow.add_node("summarize", summarize_node)
workflow.add_node("chat", chat_node)

# 6: Set the entry point to 'summarize' node
workflow.set_entry_point("summarize")  # Replace ___ with: "summarize"

# Add conditional edge from summarize node
workflow.add_conditional_edges(
    "summarize",
    should_summarize,
    {
        "summarize": "summarize",  # Loop if still over threshold
        "chat": "chat"
    }
)

# Add edge from chat to END
workflow.add_edge("chat", END)

# Compile the app
app = workflow.compile()

# 7. Test the Chatbot
print("\n--- Testing Summarizing Chatbot ---")
state = {
    "messages": [HumanMessage(content="Hello! Tell me about Python.")],
    "summary": "",
    "token_count": 0
}

final_state = app.invoke(state)
print(f"\nFinal token count: {final_state['token_count']}")
print(f"Number of messages: {len(final_state['messages'])}")
if final_state.get('summary'):
    print(f"Summary created: {final_state['summary'][:100]}...")
