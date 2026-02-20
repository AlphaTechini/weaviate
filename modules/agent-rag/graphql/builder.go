package graphql

import (
	"fmt"
	"strings"
)

// QueryBuilder builds Weaviate GraphQL queries for agent-rag
type QueryBuilder struct {
	className string
	limit     int
}

// NewQueryBuilder creates a new query builder for a specific class
func NewQueryBuilder(className string, limit int) *QueryBuilder {
	return &QueryBuilder{
		className: className,
		limit:     limit,
	}
}

// HybridQuery builds a hybrid search query (vector + keyword)
func (qb *QueryBuilder) HybridQuery(queryText string, vector []float32, alpha float64) string {
	vectorStr := formatVector(vector)
	escapedText := escapeGraphQL(queryText)
	
	query := fmt.Sprintf(`{ Get { %s(hybrid:{query:"%s",vector:[%s],alpha:%.2f},limit:%d){_additional{id,score,vector}%s}}}`,
		qb.className, escapedText, vectorStr, alpha, qb.limit, qb.buildFieldList())
	
	return query
}

// NearVectorQuery builds a nearVector search query
func (qb *QueryBuilder) NearVectorQuery(vector []float32, certainty float64) string {
	vectorStr := formatVector(vector)
	
	certaintyClause := ""
	if certainty > 0 {
		certaintyClause = fmt.Sprintf(",certainty:%.4f", certainty)
	}
	
	query := fmt.Sprintf(`{ Get { %s(nearVector:{vector:[%s]%s},limit:%d){_additional{id,score,vector}%s}}}`,
		qb.className, vectorStr, certaintyClause, qb.limit, qb.buildFieldList())
	
	return query
}

// WhereFilter represents a Weaviate where filter
type WhereFilter struct {
	Operator string        `json:"operator"`
	Path     []string      `json:"path"`
	Value    interface{}   `json:"value"`
	Operands []WhereFilter `json:"operands,omitempty"`
}

// BuildWhereClause converts a WhereFilter to GraphQL where clause
func (qb *QueryBuilder) BuildWhereClause(filter WhereFilter) string {
	return buildWhereRecursive(filter)
}

func buildWhereRecursive(filter WhereFilter) string {
	parts := []string{}
	
	if filter.Operator != "" {
		parts = append(parts, fmt.Sprintf("operator:%s", filter.Operator))
	}
	
	if len(filter.Path) > 0 {
		pathStr := strings.Join(filter.Path, ",")
		parts = append(parts, fmt.Sprintf("path:[%s]", pathStr))
	}
	
	if filter.Value != nil {
		parts = append(parts, fmt.Sprintf("value:%v", filter.Value))
	}
	
	if len(filter.Operands) > 0 {
		operandStrs := []string{}
		for _, operand := range filter.Operands {
			operandStrs = append(operandStrs, buildWhereRecursive(operand))
		}
		parts = append(parts, fmt.Sprintf("operands:[%s]", strings.Join(operandStrs, ",")))
	}
	
	return "{" + strings.Join(parts, ",") + "}"
}

// NearVectorWithWhere builds a nearVector query with where filter
func (qb *QueryBuilder) NearVectorWithWhere(vector []float32, where WhereFilter) string {
	vectorStr := formatVector(vector)
	whereClause := qb.BuildWhereClause(where)
	
	query := fmt.Sprintf(`{ Get { %s(nearVector:{vector:[%s]},where:%s,limit:%d){_additional{id,score}%s}}}`,
		qb.className, vectorStr, whereClause, qb.limit, qb.buildFieldList())
	
	return query
}

// ConversationQuery builds a query optimized for conversation retrieval
func (qb *QueryBuilder) ConversationQuery(vector []float32, sinceTime string) string {
	vectorStr := formatVector(vector)
	
	timeFilter := ""
	if sinceTime != "" {
		timeFilter = fmt.Sprintf(`,where:{operator:GreaterThanEqual,path:["timestamp"],valueDate:"%s"}`, sinceTime)
	}
	
	query := fmt.Sprintf(`{ Get { %s(nearVector:{vector:[%s]}%s,limit:%d){_additional{id,score}message,speaker,timestamp}}}`,
		qb.className, vectorStr, timeFilter, qb.limit)
	
	return query
}

// BatchDeleteQuery builds a delete query with where filter
func (qb *QueryBuilder) BatchDeleteQuery(where WhereFilter) string {
	whereClause := qb.BuildWhereClause(where)
	
	query := fmt.Sprintf(`mutation{BatchDelete{objects(class:"%s",where:%s){id}}}`,
		qb.className, whereClause)
	
	return query
}

// Helper functions

func formatVector(vector []float32) string {
	if len(vector) == 0 {
		return ""
	}
	
	strs := make([]string, len(vector))
	for i, v := range vector {
		strs[i] = fmt.Sprintf("%.6f", v)
	}
	return strings.Join(strs, ",")
}

func escapeGraphQL(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	return s
}

func (qb *QueryBuilder) buildFieldList() string {
	if qb.className == "Conversation" {
		return "message,speaker,timestamp"
	}
	return "title,content"
}
