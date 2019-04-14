package octop

import "errors"

var (
	//ErrClosed connection closed
	ErrClosed = errors.New("CONNECTION_CLOSED")
)
