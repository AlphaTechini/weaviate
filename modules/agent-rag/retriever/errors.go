package retriever

import "errors"

var (
	// ErrInvalidWeight is returned when weight values are out of range
	ErrInvalidWeight = errors.New("weight must be between 0 and 1")
	
	// ErrInvalidHalfLife is returned when half-life is not positive
	ErrInvalidHalfLife = errors.New("half-life must be positive")
	
	// ErrInvalidMinWeight is returned when minimum weight is out of range
	ErrInvalidMinWeight = errors.New("minimum weight must be between 0 and 1")
	
	// ErrNoResults is returned when search yields no results
	ErrNoResults = errors.New("no results found")
	
	// ErrIndexNotFound is returned when specified index doesn't exist
	ErrIndexNotFound = errors.New("index not found")
	
	// ErrVectorMismatch is returned when query vector dimension doesn't match
	ErrVectorMismatch = errors.New("vector dimension mismatch")
	
	// ErrClosedRetriever is returned when operating on a closed retriever
	ErrClosedRetriever = errors.New("retriever is closed")
)
