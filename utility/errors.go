package utility

import (
	"fmt"
	"strings"
)

type ErrArgs struct {
	Args []any
}

type ActionNotFinishedError ErrArgs
type GuestNoIpaddressError ErrArgs
type VolumeHasTaskError ErrArgs
type GuestHasNoIpaddressError ErrArgs
type PingLossPackage ErrArgs
type ServerNotStopped ErrArgs
type SnapshotIsNotAvailable ErrArgs
type ServerNotBooted ErrArgs
type ImageNotActive ErrArgs

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
func (e ServerNotStopped) Error() string {
	return fmt.Sprintf("server %s is not stopped", e.Args...)
}
func (e SnapshotIsNotAvailable) Error() string {
	return fmt.Sprintf("snapshot %s is not available", e.Args...)
}
func (e ImageNotActive) Error() string {
	return fmt.Sprintf("image %s is not active", e.Args...)
}

func (e ServerNotBooted) Error() string {
	return fmt.Sprintf("server %s is not booted", e.Args...)
}
func NewActionNotFinishedError(args ...any) ActionNotFinishedError {
	return ActionNotFinishedError{Args: args}
}
func NewGuestNoIpaddressError() GuestNoIpaddressError {
	return GuestNoIpaddressError{}
}
func NewVolumeHasTaskError(volumeId string) VolumeHasTaskError {
	return VolumeHasTaskError{Args: []any{volumeId}}
}
func NewGuestHasNoIpaddressError(ipAddress []string) GuestHasNoIpaddressError {
	return GuestHasNoIpaddressError{Args: []any{strings.Join(ipAddress, ", ")}}
}
func NewPingLossPackage(lossed int) PingLossPackage {
	return PingLossPackage{Args: []any{lossed}}
}
func NewServerNotStopped(serverId string) ServerNotStopped {
	return ServerNotStopped{Args: []any{serverId}}
}
func NewSnapshotIsNotAvailable(snapshotId string) SnapshotIsNotAvailable {
	return SnapshotIsNotAvailable{Args: []any{snapshotId}}
}
func NewServerNotBootedError(serverId string) ServerNotBooted {
	return ServerNotBooted{Args: []any{serverId}}
}
func NewImageNotActiveError(imageId string) ImageNotActive {
	return ImageNotActive{Args: []any{imageId}}
}
