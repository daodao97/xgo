package xrequest

import "fmt"

type RequestError struct {
	Message string
	Err     error
}

func (e *RequestError) Error() string {
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

func NewRequestError(message string, err error) *RequestError {
	return &RequestError{
		Message: message,
		Err:     err,
	}
}
