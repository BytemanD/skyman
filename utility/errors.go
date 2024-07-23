package utility

import (
	"fmt"
	"strings"
)

type ErrArgs struct {
	Args []interface{}
}

type ActionNotFinishedError ErrArgs
type GuestNoIpaddressError ErrArgs
type VolumeHasTaskError ErrArgs
type GuestHasNoIpaddressError ErrArgs
type PingLossPackage ErrArgs

func (e ActionNotFinishedError) Error() string {
	return fmt.Sprintf("action %s is not finished", e.Args...)
}
func (e GuestNoIpaddressError) Error() string {
	return "guest has no ipaddress"
}
func (e GuestHasNoIpaddressError) Error() string {
	return fmt.Sprintf("guest has no ipaddress %s", e.Args...)
}
func (e PingLossPackage) Error() string {
	return fmt.Sprintf("ping loss %v packages", e.Args...)
}
func (e VolumeHasTaskError) Error() string {
	return fmt.Sprintf("volume %s has task", e.Args...)
}

func NewActionNotFinishedError(args ...interface{}) ActionNotFinishedError {
	return ActionNotFinishedError{Args: args}
}
func NewGuestNoIpaddressError() GuestNoIpaddressError {
	return GuestNoIpaddressError{}
}
func NewVolumeHasTaskError(volumeId string) VolumeHasTaskError {
	return VolumeHasTaskError{Args: []interface{}{volumeId}}
}
func NewGuestHasNoIpaddressError(ipAddress []string) GuestHasNoIpaddressError {
	return GuestHasNoIpaddressError{Args: []interface{}{strings.Join(ipAddress, ", ")}}
}
func NewPingLossPackage(lossed int) PingLossPackage {
	return PingLossPackage{Args: []interface{}{lossed}}
}
