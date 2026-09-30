package api

import (
	"fmt"
)

// Error is an error returned by the Polyteia API.
type Error struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("polygo api error: status -> %d, code -> %s, message -> %s", e.Status, e.Code, e.Message)
}

// newError builds an Error for a failed response, falling back to the raw body when it carries no message.
func newError(status int, apiErr *Error, body []byte) *Error {
	e := &Error{Status: status}
	if apiErr != nil {
		e.Code = apiErr.Code
		e.Message = apiErr.Message
	}

	if e.Message == "" {
		e.Message = string(body)
	}

	return e
}
