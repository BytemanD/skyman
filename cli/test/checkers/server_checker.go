package checkers

import (
	"fmt"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/neutron"
	"github.com/BytemanD/skyman/openstack/model/nova"
)

type ServerChecker struct {
	Client   *openstack.Openstack
	ServerId string
}

func (c ServerChecker) MakesureServerRunning() error {
	server, err := c.Client.NovaV2().Servers().Show(c.ServerId)
	if err != nil {
		return fmt.Errorf("get server failed: %s", err)
	}
	if server.IsRunning() {
		return nil
	}
	return fmt.Errorf("server is not running (%s)", server.GetPowerState())
}

func (c ServerChecker) MakesureInterfaceExist(attachment *nova.InterfaceAttachment) error {
	interfaces, err := c.Client.NovaV2().Servers().ListInterfaces(c.ServerId)
	if err != nil {
		return err
	}
	for _, vif := range interfaces {
		if vif.PortId == attachment.PortId {
			logging.Info("[%s] server has interface: %s", c.ServerId, attachment.PortId)
			return nil
		}
	}
	return fmt.Errorf("server has not interface: %s", attachment.PortId)
}
func (c ServerChecker) MakesureInterfaceNotExists(port *neutron.Port) error {
	interfaces, err := c.Client.NovaV2().Servers().ListInterfaces(c.ServerId)
	if err != nil {
		return err
	}
	for _, vif := range interfaces {
		if vif.PortId == port.Id {
			return fmt.Errorf("server has interface: %s", port.Id)
		}
	}
	logging.Info("[%s] has no port %s", c.ServerId, port.Id)
	return nil
}

func (c ServerChecker) MakesureVolumeExist(attachment *nova.VolumeAttachment) error {
	volumes, err := c.Client.NovaV2().Servers().ListVolumes(c.ServerId)
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
	volumes, err := c.Client.NovaV2().Servers().ListVolumes(c.ServerId)
	if err != nil {
		return err
	}
	for _, vol := range volumes {
		if vol.VolumeId == attachment.VolumeId {
			return fmt.Errorf("server has volume: %s", attachment.VolumeId)
		}
	}
	logging.Info("[%s] server has not volume: %s", c.ServerId, attachment.VolumeId)
	return nil
}
