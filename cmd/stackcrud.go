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
	"github.com/BytemanD/stackcrud/cli/networking"
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
			conf, _ := cmd.Flags().GetString("conf")
			if err := common.LoadConfig(conf); err != nil {
				fmt.Printf("load config failed, %v\n", err)
				os.Exit(1)
			}

			if common.CONF.Debug {
				logging.BasicConfig(logging.LogConfig{Level: logging.DEBUG})
			} else {
				logging.BasicConfig(logging.LogConfig{Level: logging.INFO})
			}
			logging.Debug("load config file from %s", viper.ConfigFileUsed())
		},
	}

	rootCmd.PersistentFlags().BoolP("debug", "d", false, "显示Debug信息")
	rootCmd.PersistentFlags().StringP("conf", "c", "", "配置文件")
	rootCmd.PersistentFlags().String("format", "default",
		fmt.Sprintf("格式, 支持的格式: %v", common.GetOutputFormats()))

	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("format", rootCmd.PersistentFlags().Lookup("format"))

	rootCmd.AddCommand(identity.Token)

	rootCmd.AddCommand(
		identity.Service, identity.Endpoint,
		identity.User, identity.Project,
		compute.Server, compute.Flavor, compute.Hypervisor,
		compute.Keypair, compute.Compute, compute.Console,
		compute.Migration, compute.AZ, compute.Aggregate,
		image.Image,
		storage.Volume,
		networking.Router, networking.Network, networking.Port,
	)

	rootCmd.Execute()
}
