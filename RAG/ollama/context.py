import ollama

messages = []

while True:
    user_input = input("You: ")
    # Exit on '/exit' or an empty line
    if user_input.strip().lower() in ('/exit', ''):
        print("Exiting chat.")
        break

    # Append the user's message
    messages.append({'role': 'user', 'content': user_input})

    # Send the conversation history to the Ollama chat endpoint
    response = ollama.chat(model='gemma4:latest', messages=messages)

    # Extract assistant content and display it
    assistant_content = response['message']['content']
    print("Bot:", assistant_content)

    # Keep the assistant message in the history for context
    messages.append({'role': 'assistant', 'content': assistant_content})
