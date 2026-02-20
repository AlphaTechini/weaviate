package retriever

import (
	"math"
	"time"
)

// TemporalDecay applies time-based decay to conversation scores
type TemporalDecay struct {
	halfLifeMinutes float64
	minWeight       float64
	enabled         bool
}

// NewTemporalDecay creates a new temporal decay calculator
func NewTemporalDecay(halfLifeMinutes, minWeight float64, enabled bool) *TemporalDecay {
	return &TemporalDecay{
		halfLifeMinutes: halfLifeMinutes,
		minWeight:       minWeight,
		enabled:         enabled,
	}
}

// Apply calculates the decayed score based on timestamp
// score: base score from vector similarity
// timestamp: when the conversation turn occurred
// currentTime: current time for calculating age
func (td *TemporalDecay) Apply(score float64, timestamp time.Time, currentTime time.Time) float64 {
	if !td.enabled {
		return score
	}
	
	// Calculate time difference in minutes
	timeDiff := currentTime.Sub(timestamp).Minutes()
	
	// Apply exponential decay: score * e^(-ln(2) * t / half_life)
	decayFactor := math.Exp(-math.Ln2 * timeDiff / td.halfLifeMinutes)
	
	// Ensure we don't go below minimum weight
	if decayFactor < td.minWeight {
		decayFactor = td.minWeight
	}
	
	return score * decayFactor
}

// ApplyToResults applies temporal decay to a list of search results
func (td *TemporalDecay) ApplyToResults(results SearchResults, currentTime time.Time) SearchResults {
	decayed := make(SearchResults, len(results))
	for i, result := range results {
		if result.Source == SourceConversation && result.Timestamp != nil {
			decayed[i] = result
			decayed[i].Score = td.Apply(result.Score, *result.Timestamp, currentTime)
		} else {
			decayed[i] = result
		}
	}
	return decayed
}

// HalfLife returns the configured half-life in minutes
func (td *TemporalDecay) HalfLife() float64 {
	return td.halfLifeMinutes
}

// MinWeight returns the minimum weight after decay
func (td *TemporalDecay) MinWeight() float64 {
	return td.minWeight
}

// Enabled returns whether temporal decay is enabled
func (td *TemporalDecay) Enabled() bool {
	return td.enabled
}

// CalculateHalfLifeFromRetention calculates half-life based on desired retention period
// For example, if you want conversations to retain 10% relevance after 2 hours:
// halfLife = CalculateHalfLifeFromRetention(120, 0.1) ≈ 36 minutes
func CalculateHalfLifeFromRetention(retentionMinutes, targetRetention float64) float64 {
	if targetRetention <= 0 || targetRetention >= 1 {
		return 30.0 // default
	}
	// Solve: target = e^(-ln(2) * retention / halfLife)
	// halfLife = -ln(2) * retention / ln(target)
	return -math.Ln2 * retentionMinutes / math.Log(targetRetention)
}

// Example usage and test helpers
func init() {
	// Pre-calculate common half-life scenarios
	_ = map[string]float64{
		"short_conversation":  15.0,  // 15 min half-life
		"standard_session":    30.0,  // 30 min half-life (default)
		"long_running_agent":  120.0, // 2 hour half-life
		"persistent_memory":   1440.0, // 24 hour half-life
	}
}
