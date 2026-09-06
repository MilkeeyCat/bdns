package wire

import "errors"

var (
	ErrInvalidMessage = errors.New("invalid message")
	ErrShortMessage   = errors.New("short message")
)
