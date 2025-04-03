package main

import (
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/BytemanD/easygo/pkg/global/gitutils"
	"github.com/BytemanD/go-console/console"

	"github.com/BytemanD/skyman/cmd/cloud"
	"github.com/BytemanD/skyman/cmd/neutron"

	"github.com/BytemanD/skyman/cmd/benchmark"
	"github.com/BytemanD/skyman/cmd/cinder"
	"github.com/BytemanD/skyman/cmd/glance"
	"github.com/BytemanD/skyman/cmd/keystone"

	"github.com/BytemanD/skyman/cmd/nova"
	"github.com/BytemanD/skyman/cmd/quota"
	"github.com/BytemanD/skyman/cmd/templates"
	"github.com/BytemanD/skyman/cmd/test"
	"github.com/BytemanD/skyman/cmd/tool"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/common/i18n"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/utility"
)

var (
	Version       string
	GoVersion     string
	BuildDate     string
	BuildPlatform string
)

var LOGO = `
      _                             
  ___| |  _ _   _ ____  _____ ____  
 /___) |_/ ) | | |    \(____ |  _ \ 
|___ |  _ (| |_| | | | / ___ | | | |
(___/|_| \_)\__  |_|_|_\_____|_| |_|
           (____/
`

func getVersion() string {
	if Version == "" {
		return gitutils.GetVersion()
	}
	return Version
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Version of client and server",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		if GoVersion == "" {
			GoVersion = runtime.Version()
		}
		println("Client:")
		fmt.Printf("  %-14s: %s\n", "Version", getVersion())
		fmt.Printf("  %-14s: %s\n", "GoVersion", GoVersion)
		fmt.Printf("  %-14s: %s\n", "BuildDate", BuildDate)
		fmt.Printf("  %-14s: %s\n", "BuildPlatform", BuildPlatform)

		client := common.DefaultClient()

		fmt.Println("Servers:")

		identityVerion, err := client.KeystoneV3().GetStableVersion()
		utility.LogError(err, "get idendity veresion failed", true)
		if err != nil {
			return
		}
		fmt.Printf("  %-11s: %s\n", "Keystone", identityVerion.VersoinInfo())

		errors := 0

		versions, err := client.NovaV2().GetApiVersions()
		if err == nil {
			if currentVersion := versions.Current(); currentVersion != nil {
				fmt.Printf("  %-11s: %s\n", "Nova", versions.Current().VersoinInfo())
			} else {
				fmt.Printf("  %-11s: Unknown (%s)\n", "Nova", err)
			}
		} else {
			fmt.Printf("  %-11s: get api versions failed (%s)\n", "Nova", err)
			errors++
		}
		imageVerion, err := client.GlanceV2().GetCurrentVersion()
		if err == nil {
			fmt.Printf("  %-11s: %s\n", "Glance", imageVerion.VersoinInfo())
		} else {
			fmt.Printf("  %-11s: Unknown (%s)\n", "Glance", err)
			errors++
		}
		storageVerion, err := client.CinderV2().GetCurrentVersion()
		if err == nil {
			fmt.Printf("  %-11s: %s\n", "Cinder", storageVerion.VersoinInfo())
		} else {
			fmt.Printf("  %-11s: Unknown (%s)\n", "Cinder", err)
			errors++
		}
		neutronVerion, err := client.NeutronV2().GetCurrentVersion()
		if err == nil {
			fmt.Printf("  %-11s: %s\n", "Neutron", neutronVerion.VersoinInfo())
		} else {
			fmt.Printf("  %-11s: Unknown (%s)\n", "Neutron", err)
			errors++
		}
		if errors > 0 {
			os.Exit(1)
		}
	},
}

var TestCmd = &cobra.Command{Use: "test", Short: "Test tools"}

var ErrFoo = errors.New("foo error")

func main() {
	rootCmd := cobra.Command{
		Use:     "skyman",
		Short:   "Golang OpenStack Client \n" + LOGO,
		Version: getVersion(),
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			conf, _ := cmd.Flags().GetString("conf")
			if err := common.LoadConfig(conf); err != nil {
				console.Error("load config file failed: %s", err)
				os.Exit(1)
			}
			if viper.ConfigFileUsed() == "" {
				console.Debug("配置文件查找失败")
			} else {
				console.Debug("load config file from %s", viper.ConfigFileUsed())
			}
			computeApiVersion, _ := cmd.Flags().GetString("compute-api-version")
			openstack.COMPUTE_API_VERSION = computeApiVersion
		},
	}

	rootCmd.PersistentFlags().BoolP("debug", "d", false, i18n.T("enableDebug"))
	rootCmd.PersistentFlags().String("log-file", "", i18n.T("logFile"))
	rootCmd.PersistentFlags().StringP("format", "f", "table-light",
		fmt.Sprintf(i18n.T("formatAndSupported"), common.GetOutputFormats()))
	rootCmd.PersistentFlags().String("cloud", "", i18n.T("cloudName"))
	rootCmd.PersistentFlags().StringP("conf", "c", os.Getenv("SKYMAN_CONF_FILE"),
		i18n.T("thePathOfConfigFile"))

	viper.AutomaticEnv()
	viper.SetEnvPrefix("SKYMAN")

	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("format", rootCmd.PersistentFlags().Lookup("format"))
	viper.BindPFlag("logFile", rootCmd.PersistentFlags().Lookup("log-file"))
	viper.BindPFlag("cloud", rootCmd.PersistentFlags().Lookup("cloud"))

	rootCmd.PersistentFlags().String("compute-api-version", "", "Compute API version")

	TestCmd.AddCommand(
		test.TestFio, test.ServerPing, test.TestNetQos, test.TestServerAction,
	)

	rootCmd.AddCommand(
		versionCmd, cloud.CloudsCmd,

		keystone.Token,
		keystone.Service, keystone.Endpoint, keystone.Region,
		keystone.User, keystone.Project,

		nova.Server, nova.Flavor, nova.Hypervisor,
		nova.Keypair, nova.Compute, nova.Console,
		nova.Migration, nova.AZ, nova.Aggregate,
		glance.Image,
		cinder.Volume, cinder.Snapshot, cinder.Backup,

		neutron.Router, neutron.Network, neutron.Subnet, neutron.Port,
		neutron.Security, neutron.SG,

		quota.QuotaCmd,
		templates.DefineCmd, templates.UndefineCmd,
		tool.ToolCmd,
		TestCmd,
		benchmark.BenchmarkCmd,
	)
	rootCmd.Execute()
}
