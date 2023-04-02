// Package embeddings contains the request and response types for the embeddings API.
package embeddings

import "errors"

// Req represents a request to the embeddings API.
type Req struct {
	// Type is the embedding search to use. This is optional.
	Type string `json:"input_type,omitempty"`
	// Model is the mdoel ID to use. This is optional.
	Model string `json:"model,omitempty"`
	// Input is text to get embeddings for. Must not exceed 2048 tokens (2048 entries).
	// Unless you are embedding code, we suggest replacing newlines (\\n) in your input with a single space,
	// as we have observed inferior results when newlines are present. This is required.
	Input []string `json:"input"`
	// User represents your end-user, which can help monitoring and detecting abuse.
	// This is optional.
	User string `json:"user,omitempty"`
}

// Validate validates the EmbeddingsInput.
func (e Req) Validate() error {
	if len(e.Input) == 0 {
		return errors.New("input is required")
	}
	if len(e.Input) > 2048 {
		return errors.New("input cannot have more than 2048 entries")
	}
	return nil
}

// Data represents an embedding for a single token.
type Data struct {
	// Object is always "embedding".
	Object string `json:"object"`
	// Embedding is the embeddings for the token.
	Embedding []float64 `json:"embedding"`
	// Index is the index of the token in the input.
	Index int `json:"index"`
}

// Resp represents a response from the embeddings API.
type Resp struct {
	// Model is the model used.
	Model string `json:"model"`
	// Data is the embedding data. We guarantee sorted order of the data by index.
	Data []Data `json:"data"`
}
