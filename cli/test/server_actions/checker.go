package server_actions

import (
	"fmt"
	"strings"
	"time"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/common"
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
	Client   *openstack.Openstack
	ServerId string
	Host     string
}

func (c QGAChecker) makesureQGAConnected(g *guest.Guest) error {
	logging.Info("[%s] connecting to qga ...", c.ServerId)
	startTime := time.Now()
	for {
		_, err := g.HostName()
		if err == nil {
			logging.Info("[%s] qga connected", g.Domain)
			return nil
		}
		if time.Since(startTime) >= time.Second*time.Duration(common.CONF.Test.QGAChecker.QgaConnectTimeout) {
			return fmt.Errorf("connect qga timeout")
		}
		logging.Debug("[%s] get hostname failed: %s", g.Domain, err)
		time.Sleep(time.Second * 5)
	}
}

func (c QGAChecker) MakesureServerBooted() error {
	startTime := time.Now()
	serverGuest := guest.Guest{Connection: c.Host, Domain: c.ServerId}
	logging.Info("[%s] waiting server booted ...", c.ServerId)
	logging.Info("[%s] connecting to guest ...", c.ServerId)
	for {
		if err := serverGuest.Connect(); err == nil {
			if serverGuest.IsRunning() {
				break
			}
		}
		if time.Since(startTime) >= time.Second*time.Duration(common.CONF.Test.QGAChecker.GuestConnectTimeout) {
			return fmt.Errorf("connect guest timeout")
		}
		time.Sleep(time.Second * 5)
	}
	logging.Info("[%s] guest is running", c.ServerId)
	return c.makesureQGAConnected(&serverGuest)
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
