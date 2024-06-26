package checkers

import (
	"fmt"
	"strings"
	"time"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/guest"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/neutron"
	"github.com/BytemanD/skyman/openstack/model/nova"
)

type QGAChecker struct {
	Client   *openstack.Openstack
	ServerId string
	Host     string
}

func (c QGAChecker) makesureQGAConnected(g *guest.Guest) error {
	logging.Info("[%s] connecting to qga ...", c.ServerId)
	if err := g.ConnectToQGA(common.CONF.Test.QGAChecker.QgaConnectTimeout); err != nil {
		return err
	}
	logging.Info("[%s] qga connected", g.Domain)
	return nil
}

func (c QGAChecker) MakesureHostname(hostname string) error {
	serverGuest := guest.Guest{Connection: c.Host, Domain: c.ServerId}
	serverGuest.Connect()
	result := serverGuest.Exec("hostname", true)
	if result.Failed {
		return fmt.Errorf("run qga command failed")
	}
	guestHostname := strings.TrimSpace(result.OutData)
	logging.Info("[%s] guest hostname is %s", c.ServerId, guestHostname)
	if guestHostname != hostname {
		return fmt.Errorf("hostname is %s, not %s", guestHostname, hostname)
	}
	return nil
}
func (c QGAChecker) MakesureServerRunning() error {
	startTime := time.Now()
	serverGuest := guest.Guest{Connection: c.Host, Domain: c.ServerId}
	logging.Info("[%s] connecting to guest ...", c.ServerId)
	for {
		if err := serverGuest.Connect(); err == nil {
			if serverGuest.IsRunning() {
				logging.Info("[%s] guest is running ...", c.ServerId)
				break
			}
		}
		if time.Since(startTime) >= time.Second*time.Duration(common.CONF.Test.QGAChecker.GuestConnectTimeout) {
			return fmt.Errorf("connect guest timeout")
		}
		time.Sleep(time.Second * 5)
	}
	return c.makesureQGAConnected(&serverGuest)
}
func (c QGAChecker) MakesureInterfaceExist(attachment *nova.InterfaceAttachment) error {
	serverGuest := guest.Guest{Connection: c.Host, Domain: c.ServerId}
	serverGuest.Connect()
	ipaddrs := serverGuest.GetIpaddrs()
	logging.Debug("[%s] found ip addresses: %s", c.ServerId, ipaddrs)

	for _, fixedIpaddr := range attachment.FixedIps {
		found := false
		for _, ipaddr := range ipaddrs {
			if ipaddr == fixedIpaddr.IpAddress {
				logging.Info("[%s] ip address %s exists on guest", c.ServerId, fixedIpaddr.IpAddress)
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("ip address %s not found in guest, found %v",
				fixedIpaddr.IpAddress, ipaddrs)
		}
	}
	return nil
}
func (c QGAChecker) MakesureInterfaceNotExists(port *neutron.Port) error {
	serverGuest := guest.Guest{Connection: c.Host, Domain: c.ServerId}
	serverGuest.Connect()
	ipaddrs := serverGuest.GetIpaddrs()
	logging.Debug("[%s] found ip addresses: %s", c.ServerId, ipaddrs)

	for _, fixedIp := range port.FixedIps {
		for _, ipaddr := range ipaddrs {
			if ipaddr == fixedIp.IpAddress {
				return fmt.Errorf("ip address %s exists on guest", fixedIp.IpAddress)
			}
		}
	}
	logging.Info("[%s] ip address: %s not exists on guest", c.ServerId, port.GetFixedIpaddress())
	return nil
}
func (c QGAChecker) MakesureVolumeExist(attachment *nova.VolumeAttachment) error {
	serverGuest := guest.Guest{Connection: c.Host, Domain: c.ServerId}
	serverGuest.Connect()
	guestBlockDevices, err := serverGuest.GetBlockDevices()
	if err != nil {
		return fmt.Errorf("get block devices failed: %s", err)
	}
	logging.Debug("[%s] found block devices: %s", c.ServerId, guestBlockDevices.GetAllNames())
	for _, blockDevice := range guestBlockDevices {
		if blockDevice.Name == attachment.Device {
			logging.Info("[%s] guest block device %s exists", c.ServerId, attachment.Device)
			return nil
		}
	}
	return fmt.Errorf("block device %s not found in guest, found %v",
		attachment.Device, guestBlockDevices.GetAllNames())
}
func (c QGAChecker) MakesureVolumeNotExists(attachment *nova.VolumeAttachment) error {
	serverGuest := guest.Guest{Connection: c.Host, Domain: c.ServerId}
	serverGuest.Connect()
	guestBlockDevices, err := serverGuest.GetBlockDevices()
	if err != nil {
		return fmt.Errorf("get block devices failed: %s", err)
	}
	logging.Debug("[%s] found block devices: %s", c.ServerId, guestBlockDevices.GetAllNames())
	for _, blockDevice := range guestBlockDevices {
		if blockDevice.Name == attachment.Device {
			return fmt.Errorf("block device %s not found in guest, found %v",
				attachment.Device, guestBlockDevices.GetAllNames())
		}
	}
	logging.Info("[%s] block device %s not exists on guest", c.ServerId, attachment.Device)
	return nil
}

func (c QGAChecker) MakesureVolumeSizeIs(attachment *nova.VolumeAttachment, size uint) error {
	server, err := c.Client.NovaV2().Servers().Show(c.ServerId)
	if err != nil {
		return err
	}
	if server.IsShelved() {
		logging.Warning("[%s] status is %s, skip to check volume size on guest", c.ServerId, server.Status)
		return nil
	}
	serverGuest := guest.Guest{Connection: c.Host, Domain: c.ServerId}
	serverGuest.Connect()
	guestBlockDevices, err := serverGuest.GetBlockDevices()
	if err != nil {
		return fmt.Errorf("get block devices failed: %s", err)
	}
	for _, blockDevice := range guestBlockDevices {
		if blockDevice.Name == attachment.Device {
			logging.Info("[%s] size of block %s is: %s", c.ServerId, attachment.Device, blockDevice.Size)
			if blockDevice.Size == fmt.Sprintf("%dG", size) {
				return nil
			} else {
				return fmt.Errorf("block size is %s, not %d", blockDevice.Size, size)
			}
		}
	}
	return fmt.Errorf("block device %s not exists on guest", attachment.Device)
}

func GetQgaChecker(client *openstack.Openstack, server *nova.Server) (*QGAChecker, error) {
	host, err := client.NovaV2().Hypervisors().Found(server.Host)
	if err != nil {
		return nil, fmt.Errorf("get hypervisor failed: %s", err)
	}
	logging.Info("[%s] server host ip is %s", server.Id, host.HostIp)
	return &QGAChecker{Client: client, ServerId: server.Id, Host: host.HostIp}, nil
}
