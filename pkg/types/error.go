package types

import "fmt"

const (
	ErrChannelClosed = "CHANNEL_CLOSED"
)

// Error custom error type
type Error struct {
	code string
	msg  string
}

// NewError creates a new Error with the given code and message
func NewError(code, msg string) *Error {
	return &Error{
		code: code,
		msg:  msg,
	}
}

// Error returns string representation of the error
func (e *Error) Error() string {
	return fmt.Sprintf("%s - %s", e.code, e.msg)
}

// IsCode returns true if the code of the error matches the specified code
func (e *Error) IsCode(code string) bool {
	return e.code == code
}
