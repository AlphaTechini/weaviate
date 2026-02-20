package graphql

import (
	"strings"
	"testing"
)

func TestQueryBuilder_HybridQuery(t *testing.T) {
	qb := NewQueryBuilder("KnowledgeBase", 10)
	vector := []float32{0.1, 0.2, 0.3}
	
	query := qb.HybridQuery("machine learning", vector, 0.5)
	
	// Check structure
	if !strings.Contains(query, "Get { KnowledgeBase(") {
		t.Error("Query should contain class name")
	}
	
	if !strings.Contains(query, `hybrid:{query:"machine learning"`) {
		t.Error("Query should contain hybrid clause with query text")
	}
	
	if !strings.Contains(query, "vector:[0.100000,0.200000,0.300000]") {
		t.Error("Query should contain formatted vector")
	}
	
	if !strings.Contains(query, "alpha:0.50") {
		t.Error("Query should contain alpha parameter")
	}
	
	if !strings.Contains(query, "limit:10") {
		t.Error("Query should contain limit")
	}
	
	t.Logf("Generated query: %s", query)
}

func TestQueryBuilder_NearVectorQuery(t *testing.T) {
	qb := NewQueryBuilder("Conversation", 5)
	vector := []float32{0.5, 0.6, 0.7}
	
	query := qb.NearVectorQuery(vector, 0.8)
	
	if !strings.Contains(query, "nearVector:{vector:[0.500000,0.600000,0.700000]") {
		t.Error("Query should contain nearVector clause")
	}
	
	if !strings.Contains(query, "certainty:0.8000") {
		t.Error("Query should contain certainty parameter")
	}
	
	t.Logf("Generated query: %s", query)
}

func TestQueryBuilder_ConversationQuery(t *testing.T) {
	qb := NewQueryBuilder("Conversation", 10)
	vector := []float32{0.1, 0.2}
	sinceTime := "2026-02-20T10:00:00Z"
	
	query := qb.ConversationQuery(vector, sinceTime)
	
	// Check conversation-specific fields
	if !strings.Contains(query, "message,speaker,timestamp") {
		t.Error("Conversation query should include message, speaker, timestamp fields")
	}
	
	// Check time filter
	if !strings.Contains(query, `valueDate:"2026-02-20T10:00:00Z"`) {
		t.Error("Query should include time filter")
	}
	
	if !strings.Contains(query, "operator:GreaterThanEqual") {
		t.Error("Query should use GreaterThanEqual operator for time range")
	}
	
	t.Logf("Generated query: %s", query)
}

func TestQueryBuilder_WhereFilter(t *testing.T) {
	qb := NewQueryBuilder("Conversation", 10)
	
	filter := WhereFilter{
		Operator: "Equal",
		Path:     []string{"speaker"},
		Value:    "user",
	}
	
	whereClause := qb.BuildWhereClause(filter)
	
	if !strings.Contains(whereClause, "operator:Equal") {
		t.Error("Where clause should contain operator")
	}
	
	if !strings.Contains(whereClause, "path:[speaker]") {
		t.Error("Where clause should contain path")
	}
	
	if !strings.Contains(whereClause, "value:user") {
		t.Error("Where clause should contain value")
	}
	
	t.Logf("Generated where clause: %s", whereClause)
}

func TestQueryBuilder_ComplexWhereFilter(t *testing.T) {
	qb := NewQueryBuilder("KnowledgeBase", 10)
	
	// Complex filter with AND
	filter := WhereFilter{
		Operator: "And",
		Operands: []WhereFilter{
			{
				Operator: "Equal",
				Path:     []string{"category"},
				Value:    "documentation",
			},
			{
				Operator: "GreaterThan",
				Path:     []string{"updatedAt"},
				Value:    "2026-01-01T00:00:00Z",
			},
		},
	}
	
	whereClause := qb.BuildWhereClause(filter)
	
	if !strings.Contains(whereClause, "operator:And") {
		t.Error("Should contain AND operator")
	}
	
	if strings.Count(whereClause, "operator:Equal") != 1 {
		t.Error("Should contain one Equal operator")
	}
	
	t.Logf("Generated complex where clause: %s", whereClause)
}

func TestQueryBuilder_NearVectorWithWhere(t *testing.T) {
	qb := NewQueryBuilder("Conversation", 5)
	vector := []float32{0.1, 0.2, 0.3}
	
	filter := WhereFilter{
		Operator: "Equal",
		Path:     []string{"speaker"},
		Value:    "assistant",
	}
	
	query := qb.NearVectorWithWhere(vector, filter)
	
	if !strings.Contains(query, "nearVector:") {
		t.Error("Should contain nearVector clause")
	}
	
	if !strings.Contains(query, "where:{") {
		t.Error("Should contain where clause")
	}
	
	t.Logf("Generated query: %s", query)
}

func TestQueryBuilder_BatchDeleteQuery(t *testing.T) {
	qb := NewQueryBuilder("Conversation", 0)
	
	filter := WhereFilter{
		Operator: "LessThan",
		Path:     []string{"timestamp"},
		Value:    "2026-01-01T00:00:00Z",
	}
	
	query := qb.BatchDeleteQuery(filter)
	
	if !strings.Contains(query, "mutation{BatchDelete{objects(class:\"Conversation\"") {
		t.Error("Should be a mutation with class name")
	}
	
	if !strings.Contains(query, "where:{operator:LessThan") {
		t.Error("Should contain where filter")
	}
	
	t.Logf("Generated delete query: %s", query)
}

func TestFormatVector(t *testing.T) {
	tests := []struct {
		name   string
		vector []float32
		expect string
	}{
		{
			name:   "simple vector",
			vector: []float32{0.1, 0.2, 0.3},
			expect: "0.100000,0.200000,0.300000",
		},
		{
			name:   "empty vector",
			vector: []float32{},
			expect: "",
		},
		{
			name:   "single element",
			vector: []float32{0.5},
			expect: "0.500000",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatVector(tt.vector)
			if result != tt.expect {
				t.Errorf("Expected %q, got %q", tt.expect, result)
			}
		})
	}
}

func TestEscapeGraphQL(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "plain text",
			input:  "hello world",
			expect: "hello world",
		},
		{
			name:   "with quotes",
			input:  `say "hello"`,
			expect: `say \"hello\"`,
		},
		{
			name:   "with newline",
			input:  "line1\nline2",
			expect: "line1\\nline2",
		},
		{
			name:   "with backslash",
			input:  "path\\to\\file",
			expect: "path\\\\to\\\\file",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeGraphQL(tt.input)
			if result != tt.expect {
				t.Errorf("Expected %q, got %q", tt.expect, result)
			}
		})
	}
}

func TestQueryBuilder_FieldSelection(t *testing.T) {
	// Test Conversation class fields
	qbConv := NewQueryBuilder("Conversation", 10)
	queryConv := qbConv.HybridQuery("test", []float32{0.1}, 0.5)
	
	if !strings.Contains(queryConv, "message,speaker,timestamp") {
		t.Error("Conversation queries should include conversation-specific fields")
	}
	
	// Test KnowledgeBase class fields
	qbKB := NewQueryBuilder("KnowledgeBase", 10)
	queryKB := qbKB.HybridQuery("test", []float32{0.1}, 0.5)
	
	if !strings.Contains(queryKB, "title,content") {
		t.Error("KnowledgeBase queries should include KB-specific fields")
	}
}
