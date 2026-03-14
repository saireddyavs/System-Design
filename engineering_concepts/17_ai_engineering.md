# Module 17: AI Engineering

---

## 1. How AI Agents Work

### Definition
AI agents are autonomous systems that perceive their environment, reason about goals, and take actions to achieve them. Unlike chatbots that only generate text, agents change the world through tools—calling APIs, executing code, browsing the web, and sending emails.

### Problem It Solves
Traditional automation (RPA) follows predetermined steps. When a website redesigns its UI overnight, the script breaks. Agents navigate ambiguity: given a goal like "find the cheapest flight to Tokyo," they figure out the steps, adapt when things fail, and complete the task.

### How It Works

**ReAct Pattern (Reasoning + Acting):**
```
1. Reason: "Tokyo, under $500, flexible dates. I'll search flights first."
2. Act: Call flight search API with parameters
3. Observe: "12 flights found. Six within budget."
4. Reason: "Check layover times, filter for under 20 hours."
5. Act: Filter results or call another tool
6. ... continues until goal reached
```

**Core Components:**
- **LLM as processor**: Not a knowledge base—it reasons, proposes steps, interprets results. Facts come from tools and memory.
- **Tools**: Defined functions (search_flights, execute_booking) the agent calls via structured JSON. External code validates, executes, returns results.
- **Memory**: Short-term = context window. Long-term = RAG over vector DB (company policies, user preferences).
- **Planning**: ReAct for uncertain paths; Plan-and-Execute when steps are known upfront.

**Agent Architectures:**
```
Single Agent:     One LLM with tools, iterates until done
Multi-Agent:      Specialized agents (researcher, writer, critic) collaborate
Hierarchical:     Manager agent delegates to worker agents
```

### Visual
```
                    ┌─────────────────────────────────────────┐
                    │              AGENT LOOP                  │
                    └─────────────────────────────────────────┘
                                      │
    ┌─────────────────────────────────┼─────────────────────────────────┐
    │                                 ▼                                 │
    │  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐   │
    │  │  REASON  │───→│   ACT    │───→│ OBSERVE  │───→│  REASON   │   │
    │  │ (LLM)    │    │ (Tool    │    │ (Result  │    │ (Next     │   │
    │  │          │    │  call)   │    │  back)   │    │  step?)   │   │
    │  └──────────┘    └──────────┘    └──────────┘    └─────┬─────┘   │
    │       ▲                                                   │       │
    │       └───────────────────────────────────────────────────┘       │
    │                         (loop until goal)                          │
    └───────────────────────────────────────────────────────────────────┘
```

### Tradeoffs

| Pros | Cons |
|------|------|
| Adapts to changing environments | Non-deterministic; same input ≠ same output |
| Handles multi-step tasks autonomously | Context collapse in long tasks |
| Can recover from failures with retry logic | Requires careful tool scoping (rate limits, allow lists) |
| More robust than brittle RPA scripts | Runaway errors without verification checkpoints |

### Real Systems
LangChain, AutoGPT, CrewAI, Claude Code, Microsoft Copilot Agent Mode

### Interview Tip
"An agent is a chatbot that acts. The key differentiator is tool use—agents call APIs, execute code, and change state. ReAct (Reason + Act) is the core loop. For production, add verification checkpoints before irreversible actions like payments."

---

## 2. AI Coding Workflow

### Definition
AI coding assistants (Copilot, Cursor, Cody) augment developers with code generation, completion, explanation, and refactoring. They work best as an iterative loop—context first, plan before code, small steps, verify—not as one-shot "paste problem, get solution."

### Problem It Solves
Models confidently suggest code that calls non-existent functions, uses wrong library versions, or ignores constraints. Without proper workflow, AI output looks polished but fails at runtime. The fix isn't "prompt better"—it's a repeatable process that keeps the model on a short leash.

### How It Works

**The Workflow Loop:**
```
Context → Plan → Code → Review → Test → Iterate
```

1. **Context**: Share README, AGENTS.md/CLAUDE.md (rules, constraints), relevant source files. "AI is a teammate who joined five minutes ago—brief them."
2. **Plan**: Ask for strategy before code. "Do not write code until I say approved." Fix plans are cheaper than fixing code.
3. **Code**: Implement one step at a time. Small changes are reviewable and reversible.
4. **Review**: Use AI code review tools (CodeRabbit) + human pass. Flag logic errors, security issues, edge cases.
5. **Test**: Run tests; have AI generate regression tests for bug fixes.
6. **Iterate**: Debug failures, refine request, repeat.

**RAG over Codebase:**
- Embeddings index codebase; retrieve relevant chunks for context
- Semantic search finds related code by meaning, not just keywords
- Chunking strategies: by file, by function, with overlap for continuity

**Prompt Engineering for Code:**
- Role assignment: "You are a senior engineer..."
- Few-shot: Show input/output examples
- Constraints: "Follow conventions in README. Do not modify other files unless necessary."

**Evaluation:**
- Code correctness: Unit tests, integration tests
- Security: Static analysis, dependency checks, prompt injection tests

### Visual
```
┌─────────────────────────────────────────────────────────────────┐
│                    AI CODING WORKFLOW                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   ┌─────────┐   ┌─────────┐   ┌─────────┐   ┌─────────┐        │
│   │ Context │──→│  Plan   │──→│  Code   │──→│ Review   │        │
│   │ README  │   │ (approve│   │ (small  │   │ + Test   │        │
│   │ Rules   │   │  first) │   │  steps) │   │          │        │
│   │ Source  │   └─────────┘   └─────────┘   └────┬─────┘        │
│   └─────────┘                                    │              │
│        ▲                                         │              │
│        └─────────────────────────────────────────┘              │
│                      Iterate until solid                         │
└─────────────────────────────────────────────────────────────────┘
```

### Tradeoffs

| Pros | Cons |
|------|------|
| Faster scaffolding, refactoring, test generation | AI invents APIs that don't exist—verify with docs |
| Catches edge cases in review | Long chats cause context drift |
| Plan-first reduces wasted code | Requires discipline; easy to skip steps |
| Multi-role (Planner, Implementer, Tester) improves output | Security-sensitive code needs extra scrutiny |

### Interview Tip
"Treat AI output as a draft, not an answer. The workflow: context first (README, rules, relevant files), plan before code, implement in small steps, review and test. Never trust AI-generated code for payments or auth without verification."

---

## 3. What Is Context Engineering?

### Definition
Context engineering is the practice of designing the right information for the LLM to see at each step—not just crafting the instruction, but assembling the full input: system prompts, few-shot examples, retrieved documents, tool outputs, and conversation history. Andrej Karpathy: "The delicate art and science of filling the context window with just the right information for the next step."

### Problem It Solves
When an AI gives a wrong or confused answer, the issue often isn't phrasing—it's missing, buried, or irrelevant information. Context rot: as agents run many steps, useful info gets buried under outdated details. Model performance degrades as context grows; more context isn't always better.

### How It Works

**Anatomy of Context:**
```
┌─────────────────────────────────────────────────────────────┐
│  System Prompt   → Model behavior, rules, persona            │
│  User Prompt     → Current question + conversation history   │
│  Examples        → Few-shot input/output pairs               │
│  Tools           → Tool definitions + tool outputs           │
│  Retrieved Data  → RAG chunks, documents                     │
└─────────────────────────────────────────────────────────────┘
         All compete for limited context window space
```

**Context Window Management:**
- Token budgeting: Allocate space for system, history, retrieval, response
- Chunking strategies: Semantic chunks, overlap for continuity, max chunk size
- Compaction: Summarize when full; keep goals, constraints, decisions; drop used tool outputs
- Structured note-taking: External store for goal, constraints, progress; retrieve when needed

**Retrieval Strategies:**
- **Loading upfront**: RAG—retrieve before generation. Good when query is clear.
- **Just-in-time**: Retrieve as task unfolds. Cleaner context, more round-trips.
- **Progressive disclosure**: Start with summaries, drill down into relevant detail.
- **Hybrid**: Baseline upfront + retrieve as needed.

**RAG as Context Engineering:**
RAG fetches relevant chunks from a vector DB and injects them into context. It's a core context engineering technique—dynamically assembling what the model needs at inference time.

### Visual
```
┌────────────────────────────────────────────────────────────────┐
│                    CONTEXT WINDOW (limited tokens)               │
├────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │ System       │  │ Examples     │  │ Message      │         │
│  │ Prompt       │  │ (few-shot)   │  │ History      │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │ Tool defs    │  │ RAG chunks   │  │ Current      │         │
│  │ + outputs    │  │ (retrieved)   │  │ user input   │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
│                                                                 │
│  Context Engineering = curating what goes here, when            │
└────────────────────────────────────────────────────────────────┘
```

### Tradeoffs

| Pros | Cons |
|------|------|
| Right info at right step improves accuracy | Complex to implement; many moving parts |
| Reduces hallucination via grounding | Token budget forces tradeoffs |
| Enables long-horizon agent tasks | Compaction can lose important details |
| RAG + memory + state = production-ready agents | Retrieval quality limits output quality |

### Interview Tip
"Context engineering is about what the model sees, not just how you ask. It includes system prompts, examples, RAG retrieval, tool outputs, and conversation management. At scale, context engineering matters more than prompt phrasing—you need memory, RAG, and state management."

---

## 4. Context Engineering vs Prompt Engineering

### Definition
- **Prompt Engineering**: Crafting the instruction—the script you hand the model. Role, few-shot examples, chain-of-thought, constraints.
- **Context Engineering**: Crafting the full input—instruction + context + examples + tools + retrieved data. Building the stage, props, and dossier before the model speaks.

### Problem It Solves
User says "Book me a hotel in Paris for the DevOps conference." Prompt engineering alone: agent books Paris, Kentucky—no way to know which Paris without user location, conference details, or company policy. The fix is context engineering: dynamically assembling the right information at runtime.

### How It Works

**Prompt Engineering (Static):**
- Role assignment, few-shot, CoT, constraint setting
- Optimizes the script
- Works for demos; hits wall in production when info is missing

**Context Engineering (Dynamic):**
```
User Input → [Memory + RAG + State + Tools] → Assembled Context → LLM
```
- **Memory**: Short-term (last N turns) + long-term (semantic retrieval of user facts)
- **RAG**: Company policies, product docs, real-time data
- **State**: Workflow phase, constraints met/pending
- **Tools**: Function schemas; orchestrator executes, returns to context

### Comparison Table

| Aspect | Prompt Engineering | Context Engineering |
|--------|-------------------|-------------------|
| Focus | Instruction wording | Full input assembly |
| Scope | Single prompt | System architecture |
| When it helps | Format, style, reasoning | Missing information, grounding |
| Scale | Demo, simple tasks | Production, agents |
| Involves | Writing, examples | Data pipelines, vector DBs, state machines |
| "Paris, Kentucky" fix | No | Yes (inject user location, conference data) |

### Why Context Engineering Matters at Scale
- Prompts are static; context is dynamic
- Production agents need user data, policies, real-time info
- Token budget forces prioritization—context engineering decides what to include
- Memory + RAG + state + tools = orchestration layer, not just better prompts

### Tradeoffs

| Pros | Cons |
|------|------|
| Prompt eng: Quick to iterate, no infra | Prompt eng: Can't fix missing data |
| Context eng: Solves production failures | Context eng: Complex, requires pipelines |
| Context eng: Enables multi-step agents | Both needed; context eng scales further |

### Interview Tip
"Prompt engineering optimizes the instruction. Context engineering assembles everything the model receives—memory, RAG, state, tools. When the model hallucinates or makes wrong assumptions, it's usually a context problem: the right info wasn't in the context. At scale, context engineering is the bottleneck."

---

## 5. How ChatGPT Apps Work

### Definition
ChatGPT Apps are interactive widgets that render inside ChatGPT conversations. Unlike Plugins (API wrappers returning text), Apps render full UI—maps, booking forms, spreadsheets—in sandboxed iframes. Users interact directly with the app's UI; the app can call tools and continue the conversation.

### Problem It Solves
Plugins returned text that ChatGPT had to interpret and re-present—no native UI, no interactivity. GPTs are hard to share and don't give brands control. Apps provide rich, interactive experiences: "Find me a flight" → browse options, compare, book—all inside the chat.

### How It Works

**Architecture:**
```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  Frontend   │────→│  ChatGPT    │────→│  MCP Server  │
│  (User)     │     │  (Host)     │     │  (Backend)   │
└─────────────┘     └──────┬──────┘     └──────────────┘
                          │
                          │ Renders
                          ▼
                   ┌─────────────┐
                   │  Widget     │
                   │  (iframe)   │
                   └─────────────┘
```

**Components:**
1. **MCP Server**: Backend. Exposes tools (search_restaurants) and resources (UI templates). Receives tool calls from ChatGPT, returns data + output template reference.
2. **Widget**: Frontend. HTML/React in sandboxed iframe. Receives `window.openai` bridge: toolOutput, callTool(), sendFollowUpMessage(), setWidgetState().
3. **ChatGPT**: Host. Decides when to call tools, fetches and renders widgets, manages conversation.

**Flow:**
1. User: "Find Italian restaurants in Brooklyn"
2. ChatGPT matches tool `search_restaurants`, calls MCP server with params
3. Server returns content (for ChatGPT) + structuredContent (for widget) + outputTemplate (ui://widget/restaurant-list.html)
4. ChatGPT fetches HTML, renders in iframe, injects structuredContent
5. User clicks restaurant → widget calls get_restaurant_details via callTool() or sendFollowUpMessage()

**Token Streaming (SSE):**
Responses stream token-by-token via Server-Sent Events. Lowers time-to-first-token; user sees progress.

**Safety Layers:**
- Content filtering before/after generation
- RLHF for alignment
- Sandboxed iframe: no DOM access, no cookies, CSP for network

### Visual
```
┌─────────────────────────────────────────────────────────────────────┐
│                    CHATGPT APP STACK                                  │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  User Message                                                        │
│       │                                                              │
│       ▼                                                              │
│  ┌─────────────┐   tool call    ┌─────────────┐   HTTPS    ┌──────┐ │
│  │  ChatGPT    │ ─────────────→ │  MCP Server │ ←───────── │ APIs │ │
│  │  (Model +   │                │  (Your      │            └──────┘ │
│  │   Orchestr.)│                │   Backend)  │                     │
│  └──────┬──────┘                └──────┬──────┘                     │
│         │                              │                             │
│         │ render widget                │ return data + template       │
│         ▼                              │                             │
│  ┌─────────────┐                       │                             │
│  │  Widget     │ ←── structuredContent ─┘                             │
│  │  (iframe)   │     toolOutput, callTool(), sendFollowUpMessage()    │
│  └─────────────┘                                                      │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### Tradeoffs

| Pros | Cons |
|------|------|
| Rich UI inside chat | One widget per message; sequential for multi-app requests |
| Direct user interaction with app | Sandbox limits: no cookies, strict CSP |
| MCP = standard protocol; reusable | Security: validate requests from ChatGPT |
| Tool + widget = show, do, or know | Widget can't access ChatGPT DOM |

### Interview Tip
"ChatGPT Apps = MCP server (tools + resources) + Widget (iframe UI). ChatGPT orchestrates tool calls; widget renders and can call tools or send follow-up messages. Key difference from Plugins: Apps render native UI and users interact directly. MCP is the protocol; OpenAI adopted Anthropic's standard."

---

## 6. LLM Concepts, Simply Explained

### Definition
LLMs are large autocomplete systems: given a sequence of text, predict the next token. They're built from transformers (attention mechanism), trained on massive text, and adapted via fine-tuning and RLHF. Key concepts: tokenization, embeddings, temperature, RAG, hallucination, context window, inference optimization.

### Problem It Solves
Understanding these concepts helps you: choose models, tune parameters, debug failures, and design systems. Without the vocabulary, it's hard to know what to fix when outputs are wrong, slow, or inconsistent.

### How It Works

**Core Concepts:**
- **Token**: Smallest unit (word, subword, punctuation). Tokenizer splits text → IDs. BPE, SentencePiece common.
- **Embedding**: Vector representing token meaning. Similar words close in latent space.
- **Transformers / Attention**: Mechanism to weigh relevance of each token to others. Enables long-range dependencies.
- **Temperature**: Controls randomness. 0 = deterministic; higher = more varied.
- **Top-k / Top-p**: Sampling strategies. Top-k: sample from k most likely tokens. Top-p (nucleus): from smallest set whose cumulative probability ≥ p.

**Fine-tuning vs RAG vs Prompt Engineering:**
- **Fine-tuning**: Train on specific data; changes model weights. Best for style, format, domain.
- **RAG**: Retrieve docs, inject into context. Best for facts, policies, real-time data.
- **Prompt engineering**: Craft instruction. Best for format, reasoning (CoT).

**Embeddings & Vector Similarity:**
- Text → embedding model → vector
- Similar meaning → similar vectors (cosine similarity, dot product)
- Powers semantic search, RAG retrieval

**Hallucination & Grounding:**
- Hallucination: Model generates plausible but false info
- Grounding: Force output to be based on provided sources. RAG implements grounding.

**Context Window & Token Limits:**
- Max tokens model can process. 128K–1M+ for modern models.
- Exceeding = truncation or summarization. Attention degrades with length.

**Inference Optimization:**
- **Quantization**: Reduce precision (FP16→INT8) for smaller, faster models
- **Batching**: Process multiple requests together for throughput
- **KV Cache**: Cache key-value pairs from previous tokens; avoid recomputation
- **Speculative decoding**: Small model drafts; large model verifies; speeds generation

### Visual
```
┌─────────────────────────────────────────────────────────────────────┐
│                    LLM INFERENCE PIPELINE                             │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  "What is fine-tuning?"                                              │
│       │                                                              │
│       ▼ Tokenize                                                      │
│  [What] [is] [fine] [-] [tuning] [?]  →  [1023, 318, 5621, ...]     │
│       │                                                              │
│       ▼ Embed                                                         │
│  Vectors in latent space                                             │
│       │                                                              │
│       ▼ Transformer (Attention)                                       │
│  Predict next token (autoregressive)                                  │
│       │                                                              │
│       ▼ Sample (temperature, top-k, top-p)                          │
│  "Fine-tuning" → "is" → "the" → "process" → ...                     │
│       │                                                              │
│       ▼ Decode                                                        │
│  "Fine-tuning is the process of training a pre-trained model..."     │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### Tradeoffs

| Pros | Cons |
|------|------|
| RAG: No retraining; fast to update knowledge | RAG: Retrieval quality limits accuracy |
| Fine-tuning: Custom behavior, style | Fine-tuning: Cost, data, overfitting risk |
| Low temp: Consistent, factual | Low temp: Repetitive, less creative |
| Quantization: Cheaper, faster | Quantization: Slight quality loss |

### Interview Tip
"LLM = next-token predictor. Key knobs: temperature (randomness), context window (memory), RAG (grounding). For production: use RAG for facts, fine-tune for style, prompt for format. Hallucination is reduced by grounding—only answer from provided context."

---

## 7. MCP - A Deep Dive (Model Context Protocol)

### Definition
MCP is an open standard (Anthropic, adopted by OpenAI) for connecting LLMs to external tools and data. It solves the N×M integration problem: instead of each AI app building custom integrations for each data source, you build N clients + M servers. One protocol, universal connectivity.

### Problem It Solves
Every AI needs context: codebase, database, APIs. Traditionally, each connection = custom integration. Cursor adds Notion → build from scratch. Copilot adds Notion → build again. N×M integrations. MCP reduces this to N+M: one client per AI app, one server per data source.

### How It Works

**Architecture:**
```
┌─────────────────────────────────────────────────────────────────┐
│  HOST (Claude Desktop, VSCode, Cursor, custom agent)             │
│       │                                                          │
│       │ creates                                                  │
│       ▼                                                          │
│  MCP CLIENT (per server)  ←──protocol──→  MCP SERVER             │
│       │                                    │                     │
│       │ tools/call                         │ connects to          │
│       │ resources/read                     │ PostgreSQL,         │
│       │ prompts/list                       │ GitHub, Slack, etc. │
└─────────────────────────────────────────────────────────────────┘
```

**Three Primitives (Servers Expose):**
1. **Resources**: Read-only data. DB results, files, API responses. AI reads, doesn't change.
2. **Tools**: Actions. Execute code, call APIs. AI invokes with params.
3. **Prompts**: Reusable instruction templates with variables. Parametrized prompts.

**Flow: Discovery → Schema → Invocation:**
1. Client connects to server
2. Client asks: "What can you do?" (capability negotiation)
3. Server returns: list of tools, resources, prompts with schemas
4. Client (or host) invokes: tools/call, resources/read
5. Server executes, returns result
6. Result injected into LLM context

**Compare with Function Calling, Plugins:**
- **Function calling**: Model-specific (OpenAI, Anthropic). Tightly coupled.
- **Plugins**: App-specific (ChatGPT Plugins). Deprecated.
- **MCP**: Standard protocol. Any client + any server. Dynamic discovery. Decouples intelligence from data.

### Visual
```
┌─────────────────────────────────────────────────────────────────────┐
│                    MCP: BEFORE vs AFTER                               │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  BEFORE (N×M):                    AFTER (N+M):                       │
│                                                                      │
│  Cursor ──┬── GitHub (custom)      Cursor ──┬── MCP Client ──┐       │
│           ├── Notion (custom)               │                │       │
│           └── DB (custom)                   └────────────────┼──→  │
│  Copilot ─┬── GitHub (custom)      Copilot ──┬── MCP Client ──┼──→  │
│           ├── Notion (custom)               │                │  MCP │
│           └── DB (custom)                   └────────────────┼──→  │
│  Claude ──┬── GitHub (custom)     Claude ───┬── MCP Client ──┘     │
│           ├── Notion (custom)               │                       │
│           └── DB (custom)                   └──────────────────────│
│                                                                      │
│  N×M integrations                    N clients + M servers          │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### Tradeoffs

| Pros | Cons |
|------|------|
| N+M vs N×M; one integration per side | New standard; ecosystem still growing |
| Dynamic discovery; add tool, AI finds it | Requires MCP server implementation |
| Decouples model from data; swap either | Transport choices: stdio, HTTP, SSE |
| Open standard; Anthropic + OpenAI | Learning curve for server authors |

### Interview Tip
"MCP is like USB-C for AI: one protocol for connecting any LLM to any data source. Servers expose Resources (read), Tools (actions), Prompts (templates). Client discovers capabilities at connect time. Reduces N×M custom integrations to N+M. Adopted by OpenAI for ChatGPT Apps."

---

## 8. What Is Reinforcement Learning

### Definition
Reinforcement Learning (RL) is a type of machine learning where an agent learns to maximize cumulative reward by interacting with an environment. The agent takes actions, receives rewards or penalties, and updates its policy. No labeled data—learning from trial and error.

### Problem It Solves
Supervised learning needs labeled examples. For complex behaviors (helpful, harmless, human-preferred), we can't easily label every response. RL learns from feedback: "this is better than that." Used to align LLMs (RLHF) and train game-playing and robotic agents.

### How It Works

**Core Elements:**
- **Agent**: Entity that acts (LLM, robot, game player)
- **Environment**: Everything outside the agent. Responds to actions, returns state and reward
- **State**: Snapshot of environment. What the agent sees to decide
- **Action**: What the agent does
- **Reward**: Signal (positive/negative) indicating how good the action was

**Algorithms:**
- **Q-learning**: Learn value of (state, action) pairs. Pick action with highest Q-value.
- **Policy gradient**: Directly optimize policy to maximize expected reward.
- **PPO (Proximal Policy Optimization)**: Stable policy gradient; used in RLHF.
- **DPO (Direct Preference Optimization)**: Simpler alternative to RLHF; no reward model.

**RLHF (Reinforcement Learning from Human Feedback):**
How ChatGPT was trained to be helpful and safe:
1. Generate multiple answers to a prompt
2. Humans rank them (best to worst)
3. Train reward model to predict human preference
4. Use reward model to train LLM (e.g., PPO): nudge parameters toward higher-reward outputs

### Visual
```
┌─────────────────────────────────────────────────────────────────────┐
│                    RL LOOP                                            │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│       ┌─────────┐                                                    │
│       │  Agent  │                                                    │
│       │  (LLM)  │                                                    │
│       └────┬────┘                                                    │
│            │ action                                                  │
│            ▼                                                         │
│       ┌─────────┐     reward + new state                              │
│       │Environment│ ───────────────────────→  Agent updates policy   │
│       │(users,   │                           (e.g., PPO)             │
│       │ feedback)│                                                    │
│       └─────────┘                                                    │
│            │                                                         │
│            └─────────────────────────────────────────────────────────│
│                         (loop)                                        │
│                                                                      │
│  RLHF: Reward = human preference model trained on rankings           │
└─────────────────────────────────────────────────────────────────────┘
```

### Tradeoffs

| Pros | Cons |
|------|------|
| Learns from feedback, not labels | Reward design is hard; wrong reward = wrong behavior |
| RLHF aligns LLMs to human preference | Expensive: human raters, compute |
| Enables complex behaviors (helpful, safe) | Can over-optimize reward (reward hacking) |
| DPO simplifies; no reward model | PPO more stable but complex |

### Interview Tip
"RL = agent learns from rewards/penalties in an environment. RLHF trains LLMs: humans rank answers, train reward model, use PPO to optimize LLM toward higher reward. That's how ChatGPT got helpful and safe. DPO is a simpler alternative without a separate reward model."

---

## Comprehensive Comparison Table

### Fine-tuning vs RAG vs Prompt Engineering vs Context Engineering

| Approach | What It Changes | When to Use | Cost | Latency | Accuracy |
|----------|-----------------|-------------|------|---------|----------|
| **Prompt Engineering** | Instruction wording, examples, format | Format, style, reasoning (CoT), simple tasks | Very low | No impact | Limited by model + context |
| **Context Engineering** | Full input assembly (memory, RAG, state, tools) | Production agents, multi-step tasks, grounding | Medium (retrieval, infra) | Retrieval adds latency | High when context is right |
| **RAG** | Injects retrieved docs into context | Facts, policies, real-time data, knowledge not in training | Low–medium (embeddings, vector DB) | Retrieval latency | Depends on retrieval quality |
| **Fine-tuning** | Model weights | Domain style, format, specific behavior patterns | High (compute, data) | No inference impact | Good for trained patterns |

### When to Use Each

| Scenario | Best Approach |
|----------|---------------|
| Need consistent JSON output format | Prompt engineering (few-shot) |
| Need model to use company policy | RAG |
| Need model to remember user preferences across sessions | Context engineering (memory) |
| Need model to follow strict coding style | Fine-tuning |
| Need model to call APIs, search, book | Context engineering (tools) |
| Need model to reason step-by-step | Prompt engineering (CoT) |
| Need up-to-date info (prices, news) | RAG |
| Need custom persona/tone | Fine-tuning or prompt (role) |

### Cost, Latency, Accuracy Tradeoffs

```
                    Cost
                     ▲
                     │     Fine-tuning
                     │        •
                     │
                     │              Context Eng
                     │                 •
                     │
                     │   RAG     Prompt Eng
                     │    •         •
                     └──────────────────────→ Latency
                     
Accuracy: Context Eng + RAG > Fine-tuning (for facts) > Prompt only
```

| Approach | Typical Cost | Latency Impact | Accuracy Driver |
|----------|--------------|----------------|-----------------|
| Prompt Eng | ~$0 | None | Instruction clarity |
| RAG | $0.01–0.1/query (embeddings + retrieval) | +50–200ms | Retrieval relevance |
| Context Eng | Infra + retrieval | +variable | Right info at right step |
| Fine-tuning | $100s–$1000s one-time | None | Training data quality |

### Interview Tip
"For facts and real-time data: RAG. For format and style: prompt engineering or fine-tuning. For production agents: context engineering—memory, RAG, state, tools. Fine-tuning when you need deep behavioral change; RAG when you need knowledge. Often combine: RAG for grounding + prompt for format + context eng for assembly."
