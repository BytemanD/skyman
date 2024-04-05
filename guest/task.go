package guest

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/BytemanD/easygo/pkg/global/logging"

	"github.com/BytemanD/skyman/common"
)

type GuestConnection struct {
	Connection string
	Domain     string
}

type Job struct {
	SourceIp string
	DestIp   string
	Pid      int
	Logfile  string
	Output   string
	Sender   Bandwidth
	Receiver Bandwidth
}

func installIperf(guest Guest) error {
	if common.CONF.Iperf.GuestPath != "" {
		logging.Info("%s 安装 iperf3, 文件路径: %s", guest.Domain,
			common.CONF.Iperf.GuestPath)
		if err := guest.RpmInstall(common.CONF.Iperf.GuestPath); err != nil {
			return err
		}
	} else if common.CONF.Iperf.LocalPath == "" {
		return fmt.Errorf("iperf3 文件路径未配置")
	}
	logging.Info("[%s] 安装iperf3, 使用本地文件: %s", guest.Domain,
		common.CONF.Iperf.LocalPath)
	remoteFile, err := guest.CopyFile(common.CONF.Iperf.LocalPath, "/tmp")
	if err != nil {
		return err
	}
	if err := guest.RpmInstall(*remoteFile); err != nil {
		return err
	}
	return nil
}

// 使用 iperf3 工具测试网络BPS/PPS
//
// 参数为客户端和服务端实例的连接地址，格式: "连接地址:实例 UUID"。例如：
//
//	192.168.192.168:a6ee919a-4026-4f0b-8d7e-404950a91eb3
func TestNetQos(clientConn GuestConnection, serverConn GuestConnection,
	pps bool) (float64, float64, error) {
	clientGuest := Guest{
		Connection: clientConn.Connection,
		Domain:     clientConn.Domain,
		QGATimeout: 60,
		ByUUID:     true}
	serverGuest := Guest{
		Connection: serverConn.Connection,
		QGATimeout: 60,
		Domain:     serverConn.Domain,
		ByUUID:     true}
	err := clientGuest.Connect()
	if clientGuest.IsSame(serverGuest) {
		logging.Error("客户端和服务端实例不能相同")
		return 0, 0, err
	}
	logging.Info("连接客户端实例: %s", clientGuest.Domain)
	if err != nil {
		logging.Error("连接客户端实例失败, %s", err)
		return 0, 0, err
	}
	logging.Info("连接服务端实例: %s", serverGuest.Domain)
	err = serverGuest.Connect()
	if err != nil {
		logging.Error("连接服务端实例失败, %s", err)
		return 0, 0, err
	}

	logging.Info("获取客户端和服务端实例IP地址")
	clientAddresses := clientGuest.GetIpaddrs()
	serverAddresses := serverGuest.GetIpaddrs()
	logging.Info("客户端实例IP地址: %s", clientAddresses)
	logging.Info("服务端实例IP地址: %s", serverAddresses)

	if len(clientAddresses) == 0 || len(serverAddresses) == 0 {
		logging.Fatal("客户端和服务端实例必须至少有一张启用的网卡")
	}
	if !clientGuest.HasCommand("iperf3") {
		if err := installIperf(clientGuest); err != nil {
			logging.Fatal("安装iperf失败, %s", err)
		}
	}
	if !serverGuest.HasCommand("iperf3") {
		if err := installIperf(serverGuest); err != nil {
			logging.Fatal("安装iperf失败, %s", err)
		}
	}
	clientOptions := strings.Split(common.CONF.Iperf.ClientOptions, " ")
	times := 10
	for i, option := range strings.Split(common.CONF.Iperf.ClientOptions, " ") {
		if option == "-t" || option == "--time" {
			times, _ = strconv.Atoi(clientOptions[i+1])
			break
		}
	}
	if pps {
		common.CONF.Iperf.ClientOptions += "-l 16"
	}

	fomatTime := time.Now().Format(time.RFC3339)
	serverPids := []int{}
	for _, serverAddress := range serverAddresses {
		logfile := fmt.Sprintf("/tmp/iperf3_s_%s_%s", fomatTime, serverAddress)
		logging.Info("启动服务端: %s", serverAddress)
		execResult := serverGuest.RunIperfServer(
			serverAddress, logfile, common.CONF.Iperf.ServerOptions)
		if execResult.Failed {
			return 0, 0, err
		}
		serverPids = append(serverPids, execResult.Pid)
	}
	if len(serverPids) > 0 {
		defer serverGuest.Kill(9, serverPids)
	}

	jobs := []Job{}
	for i := 0; i < len(clientAddresses) && i < len(serverAddresses); i++ {
		logfile := fmt.Sprintf("/tmp/iperf3_c_%s_%s", fomatTime, serverAddresses[i])
		logging.Info("启动客户端: %s -> %s", clientAddresses[i], serverAddresses[i])
		var execResult ExecResult
		if !pps {
			execResult = clientGuest.RunIperfClient(
				clientAddresses[i], serverAddresses[i], logfile,
				common.CONF.Iperf.ClientOptions)
		} else {
			execResult = clientGuest.RunIperfClientUdp(
				clientAddresses[i], serverAddresses[i], logfile,
				common.CONF.Iperf.ClientOptions)
		}
		jobs = append(jobs, Job{
			SourceIp: clientAddresses[i],
			DestIp:   serverAddresses[i],
			Pid:      execResult.Pid,
			Logfile:  logfile,
		})
	}

	logging.Info("等待测试结束(%ds) ...", times)
	time.Sleep(time.Second * time.Duration(times))
	for _, job := range jobs {
		clientGuest.getExecStatusOutput(job.Pid)
	}
	logging.Info("测试结束")

	reports := NewIperfReports()
	for _, job := range jobs {
		execResult := clientGuest.Cat(job.Logfile)
		reports.Add(job.SourceIp, job.DestIp, execResult.OutData)
	}
	if !pps {
		reports.PrintBps()
	} else {
		reports.PrintPps()
	}
	return reports.SendTotal.Value, reports.ReceiveTotal.Value, nil
}
