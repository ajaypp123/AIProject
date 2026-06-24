
# 🎨 Prompt Engineering

Notes:
https://www.promptingguide.ai/

Master the art of AI communication through hands-on prompt engineering techniques!

## 📚 Lab Overview

This lab teaches you 4 powerful prompting techniques that form the foundation of effective AI interaction:

1. **Zero-Shot Prompting** - Direct instructions without examples
2. **One-Shot Prompting** - Learning from a single example
3. **Few-Shot Prompting** - Multiple examples for consistency
4. **Chain-of-Thought** - Step-by-step reasoning

## 🎯 Learning Objectives

By completing this lab, you will:
- ✅ Understand when to use each prompting technique
- ✅ Write effective prompts that get consistent results
- ✅ Control AI output format, tone, and style
- ✅ Solve complex problems with structured reasoning
- ✅ Compare techniques side-by-side for optimal selection

## 📊 Research-Based Performance Improvements

Based on 2024-2025 benchmark studies:

| Technique | Improvement | Use Case |
|-----------|------------|----------|
| **Zero-Shot → Specific** | **2-5%** accuracy gain | Quick, general tasks |
| **Zero-Shot → One-Shot** | **23%** improvement (25% → 48%) | Format learning |
| **Zero-Shot → Few-Shot** | **12.2%** accuracy boost | Style consistency |
| **Without CoT → With CoT** | **52%** improvement (26% → 78%) | Complex reasoning |

*Sources: GPQA Benchmark 2025, OpenAI Research, Academic Studies*

# Prompt Type

## Type 1: Zero-Shot Prompting

- Direct Instructions Without Examples
- Learn how to write clear, specific prompts that work without providing examples.

e.g.:
```
- Vague
write a data privacy policy.

- improved
Write a 200-word data privacy policy for European customers covering GDPR requirements, data retention periods of 30 days, and user rights to deletion and portability.
```

## Type 2: One-Shot Prompting

- Learning from a Single Example
- Provide one example for the AI to follow, ensuring consistent format and style.

Eg:
```
Here's an example of our policy format:
REFUND POLICY
1. Eligibility: Within 30 days of purchase
2. Conditions: Product unused and in original packaging
3. Process: Submit request via support@company.com
4. Timeline: Refund processed within 5-7 business day
5. Exceptions: Digital products and custom orders non-refundable
Now write a remote policy following this EXACT format with numbered sections
```

## Type 3: Few-Shot Prompting

- Learning from Multiple Examples
Provide multiple examples to teach the AI your specific pattern and style.

Eg:
```
Rewrite the facts below into the style of a mysterious, noir-style detective novel.

Fact: It started raining in the middle of the night.
Noir Style: The rain began to weep in the dead of night, washing away the city's sins but leaving behind the cold truth.

Fact: I lost my car keys somewhere on 5th Avenue.
Noir Style: My keys vanished somewhere along 5th Avenue, swallowed up by the shadows of a relentless city.

Fact: The suspect walked into a diner and ordered black coffee.
Noir Style: He slunk into the neon-lit diner, ordering a black coffee to drown out the ghosts of his past.

Fact: The witness was hiding in a dark alleyway.
Noir Style:
```

## Type 4: Chain-of-Thought Prompting

- Step-by-Step Reasoning
- Guide the AI through a logical thinking process to solve complex problems.

Eg:
```
The odd numbers in this group add up to an even number: 4, 8, 9, 15, 12, 2, 1.
A: Adding all the odd numbers (9, 15, 1) gives 25. The answer is False.
The odd numbers in this group add up to an even number: 17,  10, 19, 4, 8, 12, 24.
A: Adding all the odd numbers (17, 19) gives 36. The answer is True.
The odd numbers in this group add up to an even number: 16,  11, 14, 4, 8, 13, 24.
A: Adding all the odd numbers (11, 13) gives 24. The answer is True.
The odd numbers in this group add up to an even number: 17,  9, 10, 12, 13, 4, 2.
A: Adding all the odd numbers (17, 9, 13) gives 39. The answer is False.
The odd numbers in this group add up to an even number: 15, 32, 5, 13, 82, 7, 1.
A:
```
