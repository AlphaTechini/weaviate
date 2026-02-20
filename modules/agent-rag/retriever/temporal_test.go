package retriever

import (
	"math"
	"testing"
	"time"
)

func TestTemporalDecay_Apply(t *testing.T) {
	tests := []struct {
		name          string
		halfLife      float64
		minWeight     float64
		timeDiffMin   float64
		baseScore     float64
		expectedScore float64
		tolerance     float64
	}{
		{
			name:          "no decay at time zero",
			halfLife:      30.0,
			minWeight:     0.01,
			timeDiffMin:   0.0,
			baseScore:     1.0,
			expectedScore: 1.0,
			tolerance:     0.001,
		},
		{
			name:          "half score at half-life",
			halfLife:      30.0,
			minWeight:     0.01,
			timeDiffMin:   30.0,
			baseScore:     1.0,
			expectedScore: 0.5,
			tolerance:     0.01,
		},
		{
			name:          "quarter score at two half-lives",
			halfLife:      30.0,
			minWeight:     0.01,
			timeDiffMin:   60.0,
			baseScore:     1.0,
			expectedScore: 0.25,
			tolerance:     0.01,
		},
		{
			name:          "respects minimum weight",
			halfLife:      30.0,
			minWeight:     0.1,
			timeDiffMin:   300.0, // 10 half-lives, should be ~0.001
			baseScore:     1.0,
			expectedScore: 0.1, // clamped to min
			tolerance:     0.001,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			td := NewTemporalDecay(tt.halfLife, tt.minWeight, true)
			
			now := time.Now()
			past := now.Add(-time.Duration(tt.timeDiffMin) * time.Minute)
			
			result := td.Apply(tt.baseScore, past, now)
			
			if math.Abs(result-tt.expectedScore) > tt.tolerance {
				t.Errorf("Expected score %.4f, got %.4f (diff: %.4f)",
					tt.expectedScore, result, math.Abs(result-tt.expectedScore))
			}
		})
	}
}

func TestTemporalDecay_Disabled(t *testing.T) {
	td := NewTemporalDecay(30.0, 0.01, false)
	
	now := time.Now()
	past := now.Add(-60 * time.Minute)
	
	result := td.Apply(1.0, past, now)
	
	if result != 1.0 {
		t.Errorf("Expected no decay when disabled, got %.4f", result)
	}
}

func TestCalculateHalfLifeFromRetention(t *testing.T) {
	tests := []struct {
		name             string
		retentionMinutes float64
		targetRetention  float64
		expectedHalfLife float64
		tolerance        float64
	}{
		{
			name:             "10% after 2 hours",
			retentionMinutes: 120.0,
			targetRetention:  0.1,
			expectedHalfLife: 36.0,
			tolerance:        1.0,
		},
		{
			name:             "50% after 30 minutes",
			retentionMinutes: 30.0,
			targetRetention:  0.5,
			expectedHalfLife: 30.0,
			tolerance:        1.0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			halfLife := CalculateHalfLifeFromRetention(tt.retentionMinutes, tt.targetRetention)
			
			if math.Abs(halfLife-tt.expectedHalfLife) > tt.tolerance {
				t.Errorf("Expected half-life %.2f, got %.2f", tt.expectedHalfLife, halfLife)
			}
		})
	}
}

func TestTemporalDecay_ApplyToResults(t *testing.T) {
	td := NewTemporalDecay(30.0, 0.01, true)
	
	now := time.Now()
	
	results := SearchResults{
		{
			ID:        "static-1",
			Score:     0.9,
			Source:    SourceStatic,
			Timestamp: nil,
		},
		{
			ID:        "conv-recent",
			Score:     0.8,
			Source:    SourceConversation,
			Timestamp: &now,
		},
		{
			ID:        "conv-old",
			Score:     0.8,
			Source:    SourceConversation,
			Timestamp: func() *time.Time { t := now.Add(-60 * time.Minute); return &t }(),
		},
	}
	
	decayed := td.ApplyToResults(results, now)
	
	// Static should be unchanged
	if decayed[0].Score != 0.9 {
		t.Errorf("Static result should not decay, got %.4f", decayed[0].Score)
	}
	
	// Recent conversation should be unchanged (time diff = 0)
	if math.Abs(decayed[1].Score-0.8) > 0.001 {
		t.Errorf("Recent conversation should not decay, got %.4f", decayed[1].Score)
	}
	
	// Old conversation should be decayed (2 half-lives = 0.25)
	expectedOld := 0.8 * 0.25
	if math.Abs(decayed[2].Score-expectedOld) > 0.01 {
		t.Errorf("Old conversation should decay to %.4f, got %.4f", expectedOld, decayed[2].Score)
	}
}
