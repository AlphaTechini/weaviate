package graphql

// GetResultPath returns the GraphQL path to access results for a given class
func GetResultPath(className string) []string {
	return []string{"Get", className}
}
