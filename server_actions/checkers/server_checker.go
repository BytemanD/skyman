package checkers

import (
	"fmt"

	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/neutron"
	"github.com/BytemanD/skyman/openstack/model/nova"
)

type ServerChecker struct {
	Client   *openstack.Openstack
	ServerId string
}

func (c ServerChecker) MakesureServerRunning() error {
	server, err := c.Client.NovaV2().Server().Show(c.ServerId)
	if err != nil {
		return fmt.Errorf("get server failed: %s", err)
	}
	if server.IsRunning() {
		return nil
	}
	return fmt.Errorf("server is not running (%s)", server.GetPowerState())
}
func (c ServerChecker) MakesureServerStopped() error {
	server, err := c.Client.NovaV2().Server().Show(c.ServerId)
	if err != nil {
		return fmt.Errorf("get server failed: %s", err)
	}
	if server.IsStopped() {
		return nil
	}
	return fmt.Errorf("server is not stopped (%s)", server.GetPowerState())
}
func (c ServerChecker) MakesureInterfaceExist(attachment *nova.InterfaceAttachment) error {
	interfaces, err := c.Client.NovaV2().Server().ListInterfaces(c.ServerId)
	if err != nil {
		return err
	}
	for _, vif := range interfaces {
		if vif.PortId == attachment.PortId {
			console.Info("[%s] server has interface: %s", c.ServerId, attachment.PortId)
			return nil
		}
	}
	return fmt.Errorf("server has no interface: %s", attachment.PortId)
}
func (c ServerChecker) MakesureInterfaceNotExists(port *neutron.Port) error {
	interfaces, err := c.Client.NovaV2().Server().ListInterfaces(c.ServerId)
	if err != nil {
		return err
	}
	for _, vif := range interfaces {
		if vif.PortId == port.Id {
			return fmt.Errorf("server has interface: %s", port.Id)
		}
	}
	console.Info("[%s] has no interface: %s", c.ServerId, port.Id)
	return nil
}

func (c ServerChecker) MakesureVolumeExist(attachment *nova.VolumeAttachment) error {
	volumes, err := c.Client.NovaV2().Server().ListVolumes(c.ServerId)
	if err != nil {
		return err
	}
	for _, vol := range volumes {
		if vol.VolumeId == attachment.VolumeId {
			return nil
		}
	}
	return fmt.Errorf("server has not volume: %s", attachment.VolumeId)
}
func (c ServerChecker) MakesureVolumeNotExists(attachment *nova.VolumeAttachment) error {
	volumes, err := c.Client.NovaV2().Server().ListVolumes(c.ServerId)
	if err != nil {
		return err
	}
	for _, vol := range volumes {
		if vol.VolumeId == attachment.VolumeId {
			return fmt.Errorf("server has volume: %s", attachment.VolumeId)
		}
	}
	console.Info("[%s] server has not volume: %s", c.ServerId, attachment.VolumeId)
	return nil
}

func (c ServerChecker) MakesureVolumeSizeIs(attachment *nova.VolumeAttachment, size uint) error {
	volume, err := c.Client.CinderV2().Volume().Show(attachment.VolumeId)
	if err != nil {
		return err
	}
	if volume.Size != size {
		return fmt.Errorf("volume size is %d, not %d", volume.Size, size)
	}
	return nil
}
