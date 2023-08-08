package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/gitutils"
	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/stackcrud/cmd/commands"
	"github.com/BytemanD/stackcrud/common"
)

var (
	Version string
)

func getVersion() string {
	if Version == "" {
		return gitutils.GetVersion()
	}
	return fmt.Sprint(Version)
}

func main() {
	rootCmd := cobra.Command{
		Use:     "stackcurd",
		Short:   "Golang Openstack Client",
		Version: getVersion(),
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			debug, _ := cmd.Flags().GetBool("debug")
			conf, _ := cmd.Flags().GetStringArray("conf")
			level := logging.INFO
			if debug {
				level = logging.DEBUG
			}
			logging.BasicConfig(logging.LogConfig{Level: level})
			if err := common.LoadConf(conf); err != nil {
				logging.Error("load config faield, %v", err)
			}
			if !debug && common.CONF.Debug {
				logging.BasicConfig(logging.LogConfig{Level: logging.DEBUG})
			}
		},
	}

	rootCmd.PersistentFlags().BoolP("debug", "d", false, "显示Debug信息")
	rootCmd.PersistentFlags().StringArrayP("conf", "c", common.CONF_FILES, "配置文件")

	rootCmd.AddCommand(commands.Server)

	rootCmd.Execute()
}
