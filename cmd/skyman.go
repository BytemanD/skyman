package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/BytemanD/easygo/pkg/global/gitutils"
	"github.com/BytemanD/easygo/pkg/global/logging"

	"github.com/BytemanD/skyman/cli/compute"
	"github.com/BytemanD/skyman/cli/identity"
	"github.com/BytemanD/skyman/cli/image"
	"github.com/BytemanD/skyman/cli/networking"
	"github.com/BytemanD/skyman/cli/storage"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/common/i18n"
	"github.com/BytemanD/skyman/openstack"
)

var (
	Version       string
	GoVersion     string
	BuildDate     string
	BuildPlatform string
)

func getVersion() string {
	if Version == "" {
		return gitutils.GetVersion()
	}
	return fmt.Sprint(Version)
}

var LOGO = `
      _                             
  ___| |  _ _   _ ____  _____ ____  
 /___) |_/ ) | | |    \(____ |  _ \ 
|___ |  _ (| |_| | | | / ___ | | | |
(___/|_| \_)\__  |_|_|_\_____|_| |_|
           (____/    
`

func main() {
	rootCmd := cobra.Command{
		Use:     "skyman",
		Short:   "Golang OpenStack Client \n" + LOGO,
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

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Version of client and server",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, _ []string) {
			if GoVersion == "" {
				GoVersion = runtime.Version()
			}
			if BuildDate == "" {
				BuildDate = time.Now().Format("2006-01-02 15:04:05")
			}
			if BuildPlatform == "" {
				BuildPlatform = common.Uname()
			}

			fmt.Println("Version:       ", getVersion())
			fmt.Println("GoVersion:     ", GoVersion)
			fmt.Println("BuildDate:     ", BuildDate)
			fmt.Println("BuildPlatform: ", BuildPlatform)

			client := openstack.CreateInstance()
			fmt.Println("Version of servers:")

			identityVerions, err := client.Identity.GetStableVersion()
			if err == nil {
				fmt.Println("  Identity:")
				if v := identityVerions; v != nil {
					fmt.Println("    Version:", v.Id)
					if v.MinVersion != "" || v.Version != "" {
						fmt.Printf("    MicroVersion: %s ~ %s\n", v.MinVersion, v.Version)
					}
				}
			}

			computeVerions, err := client.Compute.GetCurrentVersion()
			if err == nil {
				v := computeVerions
				fmt.Println("  Compute:")
				fmt.Println("    Version:", v.Id)
				if v.MinVersion != "" || v.Version != "" {
					fmt.Printf("    MicroVersion: %s ~ %s\n", v.MinVersion, v.Version)
				}
			}

			imageVerions, err := client.Image.GetCurrentVersion()
			if err == nil {
				v := imageVerions
				fmt.Println("  Image:")
				fmt.Println("    Version:", v.Id)
				if v.MinVersion != "" || v.Version != "" {
					fmt.Printf("    MicroVersion: %s ~ %s\n", v.MinVersion, v.Version)
				}
			} else {
				logging.Error("get image api version %s", err)
			}
			storageVerions, err := client.Storage.GetCurrentVersion()
			if err == nil {
				v := storageVerions
				fmt.Println("  Storage:")
				fmt.Println("    Version:", v.Id)
				if v.MinVersion != "" || v.Version != "" {
					fmt.Printf("    MicroVersion: %s ~ %s\n", v.MinVersion, v.Version)
				}
			} else {
				logging.Error("get image api version %s", err)
			}
		},
	}

	rootCmd.PersistentFlags().BoolP("debug", "d", false, i18n.T("showDebug"))
	rootCmd.PersistentFlags().StringP("conf", "c", "", i18n.T("thePathOfConfigFile"))
	rootCmd.PersistentFlags().StringP("format", "f", "default",
		fmt.Sprintf(i18n.T("formatAndSupported"), common.GetOutputFormats()))

	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("format", rootCmd.PersistentFlags().Lookup("format"))
	rootCmd.AddCommand(
		versionCmd,
		identity.Token,
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
