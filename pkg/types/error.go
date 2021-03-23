package types

import "fmt"

const (
	ErrChannelClosed = "CHANNEL_CLOSED"
)

type Error struct {
	code string
	msg  string
}

func NewError(code, msg string) *Error {
	return &Error{
		code: code,
		msg:  msg,
	}
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s - %s", e.code, e.msg)
}

func (e *Error) IsCode(code string) bool {
	return e.code == code
}
