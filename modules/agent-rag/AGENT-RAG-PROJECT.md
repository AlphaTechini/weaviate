# Agent-Specific RAG Project

## 🎯 Project Vision

Build a **specialized RAG system optimized for AI agents** that combines:
1. **Hybrid Search** (vector + keyword + metadata)
2. **Conversation Memory** (live context as retrievable embeddings)
3. **Temporal Weighting** (recent conversations weighted higher)
4. **Simple Architecture** (80% of benefits, 20% of complexity)

---

## 📋 Problem Statement

Current RAG systems are designed for static Q&A, not agent conversations:
- ❌ Treat each query independently (no conversation context)
- ❌ Static knowledge base only (no live memory)
- ❌ Pure vector search (misses exact terms, numbers, proper nouns)
- ❌ No temporal awareness (all context treated equally)

**Result**: Agents lose conversation coherence, forget recent context, and provide disjointed responses.

---

## 🏗️ Architecture Design

### Core Components

```
┌─────────────────┐
│   User Query    │
└────────┬────────┘
         │
         ▼
┌─────────────────────────────────────┐
│      Query Processor                │
│  - Generate embedding               │
│  - Extract keywords                 │
│  - Identify intent                  │
└────────┬────────────────────────────┘
         │
    ┌────┴────┐
    │         │
    ▼         ▼
┌─────────┐ ┌──────────────────┐
│ Static  │ │   Dynamic        │
│ Knowledge│ │  Conversation    │
│  Base   │ │    Memory        │
│         │ │   (Live)         │
└────┬────┘ └────────┬─────────┘
     │               │
     │   Vector      │   Vector
     │   Search      │   Search
     │               │
     ▼               ▼
┌─────────────────────────────────┐
│      Result Merger              │
│  - Weighted scoring             │
│  - Reciprocal Rank Fusion       │
│  - Temporal decay applied       │
│                                 │
│  final_score =                  │
│    (static_score × 0.6) +       │
│    (conv_score × 0.4 × decay)   │
└────────┬────────────────────────┘
         │
         ▼
┌─────────────────┐
│  LLM + Context  │
│   Generation    │
└─────────────────┘
```

---

## 🔧 Technical Implementation

### 1. Dual Index Strategy

**Static Index** (Knowledge Base)
- Documents, code, APIs, guides
- Updated periodically (daily/hourly)
- Hybrid search: vector + BM25 + metadata filters

**Dynamic Index** (Conversation Memory)
- Live conversation turns
- Real-time updates (every message)
- Vector search with temporal decay
- Automatic pruning (older than N hours/days)

### 2. Embedding Strategy

**Single Embedding Model** for compatibility:
- Use same model for both indices
- Ensures vectors are comparable
- Recommended: `text-embedding-3-small` or `mxbai-embed-large`

**Chunking Strategy**:
- Static docs: 512-1024 tokens with overlap
- Conversation: Per-message chunks (no splitting)
- Metadata: timestamp, speaker, message_type

### 3. Temporal Decay Function

```python
def apply_time_decay(score, timestamp, current_time, half_life_minutes=30):
    """Apply exponential decay to conversation scores"""
    time_diff = current_time - timestamp
    decay_factor = math.exp(-time_diff / (half_life_minutes * 60))
    return score * decay_factor
```

**Configurable half-life**:
- Short conversations: 15-30 minutes
- Long-running agents: 2-4 hours
- Persistent memory: 24+ hours

### 4. Weighted Merge Algorithm

**Option A: Simple Weighted Sum** (Recommended for MVP)
```python
def merge_results(static_results, conv_results, static_weight=0.6):
    merged = {}
    
    # Add all static results
    for doc_id, score in static_results.items():
        merged[doc_id] = score * static_weight
    
    # Add conversation results with weight + decay
    for doc_id, score, timestamp in conv_results:
        decay = apply_time_decay(1.0, timestamp, now())
        conv_weight = 0.4 * decay
        
        if doc_id in merged:
            merged[doc_id] += score * conv_weight
        else:
            merged[doc_id] = score * conv_weight
    
    # Sort by final score and return top K
    return sorted(merged.items(), key=lambda x: x[1], reverse=True)[:K]
```

**Option B: Reciprocal Rank Fusion** (Better quality, more complex)
```python
def rrf_merge(static_ranked, conv_ranked, k=60):
    scores = {}
    
    # Score by rank position
    for rank, (doc_id, _) in enumerate(static_ranked):
        scores[doc_id] = scores.get(doc_id, 0) + 1.0 / (k + rank)
    
    for rank, (doc_id, _, _) in enumerate(conv_ranked):
        decay = apply_time_decay(1.0, timestamp, now())
        scores[doc_id] = scores.get(doc_id, 0) + (1.0 / (k + rank)) * decay
    
    return sorted(scores.items(), key=lambda x: x[1], reverse=True)[:K]
```

### 5. Context Assembly

```python
def assemble_context(merged_results, max_tokens=4000):
    context_parts = []
    total_tokens = 0
    
    for doc_id, score in merged_results:
        chunk = get_chunk(doc_id)
        chunk_tokens = count_tokens(chunk.text)
        
        if total_tokens + chunk_tokens > max_tokens:
            break
        
        # Add source annotation
        if chunk.type == 'conversation':
            source = f"[Conversation {format_time(chunk.timestamp)}]"
        else:
            source = f"[Document: {chunk.source}]"
        
        context_parts.append(f"{source}\n{chunk.text}")
        total_tokens += chunk_tokens
    
    return "\n\n---\n\n".join(context_parts)
```

---

## 📊 Performance Targets

| Metric | Target | Measurement |
|--------|--------|-------------|
| **Latency p95** | < 500ms | End-to-end retrieval |
| **Recall@10** | > 85% | Relevant chunks in top 10 |
| **Precision@10** | > 70% | Top 10 are relevant |
| **Token Efficiency** | < 3000 tokens | Context window usage |
| **Memory Overhead** | < 100MB | RAM for conversation index |
| **Write Latency** | < 50ms | Adding new conversation turn |

---

## 🛠️ Tech Stack: DECIDED ✅

### **Weaviate Module Extension** (IN PROGRESS)

**Status**: Foundation laid - fork synced, module structure created

**What's Done**:
- ✅ Weaviate fork synced with upstream (958 commits)
- ✅ Module directory created: `modules/agent-rag/`
- ✅ Basic module structure initialized
- ✅ README with full documentation created

**Next Steps**:
1. Implement core retrieval logic in `retriever/` package
2. Add GraphQL extensions for dual-index queries
3. Implement temporal decay functions
4. Build weighted merge algorithm
5. Create integration tests

**Why This Wins**:
- Leverages existing Weaviate infrastructure
- Hybrid search already built-in
- Go-based (your expertise)
- Production-ready scaling from day 1
- Can ship MVP in 1-2 weeks

**Estimated Time**: 1-2 weeks for functional MVP

---

## 🚀 Development Phases

### Phase 1: Core Retrieval (Week 1-2)
- [ ] Set up dual indices (static + conversation)
- [ ] Implement hybrid search for static index
- [ ] Implement vector search for conversation index
- [ ] Build weighted merge algorithm
- [ ] Add temporal decay function
- [ ] Basic CLI for testing

**Deliverable**: Working prototype with synthetic data

### Phase 2: Integration & Optimization (Week 3-4)
- [ ] Integrate with real conversation streams
- [ ] Optimize merge algorithm (test different weights)
- [ ] Add metadata filtering support
- [ ] Implement context assembly with token limits
- [ ] Add caching layer for frequent queries
- [ ] Benchmark performance

**Deliverable**: Production-ready library

### Phase 3: Advanced Features (Week 5-6)
- [ ] Add conversation summarization (compress old turns)
- [ ] Implement automatic pruning policies
- [ ] Add multi-agent conversation support
- [ ] Build monitoring/dashboard
- [ ] Create SDKs (Python, Node.js, Go)

**Deliverable**: Full-featured product

### Phase 4: Scale & Deploy (Week 7-8)
- [ ] Load testing (1000+ concurrent conversations)
- [ ] Horizontal scaling strategy
- [ ] Deployment guides (Docker, Kubernetes, cloud)
- [ ] Documentation and examples
- [ ] Launch on PyPI, NPM, etc.

**Deliverable**: Public release

---

## 📈 Success Metrics

**Technical**:
- ✅ Retrieval latency < 500ms p95
- ✅ Recall@10 > 85% on test queries
- ✅ Supports 100+ concurrent conversations
- ✅ Memory usage < 100MB per agent instance

**Adoption**:
- ✅ 10+ beta users in first month
- ✅ 100+ GitHub stars in first month
- ✅ 3+ production deployments
- ✅ Integration with 2+ agent frameworks (LangChain, LlamaIndex)

**Business**:
- ✅ Clear monetization path (managed service, enterprise features)
- ✅ Differentiated from generic RAG solutions
- ✅ Strong technical moat (temporal weighting, conversation-aware)

---

## 🎯 Competitive Landscape

| Solution | Conversation-Aware | Temporal Weighting | Hybrid Search | Open Source |
|----------|-------------------|-------------------|---------------|-------------|
| **Our Solution** | ✅ Yes | ✅ Yes | ✅ Yes | ✅ Yes |
| LangChain Memory | ⚠️ Partial | ❌ No | ⚠️ Depends | ✅ Yes |
| LlamaIndex Chat Engine | ⚠️ Partial | ❌ No | ⚠️ Depends | ✅ Yes |
| Weaviate + Memory | ⚠️ Manual setup | ❌ No | ✅ Yes | ✅ Yes |
| Pinecone + Redis | ⚠️ Custom build | ⚠️ Custom | ⚠️ Custom | ❌ No |
| Zep | ✅ Yes | ⚠️ Limited | ⚠️ Limited | ✅ Yes |

**Our Advantage**: Purpose-built for agent conversations with temporal awareness, not bolted-on memory.

---

## 💰 Monetization Strategy

### Tier 1: Open Source (Free)
- Core library
- Self-hosted deployment
- Community support

### Tier 2: Cloud Managed ($50-200/month)
- Hosted service
- Auto-scaling
- Monitoring dashboard
- Priority support

### Tier 3: Enterprise ($500-2000/month)
- VPC deployment
- Custom integrations
- SLA guarantees
- Dedicated support
- Advanced analytics

### Tier 4: Usage-Based ($0.01 per 1K queries)
- Pay-per-retrieval
- High-volume customers
- No minimum commitment

---

## 🔒 Security Considerations

1. **Data Isolation**: Each agent's conversation encrypted and isolated
2. **Access Control**: RBAC for who can query/retrieve conversations
3. **Audit Logging**: All retrievals logged with timestamp, user, query
4. **Data Retention**: Configurable auto-deletion policies
5. **PII Detection**: Automatic redaction of sensitive information
6. **Encryption**: At-rest and in-transit encryption mandatory

---

## 📝 Next Steps

1. **Choose Tech Stack**: Weaviate fork vs. from scratch (recommend: Weaviate)
2. **Set Up Repo**: Create GitHub repository with MIT license
3. **Build MVP**: Focus on Phase 1 (core retrieval)
4. **Test with Real Data**: Use your own conversations as testbed
5. **Iterate**: Get feedback from 3-5 beta users
6. **Launch**: Public release with documentation

---

## 🧠 Key Insights from Research

1. **Hybrid search is table stakes** - pure vector search fails on exact terms
2. **Conversation context matters** - agents need to remember what was just discussed
3. **Temporal weighting is the secret sauce** - recent conversations should dominate
4. **Simplicity wins** - don't over-engineer with attention mechanisms initially
5. **Weaviate is the right foundation** - Go-based, hybrid search, extensible

---

**Status**: Ready for development  
**Priority**: HIGH (validates pain point, clear differentiation, monetizable)  
**Estimated MVP**: 2-4 weeks  
**Recommended Owner**: ALPHA (fits your expertise in Go, infra, AI systems)
