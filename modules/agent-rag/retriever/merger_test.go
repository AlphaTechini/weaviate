package retriever

import (
	"math"
	"testing"
	"time"
)

func TestResultMerger_WeightedMerge(t *testing.T) {
	config := &MergeConfig{
		StaticWeight:         0.6,
		ConversationWeight:   0.4,
		TemporalDecayEnabled: false, // Disable for this test
		HalfLifeMinutes:      30.0,  // Required even when disabled
		MinTemporalWeight:    0.01,
		Algorithm:            "weighted",
	}
	
	merger, err := NewResultMerger(config)
	if err != nil {
		t.Fatalf("Failed to create merger: %v", err)
	}
	
	now := time.Now()
	
	staticResults := SearchResults{
		{ID: "static-1", Score: 0.9, Source: SourceStatic},
		{ID: "static-2", Score: 0.7, Source: SourceStatic},
	}
	
	convResults := SearchResults{
		{ID: "conv-1", Score: 0.8, Source: SourceConversation, Timestamp: &now},
		{ID: "conv-2", Score: 0.6, Source: SourceConversation, Timestamp: &now},
	}
	
	merged := merger.Merge(staticResults, convResults, now)
	
	// Should have 4 results
	if len(merged) != 4 {
		t.Errorf("Expected 4 results, got %d", len(merged))
	}
	
	// Check scores are weighted correctly
	// static-1: 0.9 * 0.6 = 0.54
	// conv-1: 0.8 * 0.4 = 0.32
	expectedScores := map[string]float64{
		"static-1": 0.54,
		"static-2": 0.42,
		"conv-1":   0.32,
		"conv-2":   0.24,
	}
	
	for _, result := range merged {
		expected := expectedScores[result.ID]
		if math.Abs(result.Score-expected) > 0.001 {
			t.Errorf("Result %s: expected score %.4f, got %.4f", result.ID, expected, result.Score)
		}
	}
	
	// Results should be sorted by score descending
	for i := 1; i < len(merged); i++ {
		if merged[i].Score > merged[i-1].Score {
			t.Error("Results not sorted by score descending")
			break
		}
	}
}

func TestResultMerger_WithTemporalDecay(t *testing.T) {
	config := &MergeConfig{
		StaticWeight:         0.6,
		ConversationWeight:   0.4,
		TemporalDecayEnabled: true,
		HalfLifeMinutes:      30.0,
		MinTemporalWeight:    0.01,
		Algorithm:            "weighted",
	}
	
	merger, err := NewResultMerger(config)
	if err != nil {
		t.Fatalf("Failed to create merger: %v", err)
	}
	
	now := time.Now()
	oldTime := now.Add(-60 * time.Minute) // 2 half-lives ago
	
	staticResults := SearchResults{
		{ID: "static-1", Score: 0.5, Source: SourceStatic},
	}
	
	convResults := SearchResults{
		{ID: "conv-recent", Score: 0.8, Source: SourceConversation, Timestamp: &now},
		{ID: "conv-old", Score: 0.9, Source: SourceConversation, Timestamp: &oldTime},
	}
	
	merged := merger.Merge(staticResults, convResults, now)
	
	// Find the results
	var recentScore, oldScore float64
	for _, r := range merged {
		if r.ID == "conv-recent" {
			recentScore = r.Score
		} else if r.ID == "conv-old" {
			oldScore = r.Score
		}
	}
	
	// Recent: 0.8 * 0.4 * 1.0 (no decay) = 0.32
	expectedRecent := 0.32
	if math.Abs(recentScore-expectedRecent) > 0.01 {
		t.Errorf("Recent conversation: expected %.4f, got %.4f", expectedRecent, recentScore)
	}
	
	// Old: 0.9 * 0.4 * 0.25 (2 half-lives) = 0.09
	expectedOld := 0.09
	if math.Abs(oldScore-expectedOld) > 0.01 {
		t.Errorf("Old conversation: expected %.4f, got %.4f", expectedOld, oldScore)
	}
	
	// Recent should rank higher than old despite lower base score
	if recentScore <= oldScore {
		t.Error("Recent conversation should rank higher than old with temporal decay")
	}
}

func TestResultMerger_RRF(t *testing.T) {
	config := &MergeConfig{
		StaticWeight:         0.5,
		ConversationWeight:   0.5,
		TemporalDecayEnabled: false,
		HalfLifeMinutes:      30.0,
		MinTemporalWeight:    0.01,
		Algorithm:            "rrf",
		RRFK:                 60,
	}
	
	merger, err := NewResultMerger(config)
	if err != nil {
		t.Fatalf("Failed to create merger: %v", err)
	}
	
	now := time.Now()
	
	staticResults := SearchResults{
		{ID: "static-1", Score: 0.9, Source: SourceStatic},
		{ID: "static-2", Score: 0.8, Source: SourceStatic},
	}
	
	convResults := SearchResults{
		{ID: "conv-1", Score: 0.95, Source: SourceConversation, Timestamp: &now},
		{ID: "conv-2", Score: 0.85, Source: SourceConversation, Timestamp: &now},
	}
	
	merged := merger.Merge(staticResults, convResults, now)
	
	// With RRF and equal weights, ranking depends on position in both lists
	// static-1: rank 0 in static → 1/(60+0) * 0.5 = 0.00833
	// conv-1: rank 0 in conv → 1/(60+0) * 0.5 = 0.00833
	// If same ID appears in both, scores add up
	
	if len(merged) != 4 {
		t.Errorf("Expected 4 results, got %d", len(merged))
	}
}

func TestResultMerger_ConfigValidation(t *testing.T) {
	invalidConfigs := []struct {
		name   string
		config *MergeConfig
	}{
		{
			name: "negative static weight",
			config: &MergeConfig{
				StaticWeight: -0.1,
			},
		},
		{
			name: "weight > 1",
			config: &MergeConfig{
				StaticWeight: 1.5,
			},
		},
		{
			name: "negative half-life",
			config: &MergeConfig{
				StaticWeight:    0.5,
				HalfLifeMinutes: -30.0,
			},
		},
	}
	
	for _, tc := range invalidConfigs {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewResultMerger(tc.config)
			if err == nil {
				t.Error("Expected error for invalid config")
			}
		})
	}
}

func TestResultMerger_SetConfig(t *testing.T) {
	initialConfig := DefaultMergeConfig()
	merger, err := NewResultMerger(initialConfig)
	if err != nil {
		t.Fatalf("Failed to create merger: %v", err)
	}
	
	newConfig := &MergeConfig{
		StaticWeight:         0.7,
		ConversationWeight:   0.3,
		TemporalDecayEnabled: true,
		HalfLifeMinutes:      60.0,
		Algorithm:            "rrf",
	}
	
	err = merger.SetConfig(newConfig)
	if err != nil {
		t.Fatalf("Failed to set config: %v", err)
	}
	
	got := merger.GetConfig()
	if got.StaticWeight != newConfig.StaticWeight {
		t.Errorf("StaticWeight: expected %.2f, got %.2f", newConfig.StaticWeight, got.StaticWeight)
	}
}
