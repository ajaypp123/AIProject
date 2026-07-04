import ollama

# Single generation request
result = ollama.generate(
    model='gemma4:latest',
    prompt='Tell me a funny joke about Python'
)

# The client returns a dictionary. Print the textual response.
print(result['response'])
