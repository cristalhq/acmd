package acmd

import (
	"errors"
	"fmt"
)

var ErrNoArgs = errors.New("no args provided")

// ErrCode is a number to be returned as an exit code.
type ErrCode int

func (e ErrCode) Error() string {
	return fmt.Sprintf("code %d", int(e))
}
