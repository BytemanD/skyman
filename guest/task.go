package guest

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/BytemanD/easygo/pkg/global/logging"
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

func installIperf(guest Guest, localIperf3File string) error {
	remoteFile, err := guest.CopyFile(localIperf3File, "/tmp")
	if err != nil {
		return fmt.Errorf("%s 拷贝iperf3失败: %s", guest.Domain, err)
	}
	if err := guest.RpmInstall(remoteFile); err != nil {
		return fmt.Errorf("%s 安装iperf3失败: %s", guest.Domain, err)
	}
	return nil
}

// 使用 iperf3 工具测试网络BPS/PPS
//
// 参数为客户端和服务端实例的连接地址, 格式: "连接地址@实例 UUID". e.g.
//
//	192.168.192.168@a6ee919a-4026-4f0b-8d7e-404950a91eb3
type NetQosTest struct {
	ClientGuest     Guest
	ServerGuest     Guest
	PPS             bool
	LocalIperf3File string
	ClientOptions   string
	ServerOptions   string
}

func (t *NetQosTest) Run() (float64, float64, error) {
	logging.Info("连接客户端实例: %s", t.ClientGuest)
	err := t.ClientGuest.Connect()
	if t.ClientGuest.IsSame(t.ServerGuest) {
		logging.Error("客户端和服务端实例不能相同")
		return 0, 0, err
	}
	if err != nil {
		logging.Error("连接客户端实例失败, %s", err)
		return 0, 0, err
	}
	logging.Info("连接服务端实例: %s", t.ServerGuest)
	err = t.ServerGuest.Connect()
	if err != nil {
		logging.Error("连接服务端实例失败, %s", err)
		return 0, 0, err
	}

	logging.Info("获取客户端和服务端实例IP地址")
	clientAddresses := t.ClientGuest.GetIpaddrs()
	serverAddresses := t.ServerGuest.GetIpaddrs()
	logging.Info("客户端实例IP地址: %s", clientAddresses)
	logging.Info("服务端实例IP地址: %s", serverAddresses)

	if len(clientAddresses) == 0 || len(serverAddresses) == 0 {
		logging.Fatal("客户端和服务端实例必须至少有一张启用的网卡")
	}

	if !t.ServerGuest.HasCommand("iperf3") {
		if t.LocalIperf3File == "" {
			return 0, 0, fmt.Errorf("iperf3 is not installed in server guest")
		}
		logging.Info("拷贝安装包 -> 服务端")
		if err := installIperf(t.ServerGuest, t.LocalIperf3File); err != nil {
			return 0, 0, fmt.Errorf("服务端端安装iperf3失败: %s", err)
		}
	}
	if !t.ClientGuest.HasCommand("iperf3") {
		if t.LocalIperf3File == "" {
			return 0, 0, fmt.Errorf("iperf3 is not installed in client guest")
		}
		logging.Info("拷贝安装包 -> 客户端")
		if err := installIperf(t.ClientGuest, t.LocalIperf3File); err != nil {
			return 0, 0, fmt.Errorf("客户端安装iperf3失败: %s", err)
		}
	}
	clientOptions := strings.Split(t.ClientOptions, " ")
	times := 10
	for i, option := range strings.Split(t.ClientOptions, " ") {
		if option == "-t" || option == "--time" {
			times, _ = strconv.Atoi(clientOptions[i+1])
			break
		}
	}
	if t.PPS {
		splitOptions := strings.Split(t.ClientOptions, " ")
		splitOptions = append(splitOptions, "-l", "16")
		t.ClientOptions = strings.Join(splitOptions, " ")
	}

	fomatTime := time.Now().Format("20060102_150405")
	serverPids := []int{}
	for _, serverAddress := range serverAddresses {
		logfile := fmt.Sprintf("/tmp/iperf3_s_%s_%s", fomatTime, serverAddress)
		logging.Info("启动服务端: %s", serverAddress)
		execResult := t.ServerGuest.RunIperfServer(
			serverAddress, logfile, t.ServerOptions)
		if execResult.Failed {
			return 0, 0, err
		}
		serverPids = append(serverPids, execResult.Pid)
	}
	if len(serverPids) > 0 {
		defer t.ServerGuest.Kill(9, serverPids)
	}

	jobs := []Job{}
	for i := 0; i < len(clientAddresses) && i < len(serverAddresses); i++ {
		logfile := fmt.Sprintf("/tmp/iperf3_c_%s_%s", fomatTime, serverAddresses[i])
		logging.Info("启动客户端: %s -> %s", clientAddresses[i], serverAddresses[i])
		var execResult ExecResult
		if !t.PPS {
			execResult = t.ClientGuest.RunIperfClient(
				clientAddresses[i], serverAddresses[i], logfile,
				t.ClientOptions)
		} else {
			execResult = t.ClientGuest.RunIperfClientUdp(
				clientAddresses[i], serverAddresses[i], logfile,
				t.ClientOptions)
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
		t.ClientGuest.GetExecStatusOutput(job.Pid)
	}
	logging.Info("测试结束")

	reports := NewIperfReports()
	for _, job := range jobs {
		execResult := t.ClientGuest.Cat(job.Logfile)
		reports.Add(job.SourceIp, job.DestIp, execResult.OutData)
	}
	if !t.PPS {
		reports.PrintBps()
	} else {
		reports.PrintPps()
	}
	return reports.SendTotal.Value, reports.ReceiveTotal.Value, nil
}

type FioTest struct {
	Guest   Guest
	Options FioOptions
}

func (t *FioTest) Run() error {
	logging.Debug("连接实例: %s", t.Guest)
	if err := t.Guest.Connect(); err != nil {
		return fmt.Errorf("连接实例失败, %s", err)
	}

	result, err := t.Guest.RunFio(t.Options)
	if err != nil {
		return err
	}
	fmt.Println(result)
	return nil
}
