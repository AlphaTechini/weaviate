package schema

import (
	"github.com/weaviate/weaviate/entities/models"
	"github.com/weaviate/weaviate/entities/schema"
)

func boolPtr(b bool) *bool {
	return &b
}

// GetAgentRAGSchema returns the complete schema for agent-rag module
func GetAgentRAGSchema() *models.Schema {
	return &models.Schema{
		Classes: []*models.Class{
			getKnowledgeBaseClass(),
			getConversationClass(),
		},
	}
}

// getKnowledgeBaseClass defines the static knowledge base schema
func getKnowledgeBaseClass() *models.Class {
	return &models.Class{
		Class:               "KnowledgeBase",
		Description:         "Static knowledge base for agent-rag hybrid search",
		Vectorizer:          "text2vec-transformers",
		VectorIndexType:     "hnsw",
		ReplicationConfig:   &models.ReplicationConfig{Factor: 1},
		InvertedIndexConfig: &models.InvertedIndexConfig{IndexNullState: true},
		
		Properties: []*models.Property{
			{
				Name:         "title",
				DataType:     schema.DataTypeText.PropString(),
				Description:  "Document title",
				Tokenization: "word",
				IndexFilterable: boolPtr(true),
				IndexSearchable: boolPtr(true),
			},
			{
				Name:         "content",
				DataType:     schema.DataTypeText.PropString(),
				Description:  "Document content",
				Tokenization: "word",
				IndexFilterable: boolPtr(true),
				IndexSearchable: boolPtr(true),
			},
			{
				Name:         "updatedAt",
				DataType:     schema.DataTypeDate.PropString(),
				Description:  "Last update timestamp",
				IndexFilterable: boolPtr(true),
				IndexSearchable: boolPtr(false),
			},
			{
				Name:         "category",
				DataType:     schema.DataTypeText.PropString(),
				Description:  "Document category",
				Tokenization: "word",
				IndexFilterable: boolPtr(true),
				IndexSearchable: boolPtr(true),
			},
			{
				Name:         "metadata",
				DataType:     schema.DataTypeText.PropString(),
				Description:  "Additional metadata (JSON string)",
				Tokenization: "whitespace",
				IndexFilterable: boolPtr(false),
				IndexSearchable: boolPtr(false),
			},
		},
	}
}

// getConversationClass defines the conversation memory schema
func getConversationClass() *models.Class {
	return &models.Class{
		Class:               "Conversation",
		Description:         "Live conversation memory for agent-rag with temporal awareness",
		Vectorizer:          "text2vec-transformers",
		VectorIndexType:     "hnsw",
		ReplicationConfig:   &models.ReplicationConfig{Factor: 1},
		InvertedIndexConfig: &models.InvertedIndexConfig{IndexNullState: true},
		
		Properties: []*models.Property{
			{
				Name:         "message",
				DataType:     schema.DataTypeText.PropString(),
				Description:  "Conversation message text",
				Tokenization: "word",
				IndexFilterable: boolPtr(true),
				IndexSearchable: boolPtr(true),
			},
			{
				Name:         "speaker",
				DataType:     schema.DataTypeText.PropString(),
				Description:  "Message speaker (user/assistant)",
				Tokenization: "word",
				IndexFilterable: boolPtr(true),
				IndexSearchable: boolPtr(true),
			},
			{
				Name:         "timestamp",
				DataType:     schema.DataTypeDate.PropString(),
				Description:  "Message timestamp (UTC)",
				IndexFilterable: boolPtr(true),
				IndexSearchable: boolPtr(false),
			},
			{
				Name:         "turnIndex",
				DataType:     schema.DataTypeInt.PropString(),
				Description:  "Conversation turn index",
				IndexFilterable: boolPtr(true),
				IndexSearchable: boolPtr(false),
			},
			{
				Name:         "sessionID",
				DataType:     schema.DataTypeText.PropString(),
				Description:  "Session identifier for multi-session support",
				Tokenization: "whitespace",
				IndexFilterable: boolPtr(true),
				IndexSearchable: boolPtr(false),
			},
			{
				Name:         "metadata",
				DataType:     schema.DataTypeText.PropString(),
				Description:  "Additional metadata (JSON string)",
				Tokenization: "whitespace",
				IndexFilterable: boolPtr(false),
				IndexSearchable: boolPtr(false),
			},
		},
	}
}

// GetDefaultConfig returns default configuration for agent-rag classes
func GetDefaultConfig() map[string]interface{} {
	return map[string]interface{}{
		"KnowledgeBase": map[string]interface{}{
			"vectorizer": "text2vec-transformers",
			"moduleConfig": map[string]interface{}{
				"text2vec-transformers": map[string]interface{}{
					"model": "sentence-transformers/all-MiniLM-L6-v2",
					"poolingStrategy": "mean",
				},
			},
		},
		"Conversation": map[string]interface{}{
			"vectorizer": "text2vec-transformers",
			"moduleConfig": map[string]interface{}{
				"text2vec-transformers": map[string]interface{}{
					"model": "sentence-transformers/all-MiniLM-L6-v2",
					"poolingStrategy": "mean",
				},
			},
		},
	}
}
