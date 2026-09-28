# imports
from typing_extensions import TypedDict
from langgraph.graph import StateGraph
from langgraph.checkpoint.memory import InMemorySaver
from langchain_core.messages import HumanMessage, AIMessage

class ChatState(TypedDict):
    messages: list

def chatbot_node(state: ChatState) -> ChatState:
    history = list(state.get("messages", []))
    if not history:
        # No messages yet; nothing to respond to
        return {"messages": history}

    last_message = history[-1]

    if isinstance(last_message, HumanMessage):
        user_text = last_message.content
    else:
        user_text = str(last_message)

    reply = AIMessage(
        content=f"I remember our conversation. You just said: '{user_text}'"
    )

    return {"messages": history + [reply]}

builder = StateGraph(ChatState)
builder.add_node("chatbot", chatbot_node)
builder.set_entry_point("chatbot")
builder.set_finish_point("chatbot")

checkpointer = InMemorySaver()
graph = builder.compile(checkpointer=checkpointer)

config = {"configurable": {"thread_id": "user_123"}}

result_1 = graph.invoke(
    {"messages": [HumanMessage(content="I am planning a trip to Rome next month.")]},
    config=config,
)

for msg in result_1["messages"]:
    print(type(msg).__name__, ":", msg.content)

result_2 = graph.invoke(
    {
        "messages": result_1["messages"] + [
            HumanMessage(content="Can you remind me what I told you earlier?")
        ]
    },
    config=config,
)

for msg in result_2["messages"]:
    print(type(msg).__name__, ":", msg.content)

saved_state = graph.get_state(config)

for i, msg in enumerate(saved_state["messages"], start=1):
    print(f"{i}. {type(msg).__name__}: {msg.content}")
