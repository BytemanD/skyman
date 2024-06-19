package utility

import (
	"fmt"
)

type ErrArgs struct {
	Args []interface{}
}

func (e ErrArgs) Format() string {
	return "unknown error"
}

func (e ErrArgs) Error() string {
	return fmt.Sprintf(e.Format(), e.Args...)
}

type ActionError struct {
	ErrArgs
}

func (e ActionError) Format() string {
	return "action %s is error"
}

func NewActionError(args ...interface{}) ActionError {
	return ActionError{ErrArgs{Args: args}}
}
