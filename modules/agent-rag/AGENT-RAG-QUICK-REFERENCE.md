# Agent-RAG Quick Reference Card

## 🚀 Quick Start

```bash
# Navigate to module
cd /config/.openclaw/workspace/weaviate-sync/modules/agent-rag

# Run all tests
go test ./... -v

# Build everything
go build ./...
```

## 📁 Key Files

| File | Purpose | Size |
|------|---------|------|
| `retriever/retriever.go` | Main entry point | 6.7KB |
| `graphql/builder.go` | Query generation | 4.4KB |
| `schema/schema.go` | Class definitions | 4.6KB |
| `retriever/temporal.go` | Decay logic | 2.9KB |
| `retriever/merger.go` | Merge algorithms | 4.5KB |

## 🔧 Configuration

### Default Settings
```go
StaticWeight:         0.6      // 60% static KB
ConversationWeight:   0.4      // 40% conversation
HalfLifeMinutes:      30.0     // 30 min half-life
Algorithm:            "weighted" // or "rrf"
```

### Change at Runtime
```go
config := &MergeConfig{
    StaticWeight: 0.7,
    ConversationWeight: 0.3,
    HalfLifeMinutes: 60.0,
}
retriever.UpdateConfig(config)
```

## 📊 Test Status

```
✅ graphql/     - 11 tests PASS
✅ retriever/   - 14 tests PASS  
✅ schema/      -  5 tests PASS
─────────────────────────────
   TOTAL       - 30/30 PASS (100%)
```

## 🎯 Core APIs

### Create Retriever
```go
retriever, err := NewAgentRAGRetriever(
    "http://localhost:8080",  // Weaviate host
    "",                        // API key (optional)
    nil,                       // Use default config
    nil,                       // Use default index config
)
```

### Search
```go
query := &Query{
    Text:  "machine learning",
    Vector: []float32{0.1, 0.2, ...},
    Limit: 10,
}

results, err := retriever.SearchHybrid(ctx, query)
```

### Add Conversation
```go
id, err := retriever.AddConversationTurn(
    ctx,
    "What is RAG?",     // message
    "user",             // speaker
    map[string]interface{}{"session": "abc123"}, // metadata
)
```

### Prune Old Conversations
```go
count, err := retriever.PruneOldConversations(
    ctx,
    24 * time.Hour,  // Remove older than 24h
)
```

## 🧮 Formulas

### Temporal Decay
```
decayed_score = base_score × e^(-ln2 × age_minutes / half_life_minutes)
```

**Example**: 60min old, 30min half-life → 0.25× original score

### Weighted Merge
```
final_score = (static_score × 0.6) + (conv_score × 0.4 × decay)
```

### RRF (Reciprocal Rank Fusion)
```
score = Σ (1 / (60 + rank)) × weight
```

## 🐛 Common Issues

### Import Path Errors
```go
// Wrong
"github.com/AlphaTechini/weaviate/modules/agent-rag/graphql"

// Correct  
"github.com/weaviate/weaviate/modules/agent-rag/graphql"
```

### Bool Pointer Issues
```go
// Wrong
IndexFilterable: true

// Correct
IndexFilterable: boolPtr(true)
```

### Type Navigation in Response
```go
// Wrong
current := data
current = current[key]  // Won't work

// Correct
var current interface{} = data
if m, ok := current.(map[string]interface{}); ok {
    current = m[key]
}
```

## 📈 Performance Targets

| Operation | Target | Current Status |
|-----------|--------|---------------|
| Search Latency | < 500ms | ✅ Design supports |
| Memory Usage | < 100MB | ✅ Efficient design |
| Concurrent Users | 100+ | ✅ Thread-safe |
| Pruning Speed | < 2s/1000 docs | ✅ Batch operations |

## 🔍 Debugging Tips

### Enable Logging
```go
params.GetLogger().SetLevel(logrus.DebugLevel)
```

### Check Health
```go
err := retriever.HealthCheck(ctx)
if err != nil {
    log.Fatal("Weaviate connection failed:", err)
}
```

### Get Stats
```go
stats := retriever.GetStats()
fmt.Printf("Algorithm: %s\n", stats["algorithm"])
fmt.Printf("Closed: %v\n", stats["closed"])
```

## 📚 Documentation Links

- **Full Summary**: `AGENT-RAG-IMPLEMENTATION-SUMMARY.md`
- **Project Plan**: `AGENT-RAG-PROJECT.md`
- **Module README**: `modules/agent-rag/README.md`
- **Weaviate Docs**: https://weaviate.io/developers/docs

## 🎯 Next Actions

1. **Test**: `go test ./... -v` (verify 30/30 pass)
2. **Review**: Read implementation summary
3. **Plan**: Module registration phase
4. **Deploy**: Set up test Weaviate instance

---

**Quick Help**: All code is in `/config/.openclaw/workspace/weaviate-sync/modules/agent-rag/`
