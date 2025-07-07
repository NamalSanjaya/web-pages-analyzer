package http

import "fmt"

type httpError struct {
	StatusCode int
	Message    string
}

func NewHttpError(statusCode int, message string) *httpError {
	return &httpError{StatusCode: statusCode, Message: message}
}

func (e *httpError) Error() string {
	return fmt.Sprintf("HTTP call related error, status code %d: %s", e.StatusCode, e.Message)
}

func NewHttpErrorFromErr(err error) (*httpError, bool) {
	httpErr, ok := err.(*httpError)
	return httpErr, ok
}
