package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/easygo/pkg/stringutils"
	"github.com/BytemanD/easygo/pkg/syscmd"

	"github.com/BytemanD/skyman/cli/compute"
	"github.com/BytemanD/skyman/cli/identity"
	"github.com/BytemanD/skyman/cli/image"
	"github.com/BytemanD/skyman/cli/networking"
	"github.com/BytemanD/skyman/cli/storage"
	"github.com/BytemanD/skyman/cli/templates"
	"github.com/BytemanD/skyman/cli/test"
	"github.com/BytemanD/skyman/cli/tool"
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
	describeTag, err := syscmd.GetOutput("git", "describe", "--tags", "--abbrev=1")
	if err != nil {
		return "unkown"
	}
	dirty := ""
	if porcelain, err := syscmd.GetOutput("git", "status", "--porcelain"); err != nil {
		return "unkown"
	} else if porcelain != "" {
		dirty = "-dev"
	}

	values := strings.Split(describeTag, "-")
	switch len(values) {
	case 0:
		return "unknown"
	case 1:
		return fmt.Sprintf("%s%s", strings.Join(values[0:1], "-"), dirty)
	default:
		return fmt.Sprintf("%s%s", strings.Join(values[0:2], "-"), dirty)
	}
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Version of client and server",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		if GoVersion == "" {
			GoVersion = runtime.Version()
		}
		fmt.Println("Client:")
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
		networkingVerion, err := client.NeutronV2().GetCurrentVersion()
		if err == nil {
			fmt.Printf("  %-11s: %s\n", "Neutron", networkingVerion.VersoinInfo())
		} else {
			fmt.Printf("  %-11s: Unknown (%s)\n", "Neutron", err)
			errors++
		}
		if errors > 0 {
			os.Exit(1)
		}
	},
}

func main() {
	rootCmd := cobra.Command{
		Use:     "skyman",
		Short:   "Golang OpenStack Client \n" + LOGO,
		Version: getVersion(),
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			format, _ := cmd.Flags().GetString("format")
			if format != "" && !stringutils.ContainsString(common.GetOutputFormats(), format) {
				fmt.Printf("invalid foramt '%s'\n", format)
				os.Exit(1)
			}
			conf, _ := cmd.Flags().GetString("conf")
			if err := common.LoadConfig(conf); err != nil {
				fmt.Printf("load config failed: %v\n", err)
				os.Exit(1)
			}
			logLevel := logging.INFO
			if common.CONF.Debug {
				logLevel = logging.DEBUG
			}
			if common.CONF.Debug {
				logLevel = logging.DEBUG
			}
			if common.CONF.LogFile != "" {
				// disable log color if use log file
				common.CONF.EnableLogColor = false
			}
			logging.BasicConfig(logging.LogConfig{Level: logLevel, Output: common.CONF.LogFile, EnableColor: common.CONF.EnableLogColor})
			logging.Debug("load config file from %s", viper.ConfigFileUsed())
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

	rootCmd.PersistentFlags().String("compute-api-version", "", i18n.T("logFile"))

	rootCmd.AddCommand(
		versionCmd,
		identity.Token,
		identity.Service, identity.Endpoint, identity.Region,
		identity.User, identity.Project,
		compute.Server, compute.Flavor, compute.Hypervisor,
		compute.Keypair, compute.Compute, compute.Console,
		compute.Migration, compute.AZ, compute.Aggregate,
		image.Image,
		storage.Volume, storage.Snapshot,
		networking.Router, networking.Network, networking.Subnet, networking.Port,
		networking.Security, networking.SG,
		templates.DefineCmd, templates.UndefineCmd,
		tool.ToolCmd,
		test.TestCmd,
	)
	rootCmd.Execute()
}
