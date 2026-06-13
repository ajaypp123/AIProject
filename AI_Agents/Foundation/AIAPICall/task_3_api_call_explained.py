#!/usr/bin/env python3
"""
Task 3: Making Your First API Call
Understand EVERY part of the chat completion call.
"""

import openai
import os

# Initialize client
client = openai.OpenAI(
    api_key=os.getenv("OPENAI_API_KEY"),
    base_url=os.getenv("OPENAI_API_BASE")
)

# ==========================================
# UNDERSTANDING THE API CALL STRUCTURE
# ==========================================
#
# To make an API call, you MUST provide:
# 1. model - Which AI model to use (required)
# 2. messages - Your conversation with the AI (required)
#
# The messages parameter is a list of dictionaries, each with:
# - role: Who is speaking ("user", "assistant", or "system")
# - content: What they are saying
# ==========================================

# TODO: Read each line below carefully to understand what it does
# Then uncomment ALL lines (remove the # symbols) and fill in the blanks:

response = client.chat.completions.create(
    model=os.getenv("OPENAI_API_MODEL"),  # default to a supported model
    messages=[
        {
          "role": "user",     # TODO: Use "user" - you're the user speaking
           "content": "Hello AI, please introduce yourself"   # TODO: Use "Hello AI, please introduce yourself" - your message
        }
    ]
)

# ==========================================
# REAL RESPONSE OBJECT STRUCTURE
# This is an ACTUAL response from OpenAI:
# ==========================================
"""
ChatCompletion(
    id='gen-1758773976-Ek9OxTgdgkP4Mo3ub6qf',
    choices=[
        Choice(
            finish_reason='stop',
            index=0,
            message=ChatCompletionMessage(
                content="Hello! I'm ChatGPT, an AI language model created by OpenAI. I'm here to help with a wide range of tasks such as answering questions, providing explanations, generating creative content, assisting with writing, and much more. How can I assist you today?",
                role='assistant'
            )
        )
    ],
    created=1758773976,
    model='openai/gpt-4.1-mini',
    object='chat.completion',
    usage=CompletionUsage(
        completion_tokens=55,
        prompt_tokens=13,
        total_tokens=68
    )
)
"""

# Once you uncomment and run the code above, this will execute:
try:
    if 'response' in locals() and response:
        # The AI's text is at: response.choices[0].message.content
        ai_text = response.choices[0].message.content

        print("✅ API Call Successful!")
        print(f"\n🤖 AI said: {ai_text}")
        print(f"\n📊 Total tokens used: {response.usage.total_tokens}")

        # Create marker
        os.makedirs("/tmp/markers", exist_ok=True)
        with open("/tmp/markers/task3_api_call_complete.txt", "w") as f:
            f.write("SUCCESS")
    else:
        print("❌ Complete the TODO above to make your first API call")
        print("\n📚 Required parameters:")
        print("1. model: 'openai/gpt-4.1-mini'")
        print("2. messages: [{'role': 'user', 'content': 'your message'}]")
except NameError:
    print("❌ Uncomment the code above and fill in the blanks!")
    print("\n📚 Required values:")
    print("   - model: 'openai/gpt-4.1-mini'")
    print("   - role: 'user'")
    print("   - content: 'Hello AI, please introduce yourself'")