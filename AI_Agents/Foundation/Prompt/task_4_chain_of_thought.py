#!/usr/bin/env python3
"""
Task 4: Chain-of-Thought Prompting - Step-by-Step Reasoning
Guide the AI through a logical thinking process to solve complex problems.

Learning Goal: Master chain-of-thought prompting for complex reasoning tasks.
"""

import os
from langchain_openai import ChatOpenAI
from langchain_core.prompts import PromptTemplate

def main():
    print("🎯 Task 4: Chain-of-Thought Prompting")
    print("=" * 50)

    # Initialize LLM
    llm = ChatOpenAI(
        model="openai/gpt-4.1-mini",
        api_key=os.getenv("OPENAI_API_KEY"),
        base_url=os.getenv("OPENAI_API_BASE"),
        temperature=0.7
    )

    print("\n📝 Part 1: Direct Approach (Without Chain-of-Thought)")
    print("-" * 40)

    # Direct prompt without reasoning steps
    direct_prompt = """Our current data retention policy states:\n- Customer personal data is stored indefinitely\n- There is no documented data deletion or review process\n- System backups are retained forever and include personal data\- No distinction is made between active and inactive users\nFix this policy to comply with GDPR."""

    print(f"❌ Direct prompt: {direct_prompt}")
    direct_response = llm.invoke(direct_prompt)
    print(f"\nDirect response preview: {direct_response.content[:200]}...")
    print("Problem: Lacks structure and may miss important details!")

    print("\n📝 Part 2: Chain-of-Thought Approach")
    print("-" * 40)

    # TODO 1: Create reasoning steps for the AI to follow
    reasoning_steps = """___"""  # Replace ___ with: "Step 1: Review GDPR requirements related to data retention and storage limitation\nStep 2: Identify compliance gaps in the current data retention policy\nStep 3: Reference industry best practices for data retention and deletion\nStep 4: Draft specific, GDPR-compliant policy changes\nStep 5: Propose an implementation and enforcement timeline"

    print("🧠 Reasoning steps defined:")
    print(reasoning_steps)

    # TODO 2: Build chain-of-thought prompt template
    cot_template = PromptTemplate(
        template="""To solve this problem, think through it step-by-step:

{steps}

Problem: {problem}

Now, let's work through each step systematically:""",
        input_variables=["___", "___"]  # Replace ___ with: "steps", "problem"
    )

    # TODO 3: Apply chain-of-thought to the problem
    cot_prompt = cot_template.format(
        steps=reasoning_steps,
        problem="___"  # Replace ___ with: "Fix our data retention policy to comply with GDPR"
    )

    print("\n🔄 Applying Chain-of-Thought Reasoning")
    print("-" * 40)

    # Get AI response with chain-of-thought
    cot_response = llm.invoke(cot_prompt)
    print(f"\n📥 Chain-of-Thought Response:\n{cot_response.content[:500]}...")

    # Analyze the improvement
    print("\n📊 Comparison Analysis:")
    print(f"Direct response length: {len(direct_response.content)} characters")
    print(f"CoT response length: {len(cot_response.content)} characters")

    # Check for structured thinking
    has_steps = any(f"Step {i}" in cot_response.content for i in range(1, 6))
    has_analysis = "requirement" in cot_response.content.lower()
    has_recommendations = "recommend" in cot_response.content.lower()

    print(f"\n✅ Chain-of-Thought Benefits:")
    print(f"  ✓ Structured approach: {has_steps}")
    print(f"  ✓ Thorough analysis: {has_analysis}")
    print(f"  ✓ Clear recommendations: {has_recommendations}")

    # Key takeaways
    print("\n💡 Chain-of-Thought Best Practices:")
    print("  ✓ Break complex problems into steps")
    print("  ✓ Guide the AI's thinking process")
    print("  ✓ Ensure comprehensive analysis")
    print("  ✓ Get detailed, reasoned responses")

    # Create marker for completion
    os.makedirs("/root/markers", exist_ok=True)
    with open("/root/markers/task4_complete.txt", "w") as f:
        f.write("COMPLETED")

    print("\n✅ Task 4 completed! Chain-of-thought prompting mastered!")

if __name__ == "__main__":
    main()