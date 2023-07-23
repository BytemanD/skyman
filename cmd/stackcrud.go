package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/BytemanD/easygo/pkg/global/gitutils"
	"github.com/BytemanD/easygo/pkg/global/logging"

	"github.com/BytemanD/stackcrud/cli/compute"
	"github.com/BytemanD/stackcrud/cli/identity"
	"github.com/BytemanD/stackcrud/cli/image"
	"github.com/BytemanD/stackcrud/cli/storage"
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
			conf, _ := cmd.Flags().GetString("conf")
			if err := common.LoadConfig(conf); err != nil {
				fmt.Printf("load config failed, %v\n", err)
				os.Exit(1)
			}

			if debug || common.CONF.Debug {
				logging.BasicConfig(logging.LogConfig{Level: logging.DEBUG})
			} else {
				logging.BasicConfig(logging.LogConfig{Level: logging.INFO})
			}
			logging.Debug("load config file from %s", viper.ConfigFileUsed())
		},
	}

	rootCmd.PersistentFlags().BoolP("debug", "d", false, "显示Debug信息")
	rootCmd.PersistentFlags().StringP("conf", "c", "", "配置文件")

	rootCmd.AddCommand(identity.Token)

	rootCmd.AddCommand(compute.Server)
	rootCmd.AddCommand(compute.Flavor)
	rootCmd.AddCommand(compute.Hypervisor)
	rootCmd.AddCommand(compute.Keypair)
	rootCmd.AddCommand(compute.Compute)
	rootCmd.AddCommand(compute.Console)

	rootCmd.AddCommand(image.Image)

	rootCmd.AddCommand(storage.Volume)

	rootCmd.Execute()
}
