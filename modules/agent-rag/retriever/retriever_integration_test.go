package retriever

import (
	"context"
	"testing"
	"time"
)

// TestAgentRAGRetriever_Creation tests basic retriever initialization
func TestAgentRAGRetriever_Creation(t *testing.T) {
	config := DefaultMergeConfig()
	indexConfig := DefaultIndexConfig()

	retriever, err := NewAgentRAGRetriever("http://localhost:8080", "", config, indexConfig)
	if err != nil {
		t.Fatalf("Failed to create retriever: %v", err)
	}
	defer retriever.Close()

	// Verify configuration
	gotConfig := retriever.GetConfig()
	if gotConfig.StaticWeight != config.StaticWeight {
		t.Errorf("StaticWeight mismatch: expected %.2f, got %.2f", config.StaticWeight, gotConfig.StaticWeight)
	}

	// Verify stats
	stats := retriever.GetStats()
	if stats["closed"] != false {
		t.Error("Retriever should not be closed initially")
	}
}

// TestAgentRAGRetriever_ConfigUpdate tests runtime configuration changes
func TestAgentRAGRetriever_ConfigUpdate(t *testing.T) {
	retriever, err := NewAgentRAGRetriever("http://localhost:8080", "", nil, nil)
	if err != nil {
		t.Fatalf("Failed to create retriever: %v", err)
	}
	defer retriever.Close()

	// Update configuration
	newConfig := &MergeConfig{
		StaticWeight:         0.7,
		ConversationWeight:   0.3,
		TemporalDecayEnabled: true,
		HalfLifeMinutes:      60.0,
		Algorithm:            "rrf",
	}

	err = retriever.UpdateConfig(newConfig)
	if err != nil {
		t.Fatalf("Failed to update config: %v", err)
	}

	// Verify update
	gotConfig := retriever.GetConfig()
	if gotConfig.StaticWeight != 0.7 {
		t.Errorf("StaticWeight not updated: expected 0.7, got %.2f", gotConfig.StaticWeight)
	}
	if gotConfig.Algorithm != "rrf" {
		t.Errorf("Algorithm not updated: expected 'rrf', got '%s'", gotConfig.Algorithm)
	}
}

// TestAgentRAGRetriever_Close tests proper resource cleanup
func TestAgentRAGRetriever_Close(t *testing.T) {
	retriever, err := NewAgentRAGRetriever("http://localhost:8080", "", nil, nil)
	if err != nil {
		t.Fatalf("Failed to create retriever: %v", err)
	}

	// Close once
	err = retriever.Close()
	if err != nil {
		t.Errorf("First close failed: %v", err)
	}

	if !retriever.IsClosed() {
		t.Error("Retriever should be closed after Close()")
	}

	// Close again (should be idempotent)
	err = retriever.Close()
	if err != nil {
		t.Errorf("Second close failed: %v", err)
	}
}

// TestAgentRAGRetriever_OperationAfterClose tests that operations fail gracefully after close
func TestAgentRAGRetriever_OperationAfterClose(t *testing.T) {
	retriever, err := NewAgentRAGRetriever("http://localhost:8080", "", nil, nil)
	if err != nil {
		t.Fatalf("Failed to create retriever: %v", err)
	}

	retriever.Close()

	ctx := context.Background()
	query := &Query{
		Text:  "test query",
		Limit: 10,
	}

	// Try to search after close
	_, err = retriever.SearchHybrid(ctx, query)
	if err != ErrClosedRetriever {
		t.Errorf("Expected ErrClosedRetriever, got: %v", err)
	}

	// Try to add conversation after close
	_, err = retriever.AddConversationTurn(ctx, "test", "user", nil)
	if err != ErrClosedRetriever {
		t.Errorf("Expected ErrClosedRetriever on AddConversationTurn, got: %v", err)
	}
}

// TestAgentRAGRetriever_ConcurrentAccess tests thread safety
func TestAgentRAGRetriever_ConcurrentAccess(t *testing.T) {
	retriever, err := NewAgentRAGRetriever("http://localhost:8080", "", nil, nil)
	if err != nil {
		t.Fatalf("Failed to create retriever: %v", err)
	}
	defer retriever.Close()

	ctx := context.Background()
	query := &Query{
		Text:  "test",
		Limit: 5,
	}

	// Run concurrent searches
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := retriever.SearchHybrid(ctx, query)
			// We expect errors since we don't have a real Weaviate instance
			// Just checking it doesn't panic
			done <- err == nil || err != ErrClosedRetriever
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// If we got here without panicking, test passes
}

// TestAgentRAGRetriever_HealthCheck tests health check functionality
func TestAgentRAGRetriever_HealthCheck(t *testing.T) {
	retriever, err := NewAgentRAGRetriever("http://localhost:8080", "", nil, nil)
	if err != nil {
		t.Fatalf("Failed to create retriever: %v", err)
	}
	defer retriever.Close()

	ctx := context.Background()
	err = retriever.HealthCheck(ctx)
	
	// We expect this to fail since there's no real Weaviate instance
	// But it should fail gracefully, not panic
	if err == nil {
		t.Log("Health check passed (Weaviate instance running?)")
	}
}

// TestAgentRAGRetriever_QueryVariations tests different query configurations
func TestAgentRAGRetriever_QueryVariations(t *testing.T) {
	retriever, err := NewAgentRAGRetriever("http://localhost:8080", "", nil, nil)
	if err != nil {
		t.Fatalf("Failed to create retriever: %v", err)
	}
	defer retriever.Close()

	tests := []struct {
		name  string
		query *Query
	}{
		{
			name: "basic text query",
			query: &Query{
				Text:  "what is machine learning",
				Limit: 10,
			},
		},
		{
			name: "query with vector",
			query: &Query{
				Text:   "machine learning",
				Vector: []float32{0.1, 0.2, 0.3},
				Limit:  5,
			},
		},
		{
			name: "query with time range",
			query: &Query{
				Text:  "recent conversations",
				Limit: 10,
				TimeRange: &TimeRange{
					Since: time.Now().Add(-24 * time.Hour),
					Until: time.Now(),
				},
			},
		},
		{
			name: "query with metadata filters",
			query: &Query{
				Text:  "filtered search",
				Limit: 5,
				Filters: map[string]interface{}{
					"speaker": "assistant",
				},
			},
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := retriever.SearchHybrid(ctx, tt.query)
			// Expecting errors due to no real Weaviate, but shouldn't panic
			if err == nil {
				t.Logf("Query succeeded (unexpected)")
			}
		})
	}
}

// TestDefaultConfigs tests that default configurations are sensible
func TestDefaultConfigs(t *testing.T) {
	mergeConfig := DefaultMergeConfig()
	
	if mergeConfig.StaticWeight != 0.6 {
		t.Errorf("Default StaticWeight should be 0.6, got %.2f", mergeConfig.StaticWeight)
	}
	if mergeConfig.ConversationWeight != 0.4 {
		t.Errorf("Default ConversationWeight should be 0.4, got %.2f", mergeConfig.ConversationWeight)
	}
	if mergeConfig.HalfLifeMinutes != 30.0 {
		t.Errorf("Default HalfLifeMinutes should be 30.0, got %.2f", mergeConfig.HalfLifeMinutes)
	}
	if mergeConfig.Algorithm != "weighted" {
		t.Errorf("Default Algorithm should be 'weighted', got '%s'", mergeConfig.Algorithm)
	}

	indexConfig := DefaultIndexConfig()
	if indexConfig.StaticIndexName != "KnowledgeBase" {
		t.Errorf("Default StaticIndexName should be 'KnowledgeBase', got '%s'", indexConfig.StaticIndexName)
	}
	if indexConfig.ConversationIndexName != "Conversation" {
		t.Errorf("Default ConversationIndexName should be 'Conversation', got '%s'", indexConfig.ConversationIndexName)
	}
}
