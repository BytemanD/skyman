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

	"github.com/BytemanD/skyman/cli"
	"github.com/BytemanD/skyman/cli/compute"
	"github.com/BytemanD/skyman/cli/identity"
	"github.com/BytemanD/skyman/cli/image"
	"github.com/BytemanD/skyman/cli/networking"
	"github.com/BytemanD/skyman/cli/storage"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/common/i18n"
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
		if BuildDate == "" {
			BuildDate = time.Now().Format("2006-01-02 15:04:05")
		}
		if BuildPlatform == "" {
			BuildPlatform = common.Uname()
		}

		fmt.Println("Client:")
		fmt.Printf("  %-14s: %s\n", "Version", getVersion())
		fmt.Printf("  %-14s: %s\n", "GoVersion", GoVersion)
		fmt.Printf("  %-14s: %s\n", "BuildDate", BuildDate)
		fmt.Printf("  %-14s: %s\n", "BuildPlatform", BuildPlatform)

		client := cli.GetClient()
		fmt.Println("Servers:")

		identityVerion, err := client.Identity.GetStableVersion()
		common.LogError(err, "get idendity veresion failed", true)
		if err != nil {
			return
		}
		fmt.Printf("  %-11s: %s\n", "Identity", identityVerion.VersoinInfo())

		errors := []string{}
		compute := client.ComputeClient()
		computeVerion, err := compute.GetCurrentVersion()
		if err == nil {
			fmt.Printf("  %-11s: %s\n", "Compute", computeVerion.VersoinInfo())
		} else {
			errors = append(errors, fmt.Sprintf("get compute api version failed %s", err))
		}
		imageVerion, err := client.ImageClient().GetCurrentVersion()
		if err == nil {
			fmt.Printf("  %-11s: %s\n", "Image", imageVerion.VersoinInfo())
		} else {
			errors = append(errors, fmt.Sprintf("get image api version failed %s", err))
		}
		storageVerion, err := client.StorageClient().GetCurrentVersion()
		if err == nil {
			fmt.Printf("  %-11s: %s\n", "Storage", storageVerion.VersoinInfo())
		} else {
			errors = append(errors, fmt.Sprintf("get storage api version failed %s", err))
		}
		networkingVerion, err := client.NetworkingClient().GetCurrentVersion()
		if err == nil {
			fmt.Printf("  %-11s: %s\n", "Networking", networkingVerion.VersoinInfo())
		} else {
			errors = append(errors, fmt.Sprintf("get networking api version failed %s", err))
		}
		for _, err := range errors {
			logging.Error("%s", err)
		}
	},
}

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

	rootCmd.PersistentFlags().BoolP("debug", "d", false, i18n.T("showDebug"))
	rootCmd.PersistentFlags().StringP("conf", "c", "", i18n.T("thePathOfConfigFile"))
	rootCmd.PersistentFlags().StringP("format", "f", "default",
		fmt.Sprintf(i18n.T("formatAndSupported"), common.GetOutputFormats()))

	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("format", rootCmd.PersistentFlags().Lookup("format"))
	rootCmd.AddCommand(
		versionCmd,
		identity.Token,
		identity.Service, identity.Endpoint, identity.Region,
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
