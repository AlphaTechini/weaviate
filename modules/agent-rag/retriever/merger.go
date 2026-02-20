package retriever

import (
	"sort"
	"time"
)

// ResultMerger handles merging of search results from multiple sources
type ResultMerger struct {
	config         *MergeConfig
	temporalDecay  *TemporalDecay
}

// NewResultMerger creates a new result merger with the given configuration
func NewResultMerger(config *MergeConfig) (*ResultMerger, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	
	merger := &ResultMerger{
		config: config,
		temporalDecay: NewTemporalDecay(
			config.HalfLifeMinutes,
			config.MinTemporalWeight,
			config.TemporalDecayEnabled,
		),
	}
	
	return merger, nil
}

// Merge combines static and conversation results using the configured algorithm
func (rm *ResultMerger) Merge(staticResults, convResults SearchResults, currentTime time.Time) SearchResults {
	switch rm.config.Algorithm {
	case "rrf", "reciprocal_rank_fusion":
		return rm.mergeRRF(staticResults, convResults, currentTime)
	default:
		return rm.mergeWeighted(staticResults, convResults, currentTime)
	}
}

// mergeWeighted applies simple weighted sum fusion
func (rm *ResultMerger) mergeWeighted(staticResults, convResults SearchResults, currentTime time.Time) SearchResults {
	scoreMap := make(map[string]float64)
	resultMap := make(map[string]SearchResult)
	
	// Apply weights to static results
	for _, result := range staticResults {
		weightedScore := result.Score * rm.config.StaticWeight
		scoreMap[result.ID] = weightedScore
		resultMap[result.ID] = result
	}
	
	// Apply weights and temporal decay to conversation results
	for _, result := range convResults {
		var weightedScore float64
		
		if result.Timestamp != nil {
			// Apply both weight and temporal decay
			baseScore := result.Score * rm.config.ConversationWeight
			weightedScore = rm.temporalDecay.Apply(baseScore, *result.Timestamp, currentTime)
		} else {
			weightedScore = result.Score * rm.config.ConversationWeight
		}
		
		// Add to existing score if present (cross-source match)
		if existingScore, exists := scoreMap[result.ID]; exists {
			scoreMap[result.ID] = existingScore + weightedScore
		} else {
			scoreMap[result.ID] = weightedScore
			resultMap[result.ID] = result
		}
	}
	
	// Convert map back to sorted slice
	results := make(SearchResults, 0, len(scoreMap))
	for id, score := range scoreMap {
		result := resultMap[id]
		result.Score = score
		results = append(results, result)
	}
	
	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	
	return results
}

// mergeRRF applies Reciprocal Rank Fusion
func (rm *ResultMerger) mergeRRF(staticResults, convResults SearchResults, currentTime time.Time) SearchResults {
	scoreMap := make(map[string]float64)
	resultMap := make(map[string]SearchResult)
	k := float64(rm.config.RRFK)
	
	// Score static results by rank
	for rank, result := range staticResults {
		rankScore := 1.0 / (k + float64(rank))
		weightedScore := rankScore * rm.config.StaticWeight
		scoreMap[result.ID] = weightedScore
		resultMap[result.ID] = result
	}
	
	// Score conversation results by rank with temporal decay
	// First apply temporal decay to sort order
	decayedConv := rm.temporalDecay.ApplyToResults(convResults, currentTime)
	sort.Slice(decayedConv, func(i, j int) bool {
		return decayedConv[i].Score > decayedConv[j].Score
	})
	
	for rank, result := range decayedConv {
		rankScore := 1.0 / (k + float64(rank))
		weightedScore := rankScore * rm.config.ConversationWeight
		
		if existingScore, exists := scoreMap[result.ID]; exists {
			scoreMap[result.ID] = existingScore + weightedScore
		} else {
			scoreMap[result.ID] = weightedScore
			resultMap[result.ID] = result
		}
	}
	
	// Convert map back to sorted slice
	results := make(SearchResults, 0, len(scoreMap))
	for id, score := range scoreMap {
		result := resultMap[id]
		result.Score = score
		results = append(results, result)
	}
	
	// Sort by final RRF score
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	
	return results
}

// SetConfig updates the merger configuration
func (rm *ResultMerger) SetConfig(config *MergeConfig) error {
	if err := config.Validate(); err != nil {
		return err
	}
	rm.config = config
	rm.temporalDecay = NewTemporalDecay(
		config.HalfLifeMinutes,
		config.MinTemporalWeight,
		config.TemporalDecayEnabled,
	)
	return nil
}

// GetConfig returns the current configuration
func (rm *ResultMerger) GetConfig() *MergeConfig {
	return rm.config
}
