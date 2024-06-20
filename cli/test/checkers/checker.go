package checkers

import (
	"fmt"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/neutron"
	"github.com/BytemanD/skyman/openstack/model/nova"
)

type ServerCheckerInterface interface {
	MakesureServerRunning() error
	// MakesureHostname(hostname string) error
	MakesureInterfaceExist(attachment *nova.InterfaceAttachment) error
	MakesureInterfaceNotExists(port *neutron.Port) error
	MakesureVolumeExist(attachment *nova.VolumeAttachment) error
	MakesureVolumeNotExists(attachment *nova.VolumeAttachment) error
	MakesureVolumeSizeIs(attachment *nova.VolumeAttachment, size uint) error
}

type ServerCheckers []ServerCheckerInterface

func (checkers ServerCheckers) MakesureServerRunning() error {
	for _, checker := range checkers {
		if err := checker.MakesureServerRunning(); err != nil {
			return err
		}
	}
	return nil
}
func (checkers ServerCheckers) MakesureVolumeExist(attachment *nova.VolumeAttachment) error {
	for _, checker := range checkers {
		if err := checker.MakesureVolumeExist(attachment); err != nil {
			return err
		}
	}
	return nil
}
func (checkers ServerCheckers) MakesureVolumeNotExists(attachment *nova.VolumeAttachment) error {
	for _, checker := range checkers {
		if err := checker.MakesureVolumeNotExists(attachment); err != nil {
			return err
		}
	}
	return nil
}

func (checkers ServerCheckers) MakesureInterfaceExist(attachment *nova.InterfaceAttachment) error {
	for _, checker := range checkers {
		if err := checker.MakesureInterfaceExist(attachment); err != nil {
			return err
		}
	}
	return nil
}
func (checkers ServerCheckers) MakesureInterfaceNotExists(port *neutron.Port) error {
	for _, checker := range checkers {
		if err := checker.MakesureInterfaceNotExists(port); err != nil {
			return err
		}
	}
	return nil
}
func (checkers ServerCheckers) MakesureVolumeSizeIs(attachment *nova.VolumeAttachment, size uint) error {
	for _, checker := range checkers {
		if err := checker.MakesureVolumeSizeIs(attachment, size); err != nil {
			return err
		}
	}
	return nil
}

func GetServerCheckers(client *openstack.Openstack, server *nova.Server) (ServerCheckers, error) {
	checkers := []ServerCheckerInterface{}
	checkers = append(checkers, ServerChecker{Client: client, ServerId: server.Id})
	if common.CONF.Test.QGAChecker.Enabled {
		qgaChecker, err := GetQgaChecker(client, server)
		if err != nil {
			return nil, fmt.Errorf("get qga checker failed: %s", err)
		}
		checkers = append(checkers, qgaChecker)
	}
	return checkers, nil
}
