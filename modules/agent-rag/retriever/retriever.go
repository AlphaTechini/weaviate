package retriever

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AgentRAGRetriever is the main retriever implementation for agent-rag
type AgentRAGRetriever struct {
	client       *WeaviateClient
	merger       *ResultMerger
	config       *MergeConfig
	indexConfig  *IndexConfig
	temporalDecay *TemporalDecay
	mu           sync.RWMutex
	closed       bool
}

// NewAgentRAGRetriever creates a new agent-rag retriever
func NewAgentRAGRetriever(weaviateHost, apiKey string, mergeConfig *MergeConfig, indexConfig *IndexConfig) (*AgentRAGRetriever, error) {
	if mergeConfig == nil {
		mergeConfig = DefaultMergeConfig()
	}
	if indexConfig == nil {
		indexConfig = DefaultIndexConfig()
	}

	// Validate configuration
	if err := mergeConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Create Weaviate client
	client, err := NewWeaviateClient(weaviateHost, apiKey, indexConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	// Create result merger
	merger, err := NewResultMerger(mergeConfig)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create merger: %w", err)
	}

	// Create temporal decay calculator
	temporalDecay := NewTemporalDecay(
		mergeConfig.HalfLifeMinutes,
		mergeConfig.MinTemporalWeight,
		mergeConfig.TemporalDecayEnabled,
	)

	return &AgentRAGRetriever{
		client:        client,
		merger:        merger,
		config:        mergeConfig,
		indexConfig:   indexConfig,
		temporalDecay: temporalDecay,
		closed:        false,
	}, nil
}

// SearchStatic searches only the static knowledge base
func (r *AgentRAGRetriever) SearchStatic(ctx context.Context, query *Query) (SearchResults, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return nil, ErrClosedRetriever
	}

	results, err := r.client.SearchStatic(ctx, query)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// SearchConversation searches only conversation memory
func (r *AgentRAGRetriever) SearchConversation(ctx context.Context, query *Query) (SearchResults, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return nil, ErrClosedRetriever
	}

	results, err := r.client.SearchConversation(ctx, query)
	if err != nil {
		return nil, err
	}

	// Apply temporal decay
	now := time.Now()
	decayedResults := r.temporalDecay.ApplyToResults(results, now)

	return decayedResults, nil
}

// SearchHybrid performs hybrid search across both indices with intelligent merging
func (r *AgentRAGRetriever) SearchHybrid(ctx context.Context, query *Query) (SearchResults, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return nil, ErrClosedRetriever
	}

	// Execute searches in parallel for better performance
	var staticResults, convResults SearchResults
	var staticErr, convErr error

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		staticResults, staticErr = r.client.SearchStatic(ctx, query)
	}()

	go func() {
		defer wg.Done()
		convResults, convErr = r.client.SearchConversation(ctx, query)
	}()

	wg.Wait()

	// Handle errors
	if staticErr != nil && convErr != nil {
		return nil, fmt.Errorf("both searches failed: static=%v, conversation=%v", staticErr, convErr)
	}
	
	// If one failed, use empty results for that source
	if staticErr != nil {
		staticResults = SearchResults{}
	}
	if convErr != nil {
		convResults = SearchResults{}
	}

	// Merge results using configured algorithm
	now := time.Now()
	mergedResults := r.merger.Merge(staticResults, convResults, now)

	// Apply limit from query
	if query.Limit > 0 && len(mergedResults) > query.Limit {
		mergedResults = mergedResults[:query.Limit]
	}

	return mergedResults, nil
}

// AddConversationTurn adds a new conversation turn to memory
func (r *AgentRAGRetriever) AddConversationTurn(ctx context.Context, message, speaker string, metadata map[string]interface{}) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return "", ErrClosedRetriever
	}

	return r.client.AddConversationTurn(ctx, message, speaker, metadata)
}

// AddKnowledgeDocument adds a document to the static knowledge base
func (r *AgentRAGRetriever) AddKnowledgeDocument(ctx context.Context, title, content string, metadata map[string]interface{}) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return "", ErrClosedRetriever
	}

	return r.client.AddKnowledgeDocument(ctx, title, content, metadata)
}

// PruneOldConversations removes conversations older than maxAge
func (r *AgentRAGRetriever) PruneOldConversations(ctx context.Context, maxAge time.Duration) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return 0, ErrClosedRetriever
	}

	return r.client.PruneOldConversations(ctx, maxAge)
}

// UpdateConfig updates the retriever configuration at runtime
func (r *AgentRAGRetriever) UpdateConfig(config *MergeConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := config.Validate(); err != nil {
		return err
	}

	r.config = config
	r.merger.SetConfig(config)
	r.temporalDecay = NewTemporalDecay(
		config.HalfLifeMinutes,
		config.MinTemporalWeight,
		config.TemporalDecayEnabled,
	)

	return nil
}

// GetConfig returns the current configuration
func (r *AgentRAGRetriever) GetConfig() *MergeConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.config
}

// GetIndexConfig returns the index configuration
func (r *AgentRAGRetriever) GetIndexConfig() *IndexConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.indexConfig
}

// HealthCheck verifies the retriever and Weaviate connection are healthy
func (r *AgentRAGRetriever) HealthCheck(ctx context.Context) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return ErrClosedRetriever
	}

	return r.client.HealthCheck(ctx)
}

// Close releases all resources
func (r *AgentRAGRetriever) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}

	r.closed = true
	return r.client.Close()
}

// IsClosed returns whether the retriever has been closed
func (r *AgentRAGRetriever) IsClosed() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.closed
}

// GetStats returns runtime statistics
func (r *AgentRAGRetriever) GetStats() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return map[string]interface{}{
		"closed":              r.closed,
		"algorithm":           r.config.Algorithm,
		"staticWeight":        r.config.StaticWeight,
		"conversationWeight":  r.config.ConversationWeight,
		"temporalDecayEnabled": r.config.TemporalDecayEnabled,
		"halfLifeMinutes":     r.config.HalfLifeMinutes,
		"staticIndex":         r.indexConfig.StaticIndexName,
		"conversationIndex":   r.indexConfig.ConversationIndexName,
	}
}
