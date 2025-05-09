package internal

import (
	"fmt"
	"regexp"
	"strconv"
	"syscall"
	"time"

	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/guest"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/utility"
)

type ServerLiveMigrate struct {
	ServerActionTest
	EmptyCleanup
	clientServer  *nova.Server
	clientGuest   *guest.Guest
	clientPingPid int
	enablePing    bool
}

func (t *ServerLiveMigrate) createClientServer() error {
	clientServerOpt := t.getServerBootOption(fmt.Sprintf("client-%s", t.Server.Name))
	clientServerOpt.AvailabilityZone = t.Server.AZ
	clientServer, err := t.Client.NovaV2().CreateServer(clientServerOpt)
	if err != nil {
		return fmt.Errorf("create client instance failed: %s", err)
	}
	console.Info("[%s] creating client server", t.ServerId())
	t.clientServer, err = t.Client.NovaV2().WaitServerBooted(clientServer.Id)
	if err != nil {
		return err
	}
	console.Info("[%s] client server (%s) created, host: %s", t.ServerId(),
		t.clientServer.Name, t.clientServer.Host)
	return nil
}
func (t *ServerLiveMigrate) waitServerBooted(serverId string) error {
	return utility.RetryWithErrors(
		utility.RetryCondition{
			Timeout:      time.Minute * 10,
			IntervalMin:  time.Second * 2,
			IntervalStep: time.Second,
			IntervalMax:  time.Second * 10,
		},
		[]string{"ServerNotBooted"},
		func() error {
			consoleLog, err := t.Client.NovaV2().GetServerConsoleLog(serverId, 50)
			if err != nil {
				return fmt.Errorf("get console log failed: %s", err)
			}
			// TODO: set key world by config file.
			reg := regexp.MustCompile(` login:`)
			result := reg.FindStringSubmatch(consoleLog.Output)
			console.Debug("[%s] client console log: %s", t.ServerId(), consoleLog.Output)
			if len(result) > 0 {
				return nil
			}
			return utility.NewServerNotBootedError(serverId)
		},
	)
}

func (t *ServerLiveMigrate) getClientGuest() (*guest.Guest, error) {
	if t.clientGuest == nil {
		host, err := t.Client.NovaV2().FindHypervisor(t.clientServer.Host)
		if err != nil {
			return nil, err
		}
		t.clientGuest = &guest.Guest{Connection: host.HostIp, Domain: t.clientServer.Id}
		if err := t.clientGuest.Connect(); err != nil {
			return nil, fmt.Errorf("connect guest %s faield: %s", t.clientGuest, err)
		}
		console.Info("[%s] connecting to qga ...", t.clientGuest.Domain)
		t.clientGuest.ConnectToQGA(t.Config.QGAChecker.QgaConnectTimeout)
	}
	return t.clientGuest, nil
}
func (t *ServerLiveMigrate) startPing(targetIp string) error {
	if !t.Config.QGAChecker.Enabled ||
		!t.Config.LiveMigrate.PingEnabled {
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
	console.Info("[%s] client ping -> %s", t.ServerId(), targetIp)
	result := clientGuest.Ping(targetIp, t.Config.LiveMigrate.PingInterval, 0, ipaddrs[0], false)
	if result.ErrData != "" {
		return fmt.Errorf("run ping to %s failed: %s", targetIp, result.ErrData)
	}
	t.clientPingPid = result.Pid
	console.Debug("[%s] ping process pid is: %d", t.ServerId(), t.clientPingPid)
	return nil
}
func (t *ServerLiveMigrate) stopPing() error {
	if !t.Config.QGAChecker.Enabled ||
		!t.Config.LiveMigrate.PingEnabled {
		return nil
	}
	if t.clientPingPid == 0 {
		console.Warn("[%s] ping pid is not exists", t.ServerId())
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
		console.Warn("[%s] ping pid is not exists", t.ServerId())
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
	console.Debug("[%s] ping output:\n%s", t.ServerId(), stdout)
	return "", fmt.Errorf("get ping output failed: %s", stderr)
}
func (t *ServerLiveMigrate) checkPingBeforeMigrate(targetIp string) error {
	defer func() {
		t.clientPingPid = 0
	}()
	console.Info("[%s] confirm ping packages not loss ...", t.ServerId())
	return utility.RetryWithErrors(
		utility.RetryCondition{
			Timeout:     time.Minute * 5,
			IntervalMin: time.Second * 4,
		},
		[]string{"PingLossPackage"},
		func() error {
			if err := t.startPing(targetIp); err != nil {
				return err
			}
			time.Sleep(time.Second * 30)
			if err := t.stopPing(); err != nil {
				return fmt.Errorf("stop ping process failed: %s", err)
			}
			stdout, err := t.getPingOutput()
			if err != nil {
				return err
			}
			matchedResult := utility.MatchPingResult(stdout)
			if len(matchedResult) == 0 {
				return fmt.Errorf("ping result not matched")
			}
			console.Info("[%s] ping result: %s", t.ServerId(), matchedResult[0])
			transmitted, _ := strconv.Atoi(matchedResult[1])
			received, _ := strconv.Atoi(matchedResult[2])
			if transmitted-received > 0 {
				return utility.NewPingLossPackage(transmitted - received)
			}
			return nil
		},
	)
}

func (t *ServerLiveMigrate) startLiveMigrate() error {
	err := t.Client.NovaV2().ServerLiveMigrate(t.Server.Id, "auto", "")
	if err != nil {
		return err
	}
	console.Info("[%s] live migrating", t.Server.Id)
	return t.WaitServerTaskFinished(true)
}
func (t ServerLiveMigrate) confirmServerHasIpAddress() error {
	serverHost, err := t.Client.NovaV2().FindHypervisor(t.Server.Host)
	if err != nil {
		return err
	}
	serverGuest := &guest.Guest{Connection: serverHost.HostIp, Domain: t.ServerId()}
	console.Info("[%s] connecting to guest ...", t.ServerId())
	serverGuest.Connect()
	console.Info("[%s] connecting to qga ...", t.ServerId())
	serverGuest.ConnectToQGA(t.Config.QGAChecker.QgaConnectTimeout)
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
func (t ServerLiveMigrate) Skip() (bool, string) {
	if !t.Server.IsActive() {
		return true, "server is not active"
	}
	return false, ""
}
func (t *ServerLiveMigrate) Start() error {
	if t.Config.QGAChecker.Enabled && t.Config.LiveMigrate.PingEnabled {
		t.enablePing = true
	} else if t.Config.LiveMigrate.PingEnabled {
		console.Warn("[%s] disable ping check because qga checker is disabled", t.ServerId())
	}
	if t.enablePing {
		interfaces, err := t.Client.NovaV2().ListServerInterfaces(t.ServerId())
		if err != nil {
			return err
		}
		if len(interfaces) == 0 {
			console.Warn("[%s] server has no interface, skip to run ping process", t.ServerId())
			return nil
		}
		// 检查实例是否有IP
		if err := t.confirmServerHasIpAddress(); err != nil {
			return err
		}
		// 创建客户端
		err = t.createClientServer()
		if err != nil {
			return err
		}
		console.Info("[%s] waiting client booted", t.ServerId())
		if err := t.waitServerBooted(t.clientServer.Id); err != nil {
			return err
		}
		// 检测ping 是否丢包
		err = t.checkPingBeforeMigrate(interfaces[0].GetIpAddresses()[0])
		if err != nil {
			return fmt.Errorf("ping check failed: %s", err)
		}
		console.Info("[%s] ping package not loss", t.ServerId())
		// 开始运行ping
		// TODO: 判断 IPv4 还是 IPv6
		err = t.startPing(interfaces[0].GetIpAddresses()[0])
		if err != nil {
			return fmt.Errorf("start ping process failed: %s", err)
		}
	}
	sourceHost := t.Server.Host
	console.Info("[%s] source host is %s", t.Server.Id, sourceHost)
	startTime := time.Now()
	if err := t.startLiveMigrate(); err != nil {
		return err
	}
	if err := t.confirmLiveMigrated(sourceHost); err != nil {
		console.Error("[%s] migrate failed: %s", t.ServerId(), err)
		return err
	}
	console.Info("[%s] migrated, %s -> %s, used: %v",
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
	if !t.enablePing {
		return nil
	}
	if t.clientPingPid == 0 {
		console.Warn("[%s] ping pid is not exists", t.ServerId())
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
		console.Info("[%s] ping result: %s", t.ServerId(), matchedResult[0])
		transmitted, _ := strconv.Atoi(matchedResult[1])
		received, _ := strconv.Atoi(matchedResult[2])
		if transmitted-received > t.Config.LiveMigrate.MaxLoss {
			return fmt.Errorf("ping loss %d package(s)", transmitted-received)
		}
	}
	return nil
}
func (t ServerLiveMigrate) TearDown() error {
	if t.clientServer != nil {
		console.Info("[%s] deleting client server %s", t.ServerId(), t.clientServer.Id)
		if err := t.Client.NovaV2().DeleteServer(t.clientServer.Id); err != nil {
			return err
		}
		if err := t.Client.NovaV2().WaitServerDeleted(t.clientServer.Id); err != nil {
			return err
		}
	}
	return nil
}

type ServerMigrate struct {
	ServerActionTest
	EmptyCleanup
}

func (t ServerMigrate) Start() error {
	if !t.Server.IsActive() && !t.Server.IsStopped() {
		return fmt.Errorf("server status is not active or stopped")
	}

	sourceHost := t.Server.Host
	startTime := time.Now()
	console.Info("[%s] source host is %s", t.Server.Id, sourceHost)

	err := t.Client.NovaV2().ServerMigrate(t.Server.Id, "")
	if err != nil {
		return err
	}
	console.Info("[%s] migrating", t.Server.Id)

	if err := t.WaitServerTaskFinished(false); err != nil {
		return err
	}
	if t.Server.IsError() {
		return fmt.Errorf("server is error")
	}
	if t.Server.Host == sourceHost {
		return fmt.Errorf("server host not changed")
	}
	console.Info("[%s] migrated, %s -> %s, used: %v",
		t.Server.Id, sourceHost, t.Server.Host, time.Since(startTime))
	return nil
}
