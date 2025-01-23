package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/BytemanD/easygo/pkg/global/gitutils"
	"github.com/BytemanD/go-console/console"

	"github.com/BytemanD/skyman/cmd/context"
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
	return fmt.Sprint(Version)
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

		client := openstack.DefaultClient()

		fmt.Println("Servers:")

		identityVerion, err := client.KeystoneV3().GetStableVersion()
		utility.LogError(err, "get idendity veresion failed", true)
		if err != nil {
			return
		}
		fmt.Printf("  %-11s: %s\n", "Keystone", identityVerion.VersoinInfo())

		errors := 0
		computeVerion, err := client.NovaV2().GetCurrentVersion()
		if err == nil {
			fmt.Printf("  %-11s: %s\n", "Nova", computeVerion.VersoinInfo())
		} else {
			fmt.Printf("  %-11s: Unknown (%s)\n", "Nova", err)
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

func main() {
	rootCmd := cobra.Command{
		Use:     "skyman",
		Short:   "Golang OpenStack Client \n" + LOGO,
		Version: getVersion(),
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			format, _ := cmd.Flags().GetString("format")
			if format != "" && !slice.Contain(common.GetOutputFormats(), format) {
				fmt.Printf("invalid foramt '%s'\n", format)
				os.Exit(1)
			}
			conf, _ := cmd.Flags().GetString("conf")
			if conf == "" {
				ctxConf, err := context.LoadContextConf()
				if err != nil {
					console.Debug("load context failed: %s", err)
				} else {
					if ctx := ctxConf.GetCurrent(); ctx != nil {
						console.Debug("use conf from context")
						conf = ctx.Conf
					}
				}
			}

			if err := common.LoadConfig(conf); err != nil {
				fmt.Printf("load config failed: %v\n", err)
				os.Exit(1)
			}
			if common.CONF.Debug {
				console.EnableLogDebug()
			}
			if common.CONF.LogFile != "" {
				console.SetLogFile(common.CONF.LogFile)
			}
			console.Debug("load config file from %s", viper.ConfigFileUsed())
			computeApiVersion, _ := cmd.Flags().GetString("compute-api-version")
			openstack.COMPUTE_API_VERSION = computeApiVersion
		},
	}

	rootCmd.PersistentFlags().BoolP("debug", "d", false, i18n.T("showDebug"))
	rootCmd.PersistentFlags().String("log-file", "", i18n.T("logFile"))
	rootCmd.PersistentFlags().Bool("log-color", false, i18n.T("enableLogColor"))
	rootCmd.PersistentFlags().StringP("conf", "c", os.Getenv("SKYMAN_CONF_FILE"),
		i18n.T("thePathOfConfigFile"))
	rootCmd.PersistentFlags().StringP("format", "f", "table",
		fmt.Sprintf(i18n.T("formatAndSupported"), common.GetOutputFormats()))

	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("format", rootCmd.PersistentFlags().Lookup("format"))
	viper.BindPFlag("logFile", rootCmd.PersistentFlags().Lookup("log-file"))
	viper.BindPFlag("enableLogColor", rootCmd.PersistentFlags().Lookup("log-color"))

	rootCmd.PersistentFlags().String("compute-api-version", "", "Compute API version")

	TestCmd.AddCommand(
		test.TestFio, test.ServerPing, test.TestNetQos, test.TestServerAction,
	)

	rootCmd.AddCommand(
		versionCmd, context.ContextCmd,

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
