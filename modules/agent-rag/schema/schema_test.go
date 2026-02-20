package schema

import (
	"testing"
	
	"github.com/weaviate/weaviate/entities/models"
)

func TestGetAgentRAGSchema(t *testing.T) {
	schema := GetAgentRAGSchema()
	
	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	
	if len(schema.Classes) != 2 {
		t.Errorf("Expected 2 classes, got %d", len(schema.Classes))
	}
}

func TestKnowledgeBaseClass(t *testing.T) {
	schema := GetAgentRAGSchema()
	
	var kbClass *models.Class
	for _, class := range schema.Classes {
		if class.Class == "KnowledgeBase" {
			kbClass = class
			break
		}
	}
	
	if kbClass == nil {
		t.Fatal("KnowledgeBase class not found")
	}
	
	// Verify basic properties
	if kbClass.Class != "KnowledgeBase" {
		t.Errorf("Expected class name 'KnowledgeBase', got '%s'", kbClass.Class)
	}
	
	if kbClass.Vectorizer != "text2vec-transformers" {
		t.Errorf("Expected vectorizer 'text2vec-transformers', got '%s'", kbClass.Vectorizer)
	}
	
	// Verify properties exist
	expectedProps := []string{"title", "content", "updatedAt", "category", "metadata"}
	propNames := make(map[string]bool)
	for _, prop := range kbClass.Properties {
		propNames[prop.Name] = true
	}
	
	for _, expected := range expectedProps {
		if !propNames[expected] {
			t.Errorf("Missing expected property: %s", expected)
		}
	}
}

func TestConversationClass(t *testing.T) {
	schema := GetAgentRAGSchema()
	
	var convClass *models.Class
	for _, class := range schema.Classes {
		if class.Class == "Conversation" {
			convClass = class
			break
		}
	}
	
	if convClass == nil {
		t.Fatal("Conversation class not found")
	}
	
	// Verify basic properties
	if convClass.Class != "Conversation" {
		t.Errorf("Expected class name 'Conversation', got '%s'", convClass.Class)
	}
	
	if convClass.Vectorizer != "text2vec-transformers" {
		t.Errorf("Expected vectorizer 'text2vec-transformers', got '%s'", convClass.Vectorizer)
	}
	
	// Verify conversation-specific properties
	expectedProps := []string{"message", "speaker", "timestamp", "turnIndex", "sessionID", "metadata"}
	propNames := make(map[string]bool)
	for _, prop := range convClass.Properties {
		propNames[prop.Name] = true
	}
	
	for _, expected := range expectedProps {
		if !propNames[expected] {
			t.Errorf("Missing expected property: %s", expected)
		}
	}
	
	// Verify timestamp is filterable (needed for temporal queries)
	for _, prop := range convClass.Properties {
		if prop.Name == "timestamp" {
			if prop.IndexFilterable == nil || !*prop.IndexFilterable {
				t.Error("timestamp property should be filterable for temporal queries")
			}
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	config := GetDefaultConfig()
	
	if config == nil {
		t.Fatal("Default config should not be nil")
	}
	
	// Check KnowledgeBase config
	if kbConfig, ok := config["KnowledgeBase"].(map[string]interface{}); ok {
		if _, exists := kbConfig["vectorizer"]; !exists {
			t.Error("KnowledgeBase config missing vectorizer")
		}
	} else {
		t.Error("KnowledgeBase config not found or wrong type")
	}
	
	// Check Conversation config
	if convConfig, ok := config["Conversation"].(map[string]interface{}); ok {
		if _, exists := convConfig["vectorizer"]; !exists {
			t.Error("Conversation config missing vectorizer")
		}
	} else {
		t.Error("Conversation config not found or wrong type")
	}
}

func TestSchemaDescriptions(t *testing.T) {
	schema := GetAgentRAGSchema()
	
	for _, class := range schema.Classes {
		if class.Description == "" {
			t.Errorf("Class %s should have a description", class.Class)
		}
		
		for _, prop := range class.Properties {
			if prop.Description == "" {
				t.Errorf("Property %s.%s should have a description", class.Class, prop.Name)
			}
		}
	}
}