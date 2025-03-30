package checkers

import (
	"fmt"
	"strings"
	"time"

	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/guest"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/neutron"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/utility"
	"github.com/samber/lo"
)

type QGAChecker struct {
	Client              *openstack.Openstack
	ServerId            string
	Host                string
	GuestConnectTimeout int
	QgaConnectTimeout   int
}

func (c *QGAChecker) SetGuestConnectTimeout(timeout int) {
	c.GuestConnectTimeout = timeout
}
func (c *QGAChecker) SetQgaConnectTimeout(timeout int) {
	c.QgaConnectTimeout = timeout
}

func (c QGAChecker) makesureQGAConnected(g *guest.Guest) error {
	console.Info("[%s] connecting to qga ...", c.ServerId)
	if err := g.ConnectToQGA(c.QgaConnectTimeout); err != nil {
		return err
	}
	console.Info("[%s] qga connected", g.Domain)
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
	console.Info("[%s] guest hostname is %s", c.ServerId, guestHostname)
	if guestHostname != hostname {
		return fmt.Errorf("hostname is %s, not %s", guestHostname, hostname)
	}
	return nil
}
func (c QGAChecker) MakesureServerRunning() error {
	startTime := time.Now()
	serverGuest := guest.Guest{Connection: c.Host, Domain: c.ServerId}
	console.Info("[%s] connecting to guest ...", c.ServerId)
	for {
		if err := serverGuest.Connect(); err == nil {
			if serverGuest.IsRunning() {
				console.Info("[%s] guest is running ...", c.ServerId)
				break
			}
		}
		if time.Since(startTime) >= time.Second*time.Duration(c.GuestConnectTimeout) {
			return fmt.Errorf("connect guest timeout")
		}
		time.Sleep(time.Second * 5)
	}
	return c.makesureQGAConnected(&serverGuest)
}
func (c QGAChecker) MakesureServerStopped() error {
	console.Info("[%s] connecting to guest ...", c.ServerId)
	serverGuest := guest.Guest{Connection: c.Host, Domain: c.ServerId}
	if err := serverGuest.Connect(); err != nil {
		return err
	}
	if serverGuest.IsShutoff() {
		return nil
	}
	return fmt.Errorf("guest is not shutoff")
}
func (c QGAChecker) MakesureInterfaceExist(attachment *nova.InterfaceAttachment) error {
	serverGuest := guest.Guest{Connection: c.Host, Domain: c.ServerId}
	serverGuest.Connect()
	if serverGuest.IsShutoff() {
		console.Warn("[%s] guest is shutoff, skip to check interfaces", c.ServerId)
		return nil
	}

	return utility.RetryWithErrors(
		utility.RetryCondition{
			Timeout:      time.Minute * 5,
			IntervalMin:  time.Second,
			IntervalMax:  time.Second * 10,
			IntervalStep: time.Second,
		},
		[]string{"GuestHasNoIpaddressError"},
		func() error {
			ipaddrs := serverGuest.GetIpaddrs()
			console.Debug("[%s] found ip address on guest: %v", c.ServerId, ipaddrs)
			notFoundFixedIps := lo.Filter(attachment.FixedIps, func(item nova.FixedIp, _ int) bool {
				return !lo.Contains(ipaddrs, item.IpAddress)
			})
			if len(notFoundFixedIps) > 0 {
				notFoundIpaddress := lo.Map(notFoundFixedIps, func(item nova.FixedIp, _ int) string {
					return item.IpAddress
				})
				console.Warn("[%s] ip address %s not exists on guest", c.ServerId, strings.Join(notFoundIpaddress, ","))
				return utility.NewGuestHasNoIpaddressError(notFoundIpaddress)
			}
			return nil
		},
	)
}
func (c QGAChecker) MakesureInterfaceNotExists(port *neutron.Port) error {
	serverGuest := guest.Guest{Connection: c.Host, Domain: c.ServerId}
	serverGuest.Connect()
	if serverGuest.IsShutoff() {
		console.Warn("[%s] guest is shutoff, skip to check interfaces", c.ServerId)
		return nil
	}
	ipaddrs := serverGuest.GetIpaddrs()
	console.Debug("[%s] found ip addresses: %s", c.ServerId, ipaddrs)

	for _, fixedIp := range port.FixedIps {
		for _, ipaddr := range ipaddrs {
			if ipaddr == fixedIp.IpAddress {
				return fmt.Errorf("ip address %s exists on guest", fixedIp.IpAddress)
			}
		}
	}
	console.Info("[%s] ip address: %s not exists on guest", c.ServerId, port.GetFixedIpaddress())
	return nil
}
func (c QGAChecker) MakesureVolumeExist(attachment *nova.VolumeAttachment) error {
	serverGuest := guest.Guest{Connection: c.Host, Domain: c.ServerId}
	serverGuest.Connect()
	if serverGuest.IsShutoff() {
		console.Warn("[%s] guest is shutoff, skip to check block devices", c.ServerId)
		return nil
	}
	guestBlockDevices, err := serverGuest.GetBlockDevices()
	if err != nil {
		return fmt.Errorf("get block devices failed: %s", err)
	}
	console.Debug("[%s] found block devices: %s", c.ServerId, guestBlockDevices.GetAllNames())
	for _, blockDevice := range guestBlockDevices {
		if blockDevice.Name == attachment.Device {
			console.Info("[%s] guest block device %s exists", c.ServerId, attachment.Device)
			return nil
		}
	}
	return fmt.Errorf("block device %s not found in guest, found %v",
		attachment.Device, guestBlockDevices.GetAllNames())
}
func (c QGAChecker) MakesureVolumeNotExists(attachment *nova.VolumeAttachment) error {
	serverGuest := guest.Guest{Connection: c.Host, Domain: c.ServerId}
	serverGuest.Connect()
	if serverGuest.IsShutoff() {
		console.Warn("[%s] guest is shutoff, skip to check block devices", c.ServerId)
		return nil
	}
	if serverGuest.IsShutoff() {
		console.Warn("[%s] guest is shutoff, skip to check block devices", c.ServerId)
		return nil
	}
	guestBlockDevices, err := serverGuest.GetBlockDevices()
	if err != nil {
		return fmt.Errorf("get block devices failed: %s", err)
	}
	console.Debug("[%s] found block devices: %s", c.ServerId, guestBlockDevices.GetAllNames())
	for _, blockDevice := range guestBlockDevices {
		if blockDevice.Name == attachment.Device {
			return fmt.Errorf("block device %s not found in guest, found %v",
				attachment.Device, guestBlockDevices.GetAllNames())
		}
	}
	console.Info("[%s] block device %s not exists on guest", c.ServerId, attachment.Device)
	return nil
}

func (c QGAChecker) MakesureVolumeSizeIs(attachment *nova.VolumeAttachment, size uint) error {
	server, err := c.Client.NovaV2().GetServer(c.ServerId)
	if err != nil {
		return err
	}
	if server.IsShelved() {
		console.Warn("[%s] status is %s, skip to check volume size on guest", c.ServerId, server.Status)
		return nil
	}
	serverGuest := guest.Guest{Connection: c.Host, Domain: c.ServerId}
	serverGuest.Connect()
	if serverGuest.IsShutoff() {
		console.Warn("[%s] guest is shutoff, skip to check block devices", c.ServerId)
		return nil
	}
	guestBlockDevices, err := serverGuest.GetBlockDevices()
	if err != nil {
		return fmt.Errorf("get block devices failed: %s", err)
	}
	for _, blockDevice := range guestBlockDevices {
		if blockDevice.Name == attachment.Device {
			console.Info("[%s] size of block %s is: %s", c.ServerId, attachment.Device, blockDevice.Size)
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
	host, err := client.NovaV2().FindHypervisor(server.Host)
	if err != nil {
		return nil, fmt.Errorf("get hypervisor failed: %s", err)
	}
	console.Info("[%s] server host ip is %s", server.Id, host.HostIp)
	return &QGAChecker{Client: client, ServerId: server.Id, Host: host.HostIp}, nil
}
