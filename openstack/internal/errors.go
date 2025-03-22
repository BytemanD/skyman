package internal

import "errors"

var ErrResourceNotFound = errors.New("resource not found")
var ErrResourceMulti = errors.New("multiple resources")

var ErrActionNotFinish = errors.New("action is not finished")
var ErrActionFailed = errors.New("action is failed")
var ErrServerNotStopped = errors.New("server is not stopped")
