package retriever

import (
	"time"
)

// SearchResult represents a single retrieved chunk
type SearchResult struct {
	ID        string                 `json:"id"`
	DocID     string                 `json:"docId"`
	Score     float64                `json:"score"`
	Text      string                 `json:"text"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Source    SourceType             `json:"source"`
	Timestamp *time.Time             `json:"timestamp,omitempty"`
}

// SourceType indicates where the result came from
type SourceType string

const (
	SourceStatic       SourceType = "static"
	SourceConversation SourceType = "conversation"
)

// SearchResults is a collection of search results
type SearchResults []SearchResult

// MergeConfig holds configuration for merging results
type MergeConfig struct {
	// StaticWeight is the weight for static knowledge base results (0-1)
	StaticWeight float64 `json:"staticWeight"`
	
	// ConversationWeight is the weight for conversation results (0-1)
	ConversationWeight float64 `json:"conversationWeight"`
	
	// TemporalDecayEnabled enables time-based decay for conversations
	TemporalDecayEnabled bool `json:"temporalDecayEnabled"`
	
	// HalfLifeMinutes is the half-life for exponential decay
	HalfLifeMinutes float64 `json:"halfLifeMinutes"`
	
	// MinTemporalWeight is the minimum weight after decay
	MinTemporalWeight float64 `json:"minTemporalWeight"`
	
	// Algorithm is the merge algorithm: "weighted" or "rrf"
	Algorithm string `json:"algorithm"`
	
	// RRFK is the constant k for Reciprocal Rank Fusion
	RRFK int `json:"rrfK"`
}

// DefaultMergeConfig returns sensible defaults
func DefaultMergeConfig() *MergeConfig {
	return &MergeConfig{
		StaticWeight:         0.6,
		ConversationWeight:   0.4,
		TemporalDecayEnabled: true,
		HalfLifeMinutes:      30.0,
		MinTemporalWeight:    0.01,
		Algorithm:            "weighted",
		RRFK:                 60,
	}
}

// Validate ensures config values are reasonable
func (c *MergeConfig) Validate() error {
	if c.StaticWeight < 0 || c.StaticWeight > 1 {
		return ErrInvalidWeight
	}
	if c.ConversationWeight < 0 || c.ConversationWeight > 1 {
		return ErrInvalidWeight
	}
	if c.HalfLifeMinutes <= 0 {
		return ErrInvalidHalfLife
	}
	if c.MinTemporalWeight < 0 || c.MinTemporalWeight > 1 {
		return ErrInvalidMinWeight
	}
	return nil
}

// Query represents a search query with context
type Query struct {
	Text         string                 `json:"text"`
	Vector       []float32              `json:"vector,omitempty"`
	Filters      map[string]interface{} `json:"filters,omitempty"`
	Limit        int                    `json:"limit"`
	TimeRange    *TimeRange             `json:"timeRange,omitempty"`
	IncludeMeta  bool                   `json:"includeMeta"`
}

// TimeRange specifies a time window for filtering
type TimeRange struct {
	Since time.Time `json:"since"`
	Until time.Time `json:"until"`
}

// Retriever defines the interface for retrieval operations
type Retriever interface {
	// SearchStatic searches the static knowledge base
	SearchStatic(query *Query) (SearchResults, error)
	
	// SearchConversation searches conversation memory
	SearchConversation(query *Query) (SearchResults, error)
	
	// SearchHybrid performs hybrid search across both indices
	SearchHybrid(query *Query, config *MergeConfig) (SearchResults, error)
	
	// Close releases resources
	Close() error
}

// IndexConfig holds configuration for indices
type IndexConfig struct {
	StaticIndexName       string `json:"staticIndexName"`
	ConversationIndexName string `json:"conversationIndexName"`
	Vectorizer            string `json:"vectorizer"`
	DistanceMetric        string `json:"distanceMetric"`
}

// DefaultIndexConfig returns default index configuration
func DefaultIndexConfig() *IndexConfig {
	return &IndexConfig{
		StaticIndexName:       "KnowledgeBase",
		ConversationIndexName: "Conversation",
		Vectorizer:            "text2vec-transformers",
		DistanceMetric:        "cosine",
	}
}
