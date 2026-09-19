# 🌟 Animus (`github.com/Flokey82/animus`)

**Animus** is a modular, lightweight autonomous agent runtime harness in Go. It formalizes and extracts the generic cognitive architecture powering embodied AI companions and narrative simulations:

- 🧠 **Dual-Model LLM Gateway**: Seamless support for local OpenAI-compatible backends (Lemonade, Ollama, vLLM). Supports routing fast reflex tool calls to small models (e.g. `granite-4.0-h-tiny`) and deep thinking/conversations to primary models (`Gemma-4`, `Qwen3`, `Granite-4`), with automatic `<think>...</think>` sanitization.
- 🧩 **Dynamic Contextual Tool Registry**: The `AgentTool` protocol with tag-based filtering. Eliminates token waste and hallucination by exposing only tools relevant to the agent's current location, inventory, or social context.
- ⚡ **Homeostatic Drives Engine**: Generalized physiological and psychological motives (Maslow / The Sims inspired) with tick decay, critical thresholds, and urgency heuristics.
- 🎭 **OCEAN Personality & Expressed Traits**: Five-Factor personality evaluation deriving active behavioral traits and injecting prompt guidance.
- 📚 **Multi-Tier Memory Store & Episodic Logger**: Chronological episodic event ring-buffer, short-term impression decay, long-term semantic retrieval, and core identity beliefs.
- 🌙 **Oneiros Sleep & Consolidation Cycle**: End-of-day sleep cycle synthesizing vivid dream narratives from daily logs, consolidating memories, and reflecting on identity.
- ⏰ **Circadian Clock & Day/Night Scheduler**: Biological chronotypes (*Early Bird*, *Balanced*, *Night Owl*) with automatic day/night state management.
- 🌐 **Built-in Reusable Tools**: Out-of-the-box support for Wikipedia, Hacker News, Astronomical Moon Phase calculation, Free Web Search, and Clock/Calendar.
- 📡 **Multi-Channel Hub**: Decoupled message bus for streaming outputs to Terminal, Discord, Telegram, or WebSockets.

---

## 📦 Package Structure

```
animus/
├── pkg/
│   ├── llm/         # LLM Client, Dual-Model Router, Token Sanitizer
│   ├── tools/       # AgentTool interface, Registry, & Contextual Filter
│   │   └── builtin/ # Wikipedia, HackerNews, Moon Phase, Web Search, Clock
│   ├── drives/      # Homeostatic drives & motivation engine
│   ├── traits/      # OCEAN Five-Factor personality & trait prompt synthesis
│   ├── memory/      # Episodic ring-buffer & Multi-tier memory store
│   ├── dream/       # Oneiros sleep cycle & dream consolidation
│   ├── engine/      # Cognitive agent loop & circadian scheduler
│   └── channel/     # Channel abstractions & event broadcast hub
└── examples/
    └── minimal_agent/ # Interactive standalone CLI demonstration
```

---

## 🚀 Quick Start

### 1. Run the Demo Agent (Mock Mode)

```bash
go run ./examples/minimal_agent -mock
```

### 2. Run with Local Lemonade / Ollama Backend

```bash
go run ./examples/minimal_agent -url "http://192.168.86.208:8000/api/v1" -model "Gemma-4-26B-A4B-it-MTP-GGUF"
```

---

## 🛠️ Minimal Agent Code Example

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/Flokey82/animus/pkg/engine"
    "github.com/Flokey82/animus/pkg/llm"
    "github.com/Flokey82/animus/pkg/tools/builtin"
    "github.com/Flokey82/animus/pkg/traits"
)

func main() {
    agent := engine.NewAgent(engine.AgentConfig{
        Name:         "Astra",
        BaseIdentity: "A curious explorer of science and philosophy.",
        Personality:  traits.DefaultOCEAN(),
        LLMConfig: llm.Config{
            BaseURL: "http://192.168.86.208:8000/api/v1",
            Model:   "Gemma-4-26B-A4B-it-MTP-GGUF",
        },
    })

    // Register built-in tools
    agent.Tools.Register(
        builtin.NewClockTool(nil),
        builtin.NewWikipediaTool(),
        builtin.NewCelestialMoonTool(),
    )

    // Chat with the agent
    reply, _ := agent.Chat(context.Background(), "What is the current moon phase?", "User")
    fmt.Println(reply)
}
```

---

## 🧪 Testing

Run the full unit test suite:

```bash
go test -v ./...
```
