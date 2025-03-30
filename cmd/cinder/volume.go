package cinder

import (
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack/model/cinder"
	"github.com/BytemanD/skyman/utility"
	"github.com/spf13/cobra"
)

var Volume = &cobra.Command{Use: "volume"}

var volumeList = &cobra.Command{
	Use:   "list",
	Short: "List volumes",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		client := common.DefaultClient()

		long, _ := cmd.Flags().GetBool("long")
		name, _ := cmd.Flags().GetString("name")
		status, _ := cmd.Flags().GetString("status")
		all, _ := cmd.Flags().GetBool("all")
		sort, _ := cmd.Flags().GetString("sort")
		limit, _ := cmd.Flags().GetInt("limit")

		query := url.Values{}
		if name != "" {
			query.Set("name", name)
		}
		if status != "" {
			query.Set("status", status)
		}
		if sort != "" {
			query.Set("sort", sort)
		}
		if limit > 0 {
			query.Set("limit", strconv.Itoa(limit))
		}
		if all {
			query.Set("all_tenants", "true")
		}
		volumes, err := client.CinderV2().ListVolume(query, true)
		utility.LogError(err, "list volume falied", true)
		common.PrintVolumes(volumes, long)
	},
}

var volumeShow = &cobra.Command{
	Use:   "show <id or name>",
	Short: "Show volume",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := common.DefaultClient()
		volume, err := client.CinderV2().FindVolume(args[0], client.IsAdmin())
		utility.LogError(err, "get volume failed", true)
		common.PrintVolume(*volume)
	},
}
var volumeDelete = &cobra.Command{
	Use:   "delete <volume1> [<volume2> ...]",
	Short: "Delete volume",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := common.DefaultClient()
		force, _ := cmd.Flags().GetBool("force")
		cascade, _ := cmd.Flags().GetBool("cascade")

		for _, idOrName := range args {
			volume, err := client.CinderV2().FindVolume(idOrName)
			if err != nil {
				utility.LogError(err, "get volume failed", false)
				continue
			}
			err = client.CinderV2().DeleteVolume(volume.Id, force, cascade)
			if err == nil {
				println("Requested to delete volume", idOrName)
			} else {
				println(err)
			}
		}
	},
}

var volumeCreate = &cobra.Command{
	Use:   "create <name>",
	Short: "Create volume",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		size, _ := cmd.Flags().GetUint("size")
		volumeType, _ := cmd.Flags().GetString("type")
		multiattach, _ := cmd.Flags().GetBool("multiattach")

		params := map[string]interface{}{"name": args[0]}
		if size > 0 {
			params["size"] = size
		}
		if volumeType != "" {
			params["volume_type"] = volumeType
		}
		if multiattach {
			params["multiattach"] = multiattach
		}

		client := common.DefaultClient()

		volume, err := client.CinderV2().CreateVolume(params)
		if err != nil {
			println(err)
			os.Exit(1)
		}
		common.PrintVolume(*volume)
	},
}
var volumeExtend = &cobra.Command{
	Use:   "extend <volume> <new size>",
	Short: "Extend volume",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(2)(cmd, args); err != nil {
			return err
		}
		if _, err := strconv.Atoi(args[1]); err != nil {
			return err
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		idOrName := args[0]
		size, _ := strconv.Atoi(args[1])
		client := common.DefaultClient()
		volume, err := client.CinderV2().FindVolume(idOrName)
		utility.LogError(err, "get volume falied", true)

		err = client.CinderV2().ExtendVolume(volume.Id, size)
		utility.LogError(err, "extend volume falied", true)
	},
}
var volumeRetype = &cobra.Command{
	Use:   "retype <volume> <new type>",
	Short: "Retype volume",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(2)(cmd, args); err != nil {
			return err
		}
		migrationPolicy, _ := cmd.Flags().GetString("migration-policy")
		if err := cinder.InvalidMIgrationPoicy(migrationPolicy); err != nil {
			return err
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		idOrName := args[0]
		newType := args[1]
		migrationPolicy, _ := cmd.Flags().GetString("migration-policy")

		client := common.DefaultClient()
		volume, err := client.CinderV2().FindVolume(idOrName)
		utility.LogError(err, "get volume falied", true)

		err = client.CinderV2().RetypeVolume(volume.Id, newType, migrationPolicy)
		utility.LogError(err, "extend volume falied", true)
	},
}

func init() {
	volumeList.Flags().BoolP("long", "l", false, "List additional fields in output")
	volumeList.Flags().Bool("all", false, "List volumes of all tenants")
	volumeList.Flags().StringP("name", "n", "", "Search by volume name")
	volumeList.Flags().String("status", "", "Search by volume status")
	volumeList.Flags().String("sort", "", "Sort by specified field")
	volumeList.Flags().Int("limit", 0, "limit")

	volumeCreate.Flags().Uint("size", 0, "Volume size (GB)")
	volumeCreate.Flags().String("type", "", "Volume type")
	volumeCreate.Flags().Bool("multiattach", false, "Allow multiattach")
	volumeCreate.MarkFlagRequired("size")

	volumeDelete.Flags().Bool("force", false, "Force delete volume.")
	volumeDelete.Flags().Bool("cascade", false, "Remove any snapshots along with volume")

	volumeRetype.Flags().StringP("migration-policy", "p", "never",
		fmt.Sprintf("Migration policy during retype of volume,\ninvalid values: %s",
			cinder.MIGRATION_POLICYS))

	Volume.AddCommand(
		volumeList, volumeShow, volumeCreate, volumeExtend, volumeRetype,
		volumeDelete,
	)
}
