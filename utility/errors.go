package utility

import (
	"fmt"
	"strings"
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

type GuestNoIpaddressError struct {
}
type VolumeHasTaskError struct {
	ErrArgs
}

func (e VolumeHasTaskError) Format() string {
	return "volume %s has task"
}
func (e GuestNoIpaddressError) Error() string {
	return "guest has no ipaddress"
}

type GuestHasNoIpaddressError struct {
	ErrArgs
}
type PingLossPackage struct {
	ErrArgs
}

func (e PingLossPackage) Error() string {
	return "ping loss %v packages"
}
func (e GuestHasNoIpaddressError) Format() string {
	return "guest has no ipaddress %s"
}

func NewActionError(args ...interface{}) ActionError {
	return ActionError{ErrArgs{Args: args}}
}
func NewGuestNoIpaddressError() GuestNoIpaddressError {
	return GuestNoIpaddressError{}
}
func NewVolumeHasTaskError(volumeId string) VolumeHasTaskError {
	return VolumeHasTaskError{
		ErrArgs{Args: []interface{}{volumeId}},
	}
}
func NewGuestHasNoIpaddressError(ipAddress []string) GuestHasNoIpaddressError {
	return GuestHasNoIpaddressError{
		ErrArgs{Args: []interface{}{strings.Join(ipAddress, ", ")}},
	}
}
func NewPingLossPackage(lossed int) PingLossPackage {
	return PingLossPackage{ErrArgs{Args: []interface{}{lossed}}}
}
