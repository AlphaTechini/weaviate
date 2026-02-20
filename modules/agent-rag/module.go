//                           _       _
// __      _____  __ ___   ___  __ _| |_ ___
// \ \ /\ / / _ \/ _` \ \ / / |/ _` | __/ _ \
//  \ V  V /  __/ (_| |\ V /| | (_| | ||  __/
//   \_/\_/ \___|\__,_| \_/ |_|\__,_|\__\___|
//
//  Copyright © 2016 - 2026 Weaviate B.V. All rights reserved.
//
//  CONTACT: hello@weaviate.io
//

package modagentrag

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/weaviate/weaviate/entities/moduletools"
)

const Name = "agent-rag"

func New() *AgentRAGModule {
	return &AgentRAGModule{}
}

type AgentRAGModule struct {
	logger logrus.FieldLogger
}

// Name returns the name of the module
func (m *AgentRAGModule) Name() string {
	return Name
}

// Init initializes the module
func (m *AgentRAGModule) Init(ctx context.Context, params moduletools.ModuleInitParams) error {
	m.logger = params.GetLogger()
	
	m.logger.Info("Agent-RAG module initialized")
	m.logger.Info("Features:")
	m.logger.Info("  - Hybrid search with conversation memory")
	m.logger.Info("  - Temporal decay for recent context")
	m.logger.Info("  - Weighted result merging")
	
	return nil
}

// MetaInfo returns metadata about the module
func (m *AgentRAGModule) MetaInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":        Name,
		"version":     "1.0.0",
		"description": "Agent-specific RAG with conversation memory and temporal weighting",
		"features": []string{
			"conversation-memory",
			"temporal-decay",
			"weighted-merge",
			"hybrid-search",
		},
	}
}

// Placeholder for module registration - will be implemented in Phase 3
