package utility

import (
	"fmt"
	"reflect"
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

// TODO: move to easygo

func Equals(err1 error, err2 error) bool {
	if err1 == nil && err2 == nil {
		return true
	}
	return IsError(err1, reflect.ValueOf(err1).Type().Name())
}

func IsError(err error, errorName string) bool {
	return reflect.ValueOf(err).Type().Name() == errorName
}
