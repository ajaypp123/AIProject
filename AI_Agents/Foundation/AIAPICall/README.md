# Lab 1: AI Fundamentals - Your First AI Journey

## 🎯 What You'll Learn
Making your first AI API call through 5 micro-tasks - each building on the previous one, like climbing stairs one step at a time.

### Your 5-Step Journey:
1. **Task 1**: Import the OpenAI library (2 lines)
2. **Task 2**: Initialize the client with credentials (2 lines)
3. **Task 3**: Make your first API call (3 lines)
4. **Task 4**: Extract the AI's response (1 line)
5. **Task 5**: Understand token costs (2 lines)

## 🚀 Before You Start
- You've never coded with AI before? **Perfect!** This lab is for absolute beginners.
- Each task is just 1-3 lines to fill in
- Complete them in order (1→2→3→4→5)
- Total time: ~15 minutes

## 📚 Key Concepts Explained Simply

### What is an API Call?
Think of it like ordering food at a restaurant:
1. You (your code) place an order (send a message)
2. The kitchen (OpenAI's servers) prepares your order
3. The waiter (API) brings you the food (AI's response)
4. You enjoy your meal (use the response in your app)

### What are Tokens?
- Tokens are like "word pieces" the AI uses to understand and create text
- Both your question and AI's answer use tokens
- More tokens = higher cost (like phone minutes)
- **Rule of thumb**: 1 token ≈ 4 characters or 0.75 words

### The Response Object
When you call the API, you get back a "package" containing:
- `choices[0].message.content` → The actual text response
- `usage.prompt_tokens` → Tokens in your question
- `usage.completion_tokens` → Tokens in AI's answer
- `usage.total_tokens` → Total tokens (what you pay for)

# **📁 Lab Files - 5 Progressive Tasks**

# Task 1

## What is OpenAI?
OpenAI is the company that created ChatGPT. They build the most advanced AI models in the world.

### 🤖 Their AI Models
- GPT-4: The smartest (but expensive)
- GPT-4.1-mini: Fast and cheap (we'll use this)
- GPT-3.5: The older, simpler version

### 🔑 The OpenAI Library
The openai Python library is your gateway to these AI models. It handles:
- Authentication with your API key
- Sending your questions to AI servers
- Getting responses back

💡 Think of it like this: OpenAI runs the AI brain in the cloud. Your Python code talks to it using their library.

## 1. **task_1_import_explained.py** (Start Here! 🚀)
Import the OpenAI library - your gateway to AI
- **TODO**: Fill in 2 import statements
- **Lines to complete**: 13, 14
- **Learn**: What libraries you need for AI
- **Time**: 1 minute

```sh
python3 /root/code/task_1_import_setup.py
```

# Task 2

## Understanding API Authentication & Client Setup

### 🤝 What is an API Client?
Think of the client as your connection manager to OpenAI. It handles all the complex networking so you can just focus on asking questions!

### 🔑 What's an API Key?
Your API key is like a password that tells OpenAI:

- Who you are (authentication)
- That you're allowed to use their AI
- Which account to bill for usage
OPENAI_API_KEY = Your secret access code

### 🌐 What's the Base URL?
The base URL tells your client where to send requests:

- It's the address of OpenAI's servers
- Like a web address, but for API calls
- Different URLs can point to different AI services
OPENAI_API_BASE = Where the AI servers live

### 🚀 What's Next?
In Task 2, you'll create this client using your API key and base URL - connecting your code to OpenAI's powerful AI!


## 2. **task_2_client_initialization.py**
Connect to OpenAI servers with your credentials
- **TODO**: Fill in 2 environment variable names
- **Lines to complete**: 16, 17
- **Learn**: How authentication works
- **Time**: 2 minutes

# Task 3

## 💬 What are Chat Completions?

### 🗣️ It's a Conversation with AI
Chat Completions is OpenAI's conversational API. You send messages, AI responds - just like texting! The magic happens with client.chat.completions.create()

### 🧠 Choose Your AI Model
Different models = different AI brains. You MUST specify which one:

- gpt-4.1-mini: Fast, cheap, smart enough for most tasks (we'll use this!)
- gpt-4: Genius level, but slow and expensive
- gpt-3.5: Older, less capable

### 🎭 The Three Roles in Conversations
- "system": Instructions for how AI should behave (optional)
- "user": That's you asking questions
- "assistant": The AI's responses

### 📝 The Messages Format
Every API call needs these 2 parameters:
```py
client.chat.completions.create(
    model="openai/gpt-4.1-mini",     # Which AI brain
    messages=[                        # The conversation
        {"role": "user", "content": "Your question here"}
    ]
)
```
Messages is an array because you can send conversation history!

### 🚀 What's Next?
In Task 3, you'll use this exact format to make your first real API call and have your first conversation with AI!

## 3. **task_3_api_call_explained.py**
Make your first real API call!
- **TODO**: Uncomment and fill in 3 values
- **Lines to complete**: 33, 36, 37
- **Learn**: How to talk to AI
- **Time**: 3 minutes

```sh
python3 /root/code/task_3_api_call_explained.py
```

# Task 4

## 📦 Understanding the Response Object

### 🔍 Why This Weird Path?
The response path response.choices[0].message.content seems complex, but each part has a reason:

### 📊 Breaking It Down
- response: The entire response object from OpenAI
- .choices: Array of possible responses (AI can generate multiple)
- [0]: Get the first (and usually only) choice
- .message: The message object containing role and content
- .content: The actual text of the AI's response!

### 🎯 Other Useful Fields
- response.usage: Token counts (prompt, completion, total)
- response.model: Which model actually responded
- response.created: Timestamp of the response
- response.id: Unique identifier for this response

### 💡 Pro Tip:
99% of the time, you just need response.choices[0].message.content for the text!


## 4. **task_4_extract_response.py**
Get the AI's actual text response
- **TODO**: Fill in the magic path to the content
- **Line to complete**: 56
- **Learn**: Navigate the response object
- **Time**: 2 minutes

```sh
python3 /root/code/task_4_extract_response.py
```

# Task 5

## 💰 Understanding Tokens & AI Economics

### 🧩 What Are Tokens?
Tokens are pieces of words that AI uses to process text:

- Simple words = 1 token ("cat", "run")
- Complex words = multiple tokens ("unbelievable" = ~3 tokens)
- Rough estimate: 1 token ≈ 4 characters
- Average: 1 token ≈ 0.75 words

### 📊 The Three Token Types
- prompt_tokens: Your question (what you send)
- completion_tokens: AI's answer (what you get back)
- total_tokens: Sum of both (what you pay for)
Find them in: `response.usage`

### 💸 Why Tokens = Money
AI companies charge by tokens consumed:

- Input tokens: $0.80 per million ($0.0008/1K)
- Output tokens: $3.20 per million ($0.0032/1K)
Notice: Output costs 4x more than input!

💡 This is why keeping AI responses concise saves money!

### 🏢 Real Business Impact
Example: Customer Support Chatbot

1,000 queries/day × $0.001 per query = $1/day
→ Monthly: $30
→ Yearly: $365

vs. Human agent: $25/hour × 8 hours = $200/day!

### 🚀 What's Next?
In Task 5, you'll calculate real costs for API calls and see how to optimize spending for your business use case!

### 5. **task_5_token_costs.py**
Understand the economics - tokens = money
- **TODO**: Fill in 2 token extraction paths
- **Lines to complete**: 50, 51
- **Learn**: How AI usage translates to costs
- **Time**: 3 minutes

---

## 🎮 How to Complete This Lab

### Step 1: Verify Your Environment
```bash
source /root/venv/bin/activate
python /root/code/verify_environment.py
```

### Step 2: Complete Each Task
Look for the `TODO` markers and fill in the `___` blanks:

```python
# TODO: Import OpenAI (Line 13)
import ___  # ← Replace ___ with: openai
```

Each TODO tells you EXACTLY what to type!

### Step 3: Run Your Code
Start with Task 1:
```bash
python /root/code/task_1_import_explained.py
```

Then continue with Tasks 2, 3, 4, and 5 in order.

### Step 4: Check Your Progress
Each successful task creates a marker file. You'll see:
```
✅ Congratulations! You made your first AI API call!
```

## 💡 Tips for Success

### For Task 1 (Imports):
- Line 13: `import openai`
- Line 14: `from openai import OpenAI`

### For Task 2 (Client):
- Line 16: `"OPENAI_API_KEY"`
- Line 17: `"OPENAI_API_BASE"`

### For Task 3 (API Call):
- Line 33: model = `"openai/gpt-4.1-mini"`
- Line 36: role = `"user"`
- Line 37: content = `"Hello AI, please introduce yourself"`

### For Task 4 (Response):
- Line 56: `choices[0].message.content`
- This is THE path you'll always use!

### For Task 5 (Tokens):
- Line 50: `response.usage.prompt_tokens`
- Line 51: `response.usage.completion_tokens`

## ❓ Common Questions

**Q: What if I get an error?**
A: Each TODO shows EXACTLY what to type. Copy it character-for-character!

**Q: Why 5 separate tasks instead of one big file?**
A: Breaking it down helps you understand each piece. Once you get it, you'll combine them all!

**Q: What's with all the underscores (___)?**
A: They mark exactly where you need to fill in code. Replace them with the suggested values.

**Q: Can I skip ahead?**
A: Please don't! Each task builds on the previous one. Task 3 needs Task 2's client, etc.

## 🎉 After Completing This Lab

You'll understand:
- ✅ How to import AI libraries
- ✅ How to authenticate with AI services
- ✅ How to make API calls
- ✅ How to extract AI responses
- ✅ How tokens and costs work

You'll have written your first 10 lines of AI code!

Ready? Let's start with Task 1 - just 2 lines to import! 🚀