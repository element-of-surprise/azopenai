// Package errors contains some custom error handling for the SDK. In addition to make
// renaming unnecessary for those who import this package, it also provides all the
// functions from the "errors" package at version 1.20 .
package errors

import (
	"errors"
)

// New returns an error that formats as the given text. Each call to New returns a distinct
// error value even if the text is identical.
func New(text string) error {
	return errors.New(text)
}

// As calls errors.As function from the "errors" package.
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// Is calls errors.Is function from the "errors" package.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// Join returns an error that wraps the given errors. Any nil error values are discarded.
// Join returns nil if errs contains no non-nil values. The error formats as the concatenation
// of the strings obtained by calling the Error method of each element of errs, with a newline
// between each string.
func Join(errs ...error) error {
	return errors.Join(errs...)
}

// Unwrap calls errors.Unwrap function from the "errors" package.
func Unwrap(err error) error {
	return errors.Unwrap(err)
}

// JSON implements Error in order to JSON decode an error message from the server.
type JSON struct {
	// JSON is the JSON error repsonse received for deeper introspection.
	JSON map[string]any
	// Message is the error message received in its raw format.
	Message string
	// StatusCode is the HTTP error code received.
	StatusCode int
}

// Error implements error.
func (j JSON) Error() string {
	return j.Message
}

// StatusCode implements error when we receive a non-200 response from the server
// and the message is not JSON decodable.
type StatusCode struct {
	// Message is the error message received.
	Message string
	// StatusCode is the HTTP error code received.
	StatusCode int
}

// Error implements error.
func (s StatusCode) Error() string {
	return s.Message
}
