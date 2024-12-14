package test

import (
	"fmt"
	"strings"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/easygo/pkg/stringutils"
	"github.com/BytemanD/skyman/common/i18n"
	"github.com/BytemanD/skyman/guest"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/utility"
	"github.com/spf13/cobra"
)

var VALID_TEST_TYPES = []string{
	"iops", "bandwidth", "latency",
}

const SERVER_FIO_TEST_EXAMPLE = `
1. 测试IOPS:
  skyman test server-fio <server> iops --disk-path /dev/vdb1
`

var TestFio = &cobra.Command{
	Use:     fmt.Sprintf("server-fio <server> <%s>", strings.Join(VALID_TEST_TYPES, "|")),
	Short:   i18n.T("testServerDiskIO"),
	Long:    "基于fio工具测试实例磁盘IO",
	Example: strings.TrimRight(SERVER_FIO_TEST_EXAMPLE, "\n"),
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(2)(cmd, args); err != nil {
			return err
		}
		testType := args[1]
		if testType == "" || !stringutils.ContainsString(VALID_TEST_TYPES, testType) {
			return fmt.Errorf("invalid test type: %s, must be one of %s", testType, VALID_TEST_TYPES)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		idOrName, testType := args[0], args[1]
		numjobs, _ := cmd.Flags().GetInt("numjobs")
		runtime, _ := cmd.Flags().GetInt("runtime")
		filename, _ := cmd.Flags().GetString("filename")
		size, _ := cmd.Flags().GetString("size")

		openstackClient := openstack.DefaultClient()
		server, err := openstackClient.NovaV2().Server().Found(idOrName)
		utility.LogError(err, "get server failed", true)

		if !server.IsActive() {
			logging.Fatal("instance %s is not active", server.Id)
		}

		logging.Info("get server host and client host")
		serverHost, err := openstackClient.NovaV2().Hypervisor().Found(server.Host)
		utility.LogError(err, "get server host failed", true)

		serverConn := guest.Guest{Connection: serverHost.HostIp, Domain: server.Id}
		logging.Info("start test with QGA")
		testOptsList := []guest.FioOptions{}

		switch testType {
		case "iops":
			testOptsList = append(testOptsList,
				guest.FioOptions{
					Numjobs: numjobs, Runtime: runtime, Size: size, FileName: filename,
					RW: "randwrite", IODepth: 128, BS: "4K"},
				guest.FioOptions{
					Numjobs: numjobs, Runtime: runtime, Size: size, FileName: filename,
					RW: "randread", IODepth: 128, BS: "4K"},
			)
		case "bandwidth":
			testOptsList = append(testOptsList,
				guest.FioOptions{
					Numjobs: numjobs, Runtime: runtime, Size: size, FileName: filename,
					RW: "write", IODepth: 64, BS: "1024k"},
				guest.FioOptions{
					Numjobs: numjobs, Runtime: runtime, Size: size, FileName: filename,
					RW: "read", IODepth: 64, BS: "1024k"})
		case "latency":
			testOptsList = append(testOptsList,
				guest.FioOptions{
					Numjobs: numjobs, Runtime: runtime, Size: size, FileName: filename,
					RW: "randwrite", IODepth: 1, BS: "4k"},
				guest.FioOptions{
					Numjobs: numjobs, Runtime: runtime, Size: size, FileName: filename,
					RW: "randread", IODepth: 1, BS: "4k"},
			)
		}
		for _, opts := range testOptsList {
			logging.Info("============= Test %s %s =================", testType, opts.RW)
			job := guest.FioTest{
				Guest: serverConn, Options: opts,
			}
			err = job.Run()
			utility.LogIfError(err, true, "test failed")
		}
	},
}

func init() {
	TestFio.Flags().String("filename", "", "Filename")
	TestFio.Flags().String("size", "1G", "Size")
	TestFio.Flags().Int("runtime", 10, "Runtime")
	TestFio.Flags().Int("numjobs", 1, "Num jobs")
}
