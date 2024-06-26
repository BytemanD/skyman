package server_actions

import (
	"fmt"
	"strconv"
	"syscall"
	"time"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/guest"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/utility"
)

type ServerLiveMigrate struct {
	ServerActionTest
	clientServer  *nova.Server
	clientGuest   *guest.Guest
	clientPingPid int
}

func (t *ServerLiveMigrate) createClientServer() error {
	clientServerOpt := t.getServerBootOption(fmt.Sprintf("client-%s", t.Server.Name))
	clientServer, err := t.Client.NovaV2().Servers().Create(clientServerOpt)
	if err != nil {
		return fmt.Errorf("create client instance failed: %s", err)
	}
	logging.Info("[%s] creating client server", t.ServerId())
	clientServer, err = t.Client.NovaV2().Servers().WaitBooted(clientServer.Id)
	if err != nil {
		return err
	}
	t.clientServer = clientServer
	logging.Info("[%s] client server (%s) created", t.ServerId(), t.clientServer.Name)
	return nil
}
func (t *ServerLiveMigrate) getClientGuest() (*guest.Guest, error) {
	if t.clientGuest == nil {
		host, err := t.Client.NovaV2().Hypervisors().Found(t.clientServer.Host)
		if err != nil {
			return nil, err
		}
		t.clientGuest = &guest.Guest{Connection: host.HostIp, Domain: t.clientServer.Id}
		t.clientGuest.Connect()
		logging.Info("[%s] connecting to qga ...", t.clientGuest.Domain)
		t.clientGuest.ConnectToQGA(common.CONF.Test.QGAChecker.QgaConnectTimeout)
	}
	return t.clientGuest, nil
}
func (t *ServerLiveMigrate) startPing(targetIp string) error {
	if !common.CONF.Test.QGAChecker.Enabled ||
		!common.CONF.Test.LiveMigrate.PingEnabled {
		return nil
	}
	clientGuest, err := t.getClientGuest()
	if err != nil {
		return err
	}
	ipaddrs := []string{}
	err = utility.RetryWithErrors(
		utility.RetryCondition{
			Timeout:     time.Second * 60,
			IntervalMin: time.Second * 2},
		[]string{"GuestNoIpaddressError"},
		func() error {
			ipaddrs = clientGuest.GetIpaddrs()
			if len(ipaddrs) == 0 {
				return utility.NewGuestNoIpaddressError()
			}
			return nil
		},
	)
	if err != nil {
		return err
	}
	result := clientGuest.Ping(targetIp, common.CONF.Test.LiveMigrate.PingInterval, 0, ipaddrs[0], false)
	if result.ErrData != "" {
		return fmt.Errorf("run ping to %s failed: %s", targetIp, result.ErrData)
	}
	t.clientPingPid = result.Pid
	logging.Info("[%s] ping process pid is: %d", t.ServerId(), t.clientPingPid)
	return nil
}
func (t *ServerLiveMigrate) stopPing() error {
	if !common.CONF.Test.QGAChecker.Enabled ||
		!common.CONF.Test.LiveMigrate.PingEnabled {
		return nil
	}
	if t.clientPingPid == 0 {
		logging.Warning("[%s] ping pid is not exists", t.ServerId())
		return nil
	}
	clientGuest, err := t.getClientGuest()
	if err != nil {
		return err
	}
	clientGuest.Kill(int(syscall.SIGINT), []int{t.clientPingPid})
	return nil
}
func (t *ServerLiveMigrate) getPingOutput() (string, error) {
	if t.clientPingPid == 0 {
		logging.Warning("[%s] ping pid is not exists", t.ServerId())
		return "", nil
	}
	clientGuest, err := t.getClientGuest()
	if err != nil {
		return "", err
	}
	stdout, stderr := clientGuest.GetExecStatusOutput(t.clientPingPid)
	if stdout != "" {
		return stdout, nil
	}
	logging.Debug("[%s] ping output:\n%s", t.ServerId(), stdout)
	return "", fmt.Errorf("get ping output failed: %s", stderr)
}
func (t *ServerLiveMigrate) startLiveMigrate() error {
	err := t.Client.NovaV2().Servers().LiveMigrate(t.Server.Id, "auto", "")
	if err != nil {
		return err
	}
	logging.Info("[%s] live migrating", t.Server.Id)
	return t.WaitServerTaskFinished(true)
}
func (t ServerLiveMigrate) confirmServerHasIpAddress() error {
	serverHost, err := t.Client.NovaV2().Hypervisors().Found(t.Server.Host)
	if err != nil {
		return err
	}
	serverGuest := &guest.Guest{Connection: serverHost.HostIp, Domain: t.ServerId()}
	serverGuest.Connect()
	logging.Info("[%s] connecting to qga ...", t.ServerId())
	serverGuest.ConnectToQGA(common.CONF.Test.QGAChecker.QgaConnectTimeout)
	err = utility.RetryWithErrors(
		utility.RetryCondition{Timeout: time.Second * 60, IntervalMin: time.Second * 2},
		[]string{"GuestNoIpaddressError"},
		func() error {
			ipaddrs := serverGuest.GetIpaddrs()
			if len(ipaddrs) == 0 {
				return utility.NewGuestNoIpaddressError()
			}
			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("server has no ipaddress")
	}
	return nil
}

func (t *ServerLiveMigrate) Start() error {
	if !common.CONF.Test.QGAChecker.Enabled &&
		common.CONF.Test.LiveMigrate.PingEnabled {
		logging.Warning("[%s] qga checker is disabled but ping is enabled, ping check will be ignored", t.ServerId())
	}
	t.RefreshServer()
	if !t.Server.IsActive() {
		return fmt.Errorf("server is not active")
	}

	if common.CONF.Test.LiveMigrate.PingEnabled {
		interfaces, err := t.Client.NovaV2().Servers().ListInterfaces(t.ServerId())
		if err != nil {
			return err
		}
		if len(interfaces) == 0 {
			logging.Warning("[%s] server has no interface, skip to run ping process", t.ServerId())
			return nil
		}
		// 先检查实例是否有IP
		if err := t.confirmServerHasIpAddress(); err != nil {
			return err
		}
		// 创建客户端
		err = t.createClientServer()
		if err != nil {
			return err
		}
		// 开始运行ping
		// TODO: 判断 IPv4 还是 IPv6
		err = t.startPing(interfaces[0].GetIPAddresses()[0])
		if err != nil {
			return fmt.Errorf("start ping process failed: %s", err)
		}
	}
	sourceHost := t.Server.Host
	logging.Info("[%s] source host is %s", t.Server.Id, sourceHost)
	startTime := time.Now()
	if err := t.startLiveMigrate(); err != nil {
		return err
	}
	if err := t.confirmLiveMigrated(sourceHost); err != nil {
		return err
	}
	logging.Info("[%s] migrated, %s -> %s, used: %v",
		t.Server.Id, sourceHost, t.Server.Host, time.Since(startTime))
	if err := t.confirmPingResult(); err != nil {
		return err
	}
	return nil
}
func (t *ServerLiveMigrate) confirmLiveMigrated(sourceHost string) error {
	if t.Server.IsError() {
		return fmt.Errorf("server is error")
	}
	if !t.Server.IsActive() {
		return fmt.Errorf("server is not active")
	}
	if t.Server.Host == sourceHost {
		return fmt.Errorf("server host not changed")
	}
	return nil
}
func (t *ServerLiveMigrate) confirmPingResult() error {
	if !common.CONF.Test.QGAChecker.Enabled ||
		!common.CONF.Test.LiveMigrate.PingEnabled {
		return nil
	}
	if t.clientPingPid == 0 {
		logging.Warning("[%s] ping pid is not exists", t.ServerId())
		return nil
	}
	if err := t.stopPing(); err != nil {
		return fmt.Errorf("stop ping process failed: %s", err)
	}
	if stdout, err := t.getPingOutput(); err != nil {
		return err
	} else {
		matchedResult := utility.MatchPingResult(stdout)
		if len(matchedResult) == 0 {
			return fmt.Errorf("ping result not matched")
		}
		logging.Info("[%s] ping result: %s", t.ServerId(), matchedResult[0])
		transmitted, _ := strconv.Atoi(matchedResult[1])
		received, _ := strconv.Atoi(matchedResult[2])
		if transmitted-received > common.CONF.Test.LiveMigrate.MaxLoss {
			return fmt.Errorf("ping loss %d package(s)", transmitted-received)
		}
	}
	return nil
}
func (t ServerLiveMigrate) Cleanup() {
	if t.clientServer != nil {
		logging.Info("[%s] deleting client server %s", t.ServerId(), t.clientServer.Id)
		t.Client.NovaV2().Servers().Delete(t.clientServer.Id)
		t.Client.NovaV2().Servers().WaitDeleted(t.clientServer.Id)
	}
}

type ServerMigrate struct {
	ServerActionTest
	EmptyCleanup
}

func (t ServerMigrate) Start() error {
	t.RefreshServer()
	if !t.Server.IsActive() {
		return fmt.Errorf("server is not active")
	}

	sourceHost := t.Server.Host
	startTime := time.Now()
	logging.Info("[%s] source host is %s", t.Server.Id, sourceHost)

	err := t.Client.NovaV2().Servers().Migrate(t.Server.Id, "")
	if err != nil {
		return err
	}
	logging.Info("[%s] migrating", t.Server.Id)

	if err := t.WaitServerTaskFinished(true); err != nil {
		return err
	}
	if t.Server.IsError() {
		return fmt.Errorf("server is error")
	}
	if t.Server.Host == sourceHost {
		return fmt.Errorf("server host not changed")
	}
	logging.Info("[%s] migrated, %s -> %s, used: %v",
		t.Server.Id, sourceHost, t.Server.Host, time.Since(startTime))
	return nil
}
