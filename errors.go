package acmd

import "fmt"

// ErrCode contains an int to be returned as an exit code.
type ErrCode int

func (e ErrCode) Error() string {
	return fmt.Sprintf("code %d", int(e))
}
