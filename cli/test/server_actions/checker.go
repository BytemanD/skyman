package server_actions

import (
	"fmt"
	"time"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/guest"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/nova"
)

type ServerChecker interface {
	MakesureServerBooted(server *nova.Server) error
	MakesureHostname(server *nova.Server) error
	MakesureInterfaceAttached(server *nova.Server, address string) error
	MakesureInterfaceDetached(server *nova.Server, address string) error
	MakesureVolumeAttached(server *nova.Server, device string) error
	MakesureVolumeDetached(server *nova.Server, device string) error
}

type QGAChecker struct {
	Client *openstack.Openstack
}

func (c QGAChecker) MakesureServerBooted(server *nova.Server) error {
	host, err := c.Client.NovaV2().Hypervisors().Found(server.Host)
	if err != nil {
		return fmt.Errorf("get hypervisor failed: %s", err)
	}

	logging.Info("[%s] server host ip is %s", server.Id, host.HostIp)
	startTime := time.Now()
	serverGuest := guest.Guest{Connection: host.HostIp, Domain: server.Id}
	for {
		serverGuest.Connect()
		if serverGuest.IsRunning() {
			logging.Info("[%s] domain is running", server.Id)
			return nil
		}
		if time.Since(startTime) >= time.Second*300 {
			break
		}
		time.Sleep(time.Second * 5)
	}
	return fmt.Errorf("connect guest timeout")
}

func (c QGAChecker) MakesureHostname(server *nova.Server, hostname string) error {
	host, err := c.Client.NovaV2().Hypervisors().Found(server.Host)
	if err != nil {
		return fmt.Errorf("get hypervisor failed: %s", err)
	}
	serverGuest := guest.Guest{Connection: host.HostIp, Domain: server.Id}
	serverGuest.Connect()
	result := serverGuest.Exec("hostname", true)
	if result.Failed {
		return fmt.Errorf("run qga command failed")
	}
	logging.Info("[%s] guest hostname is %s", server.Id, result.OutData)
	if hostname != result.OutData {
		return fmt.Errorf("hostname is %s, not %s", result.OutData, hostname)
	}
	return nil
}
