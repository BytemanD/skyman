package internal

import "errors"

var ErrResourceNotFound = errors.New("resource not found")
var ErrResourceMulti = errors.New("multiple resources")

var ErrActionNotFinish = errors.New("action is not finished")
var ErrActionFailed = errors.New("action is failed")
var ErrServerNotStopped = errors.New("server is not stopped")

var ErrServerIsNotDeleted = errors.New("server is not deleted")
var ErrServerIsError = errors.New("server is error")
var ErrServerStatusNotExpect = errors.New("server status is not expect")
