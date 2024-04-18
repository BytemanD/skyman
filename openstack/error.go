package openstack

import "fmt"

type OpenstackError struct {
	Message string
}

func (err OpenstackError) Error() string {
	if err.Message != "" {
		return err.Message
	}
	return "found multi items"
}

type MultiItems struct {
	OpenstackError
}

type ItemNotFound struct {
	OpenstackError
}

func MultiItemsError(format string, args ...interface{}) MultiItems {
	return MultiItems{
		OpenstackError: OpenstackError{Message: fmt.Sprintf(format, args...)},
	}
}

func ItemNotFoundError(format string, args ...interface{}) ItemNotFound {
	return ItemNotFound{
		OpenstackError: OpenstackError{Message: fmt.Sprintf(format, args...)},
	}
}
