package retriever

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/weaviate/weaviate/modules/agent-rag/graphql"
)

// WeaviateClient wraps the Weaviate GraphQL/REST client for agent-rag operations
type WeaviateClient struct {
	host         string
	apiKey       string
	httpClient   *http.Client
	config       *IndexConfig
	queryBuilder *graphql.QueryBuilder
}

// GraphQLResponse represents a Weaviate GraphQL response
type GraphQLResponse struct {
	Data   map[string]interface{} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

// NewWeaviateClient creates a new Weaviate client
func NewWeaviateClient(host string, apiKey string, config *IndexConfig) (*WeaviateClient, error) {
	if config == nil {
		config = DefaultIndexConfig()
	}

	client := &WeaviateClient{
		host:   host,
		apiKey: apiKey,
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	return client, nil
}

// SearchStatic performs hybrid search on the static knowledge base
func (wc *WeaviateClient) SearchStatic(ctx context.Context, query *Query) (SearchResults, error) {
	qb := graphql.NewQueryBuilder(wc.config.StaticIndexName, query.Limit)
	
	// Build hybrid query
	graphQLQuery := qb.HybridQuery(query.Text, query.Vector, 0.5)
	
	// Execute query
	response, err := wc.executeGraphQL(ctx, graphQLQuery)
	if err != nil {
		return nil, fmt.Errorf("static search failed: %w", err)
	}
	
	// Parse results
	return wc.parseSearchResults(response, graphql.GetResultPath(wc.config.StaticIndexName))
}

// SearchConversation performs vector search on conversation memory with temporal filtering
func (wc *WeaviateClient) SearchConversation(ctx context.Context, query *Query) (SearchResults, error) {
	qb := graphql.NewQueryBuilder(wc.config.ConversationIndexName, query.Limit)
	
	// Build conversation-optimized query with time filter if provided
	var sinceTime string
	if query.TimeRange != nil {
		sinceTime = query.TimeRange.Since.Format(time.RFC3339)
	}
	
	graphQLQuery := qb.ConversationQuery(query.Vector, sinceTime)
	
	// Execute query
	response, err := wc.executeGraphQL(ctx, graphQLQuery)
	if err != nil {
		return nil, fmt.Errorf("conversation search failed: %w", err)
	}
	
	// Parse results with timestamps
	return wc.parseConversationResults(response, graphql.GetResultPath(wc.config.ConversationIndexName))
}

// AddConversationTurn adds a new conversation turn to the dynamic index
func (wc *WeaviateClient) AddConversationTurn(ctx context.Context, message, speaker string, metadata map[string]interface{}) (string, error) {
	object := map[string]interface{}{
		"class": wc.config.ConversationIndexName,
		"properties": map[string]interface{}{
			"message":   message,
			"speaker":   speaker,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		},
	}
	
	// Add metadata if provided
	for k, v := range metadata {
		object["properties"].(map[string]interface{})[k] = v
	}
	
	// Create object via REST API
	id, err := wc.createObject(ctx, object)
	if err != nil {
		return "", fmt.Errorf("failed to add conversation turn: %w", err)
	}
	
	return id, nil
}

// AddKnowledgeDocument adds a document to the static knowledge base
func (wc *WeaviateClient) AddKnowledgeDocument(ctx context.Context, title, content string, metadata map[string]interface{}) (string, error) {
	object := map[string]interface{}{
		"class": wc.config.StaticIndexName,
		"properties": map[string]interface{}{
			"title":     title,
			"content":   content,
			"updatedAt": time.Now().UTC().Format(time.RFC3339),
		},
	}
	
	// Add metadata if provided
	for k, v := range metadata {
		object["properties"].(map[string]interface{})[k] = v
	}
	
	id, err := wc.createObject(ctx, object)
	if err != nil {
		return "", fmt.Errorf("failed to add document: %w", err)
	}
	
	return id, nil
}

// PruneOldConversations removes conversations older than the specified age
func (wc *WeaviateClient) PruneOldConversations(ctx context.Context, maxAge time.Duration) (int, error) {
	cutoffTime := time.Now().UTC().Add(-maxAge)
	
	qb := graphql.NewQueryBuilder(wc.config.ConversationIndexName, 0)
	
	filter := graphql.WhereFilter{
		Operator: "LessThan",
		Path:     []string{"timestamp"},
		Value:    cutoffTime.Format(time.RFC3339),
	}
	
	deleteQuery := qb.BatchDeleteQuery(filter)
	
	count, err := wc.executeBatchDelete(ctx, deleteQuery)
	if err != nil {
		return 0, fmt.Errorf("pruning failed: %w", err)
	}
	
	return count, nil
}

// GetMetaInfo returns metadata about the indices
func (wc *WeaviateClient) GetMetaInfo(ctx context.Context) (map[string]interface{}, error) {
	metaQuery := `{Meta{hostname,version}}`
	
	response, err := wc.executeGraphQL(ctx, metaQuery)
	if err != nil {
		return nil, err
	}
	
	return response, nil
}

// HealthCheck verifies connection to Weaviate
func (wc *WeaviateClient) HealthCheck(ctx context.Context) error {
	_, err := wc.executeGraphQL(ctx, "{Meta{hostname}}")
	return err
}

// Close releases client resources
func (wc *WeaviateClient) Close() error {
	wc.httpClient.CloseIdleConnections()
	return nil
}

// Private helper methods

func (wc *WeaviateClient) executeGraphQL(ctx context.Context, query string) (map[string]interface{}, error) {
	requestBody := map[string]string{
		"query": query,
	}
	
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}
	
	url := fmt.Sprintf("%s/v1/graphql", wc.host)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	if wc.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+wc.apiKey)
	}
	
	resp, err := wc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	var gqlResp GraphQLResponse
	if err := json.Unmarshal(body, &gqlResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	if len(gqlResp.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL error: %s", gqlResp.Errors[0].Message)
	}
	
	return gqlResp.Data, nil
}

func (wc *WeaviateClient) parseSearchResults(data map[string]interface{}, path []string) (SearchResults, error) {
	// Navigate to results using path
	var current interface{} = data
	for _, key := range path {
		if m, ok := current.(map[string]interface{}); ok {
			if val, exists := m[key]; exists {
				current = val
			} else {
				return nil, fmt.Errorf("key not found: %s", key)
			}
		} else {
			return nil, fmt.Errorf("unexpected type at path %s", key)
		}
	}
	
	results := make(SearchResults, 0)
	
	// Extract objects from the result
	if objects, ok := current.([]interface{}); ok {
		for _, obj := range objects {
			if objMap, ok := obj.(map[string]interface{}); ok {
				result := wc.extractSearchResult(objMap)
				result.Source = SourceStatic
				results = append(results, result)
			}
		}
	}
	
	return results, nil
}

func (wc *WeaviateClient) parseConversationResults(data map[string]interface{}, path []string) (SearchResults, error) {
	// Navigate to results using path
	var current interface{} = data
	for _, key := range path {
		if m, ok := current.(map[string]interface{}); ok {
			if val, exists := m[key]; exists {
				current = val
			} else {
				return nil, fmt.Errorf("key not found: %s", key)
			}
		} else {
			return nil, fmt.Errorf("unexpected type at path %s", key)
		}
	}
	
	results := make(SearchResults, 0)
	
	// Extract objects with timestamps
	if objects, ok := current.([]interface{}); ok {
		for _, obj := range objects {
			if objMap, ok := obj.(map[string]interface{}); ok {
				result := wc.extractSearchResult(objMap)
				result.Source = SourceConversation
				
				// Extract timestamp if present
				if tsStr, ok := objMap["timestamp"].(string); ok {
					if ts, err := time.Parse(time.RFC3339, tsStr); err == nil {
						result.Timestamp = &ts
					}
				}
				
				results = append(results, result)
			}
		}
	}
	
	return results, nil
}

func (wc *WeaviateClient) extractSearchResult(objMap map[string]interface{}) SearchResult {
	result := SearchResult{
		Metadata: make(map[string]interface{}),
	}
	
	// Extract _additional fields
	if additional, ok := objMap["_additional"].(map[string]interface{}); ok {
		if id, ok := additional["id"].(string); ok {
			result.ID = id
		}
		if score, ok := additional["score"].(float64); ok {
			result.Score = score
		}
	}
	
	// Extract content fields based on class
	if message, ok := objMap["message"].(string); ok {
		result.Text = message
		result.DocID = result.ID
	} else if content, ok := objMap["content"].(string); ok {
		result.Text = content
		result.DocID = result.ID
	} else if title, ok := objMap["title"].(string); ok {
		result.Text = title
		result.DocID = result.ID
	}
	
	// Store other fields as metadata
	for k, v := range objMap {
		if k != "_additional" && k != "message" && k != "content" && k != "title" {
			result.Metadata[k] = v
		}
	}
	
	return result
}

func (wc *WeaviateClient) createObject(ctx context.Context, object map[string]interface{}) (string, error) {
	jsonBody, err := json.Marshal(object)
	if err != nil {
		return "", fmt.Errorf("failed to marshal object: %w", err)
	}
	
	url := fmt.Sprintf("%s/v1/objects", wc.host)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	if wc.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+wc.apiKey)
	}
	
	resp, err := wc.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}
	
	if id, ok := result["id"].(string); ok {
		return id, nil
	}
	
	return "", fmt.Errorf("no ID in response")
}

func (wc *WeaviateClient) executeBatchDelete(ctx context.Context, query string) (int, error) {
	// Execute batch delete mutation
	_, err := wc.executeGraphQL(ctx, query)
	if err != nil {
		return 0, err
	}
	
	// Count deleted objects (placeholder - actual implementation depends on response structure)
	return 0, nil
}
