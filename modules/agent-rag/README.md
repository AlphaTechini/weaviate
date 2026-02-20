# Agent-RAG Module for Weaviate

**Agent-specific Retrieval Augmented Generation with conversation memory and temporal awareness**

## 🎯 Overview

This module extends Weaviate with **conversation-aware RAG** optimized for AI agents. Unlike standard RAG systems that treat each query independently, Agent-RAG maintains live conversation context and intelligently merges it with static knowledge bases.

## ✨ Key Features

- **🧠 Conversation Memory**: Live embedding of conversation turns as retrievable context
- **⏰ Temporal Weighting**: Recent conversations automatically weighted higher
- **🔀 Hybrid Search**: Combines vector similarity + keyword matching + metadata filters
- **⚡ Real-time Updates**: Conversation index updates with every message
- **🎛️ Configurable Decay**: Adjustable half-life for conversation relevance

## 🏗️ Architecture

```
User Query → [Query Processor]
                ↓
        ┌───────┴───────┐
        ↓               ↓
  [Static KB]    [Conversation Memory]
  (Vector +      (Vector + Temporal)
   BM25)         
        ↓               ↓
    [Weighted Merger]
    (60% static + 40% conv × decay)
            ↓
      [LLM Context]
```

## 📦 Installation

### From Source (Development)

```bash
# Clone your fork
git clone git@github.com:AlphaTechini/weaviate.git
cd weaviate

# Build with agent-rag module
go build -tags agent-rag -o weaviate ./cmd/weaviate
```

### Docker (Coming Soon)

```bash
docker run -p 8080:8080 alphatechini/weaviate:agent-rag
```

## 🚀 Quick Start

### 1. Enable the Module

Set environment variable when starting Weaviate:

```bash
export ENABLE_MODULES="agent-rag"
./weaviate --host localhost --port 8080
```

### 2. Create Conversation-Enabled Class

```graphql
mutation {
  CreateClass(class: {
    class: "Conversation"
    vectorizer: "text2vec-transformers"
    properties: [
      {
        name: "message"
        dataType: ["text"]
      }
      {
        name: "timestamp"
        dataType: ["date"]
      }
      {
        name: "speaker"
        dataType: ["text"]
      }
    ]
    moduleConfig: {
      "agent-rag": {
        "conversationMemory": true
        "temporalDecay": {
          "enabled": true
          "halfLifeMinutes": 30
        }
      }
    }
  })
}
```

### 3. Add Conversation Turns

```graphql
mutation {
  CreateObjects(objects: [
    {
      class: "Conversation"
      properties: {
        message: "What is the capital of France?"
        timestamp: "2026-02-20T14:00:00Z"
        speaker: "user"
      }
    }
    {
      class: "Conversation"
      properties: {
        message: "The capital of France is Paris."
        timestamp: "2026-02-20T14:00:05Z"
        speaker: "assistant"
      }
    }
  ])
}
```

### 4. Query with Conversation Context

```graphql
query {
  Get {
    KnowledgeBase(
      nearText: {
        concepts: ["Paris tourism"]
      }
      hybrid: {
        query: "Paris tourism"
        alpha: 0.5
      }
      limit: 5
    ) {
      title
      content
      _additional {
        score
      }
    }
    
    Conversation(
      nearText: {
        concepts: ["Paris tourism"]
      }
      where: {
        operator: GreaterThan
        path: ["timestamp"]
        valueDate: "2026-02-20T13:00:00Z"
      }
      limit: 5
    ) {
      message
      speaker
      timestamp
      _additional {
        score
        temporalWeight
      }
    }
  }
}
```

## ⚙️ Configuration

### Module Settings

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `conversationMemory.enabled` | bool | `true` | Enable conversation memory indexing |
| `temporalDecay.enabled` | bool | `true` | Enable temporal decay function |
| `temporalDecay.halfLifeMinutes` | int | `30` | Half-life for exponential decay |
| `merge.staticWeight` | float | `0.6` | Weight for static knowledge base (0-1) |
| `merge.conversationWeight` | float | `0.4` | Weight for conversation memory (0-1) |
| `merge.algorithm` | string | `"weighted"` | Merge algorithm: `weighted` or `rrf` |

### Example Configuration

```json
{
  "modules": {
    "agent-rag": {
      "conversationMemory": {
        "enabled": true,
        "maxAgeHours": 24,
        "autoPrune": true
      },
      "temporalDecay": {
        "enabled": true,
        "halfLifeMinutes": 30,
        "minWeight": 0.01
      },
      "merge": {
        "staticWeight": 0.6,
        "conversationWeight": 0.4,
        "algorithm": "weighted",
        "rrfK": 60
      }
    }
  }
}
```

## 🔧 Development

### Running Tests

```bash
cd modules/agent-rag
go test ./... -v
```

### Building Module

```bash
go build -buildmode=plugin -o agent-rag.so .
```

### Adding New Features

1. **Custom Merge Algorithms**: Implement in `merger/` directory
2. **Temporal Functions**: Add to `temporal/` package
3. **GraphQL Extensions**: Update `gql/` resolvers

## 📊 Performance Benchmarks

| Metric | Target | Current |
|--------|--------|---------|
| Retrieval Latency (p95) | < 500ms | TBD |
| Recall@10 | > 85% | TBD |
| Write Latency | < 50ms | TBD |
| Memory Overhead | < 100MB | TBD |

*Run benchmarks: `go test -bench=. -benchmem`*

## 🎯 Use Cases

### 1. Customer Support Chatbots
Maintain conversation context across multiple turns while accessing product documentation.

### 2. AI Coding Assistants
Remember recent code discussions while searching through API documentation.

### 3. Research Assistants
Track research questions and findings while querying academic papers.

### 4. Personal AI Agents
Long-term memory of user preferences with recent context prioritization.

## 🔒 Security Considerations

- **Data Isolation**: Each conversation encrypted and isolated
- **Access Control**: RBAC for conversation retrieval
- **Audit Logging**: All queries logged with timestamps
- **Auto-Pruning**: Configurable data retention policies
- **PII Detection**: Automatic redaction of sensitive information

## 🤝 Contributing

Contributions welcome! See main Weaviate contributing guidelines.

### Areas Needing Help:
- [ ] GraphQL query optimizer
- [ ] Advanced merge algorithms (ML-based)
- [ ] Multi-agent conversation support
- [ ] Dashboard/monitoring UI
- [ ] SDK implementations (Python, Node.js, Go)

## 📄 License

Same as Weaviate: BSD 3-Clause

## 🙏 Acknowledgments

Built on top of Weaviate's excellent modular architecture. Inspired by research on conversational RAG and temporal attention mechanisms.

---

**Module Status**: Alpha (Under Active Development)  
**Version**: 0.1.0  
**Maintainer**: @AlphaTechini
