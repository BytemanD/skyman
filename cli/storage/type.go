package storage

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/BytemanD/skyman/cli"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack/storage"
	"github.com/BytemanD/skyman/utility"
	"github.com/spf13/cobra"
)

var VolumeType = &cobra.Command{Use: "type"}

var typeList = &cobra.Command{
	Use:   "list",
	Short: "List volume types",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(0)(cmd, args); err != nil {
			return err
		}
		public, _ := cmd.Flags().GetBool("public")
		private, _ := cmd.Flags().GetBool("private")
		if public && private {
			return fmt.Errorf("argument --private not allowed with argument --public")
		}
		argDefault, _ := cmd.Flags().GetBool("default")
		if argDefault {
			if public {
				return fmt.Errorf("argument --default not allowed with argument --public")
			}
			if private {
				return fmt.Errorf("argument --default not allowed with argument --private")
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, _ []string) {
		client := cli.GetClient()

		long, _ := cmd.Flags().GetBool("long")
		public, _ := cmd.Flags().GetBool("public")
		private, _ := cmd.Flags().GetBool("private")
		argDefault, _ := cmd.Flags().GetBool("default")

		volumeTypes := []storage.VolumeType{}
		var err error

		if argDefault {
			volumeType, err := client.StorageClient().VolumeTypeDefaultGet()
			volumeTypes = append(volumeTypes, *volumeType)
			utility.RaiseIfError(err, "list default volume falied")
		} else {
			query := url.Values{}
			if public {
				query.Set("is_public", "true")
			}
			if private {
				query.Set("is_public", "false")
			}
			volumeTypes, err = client.StorageClient().VolumeTypeList(query)
			utility.RaiseIfError(err, "list volume type falied")
		}

		table := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id"}, {Name: "Name"}, {Name: "IsPublic"},
			},
			LongColumns: []common.Column{
				{Name: "Description"},
				{Name: "ExtraSpecs", Slot: func(item interface{}) interface{} {
					obj, _ := (item).(storage.VolumeType)
					return strings.Join(obj.GetExtraSpecsList(), "\n")
				}},
			},
		}
		table.AddItems(volumeTypes)
		if long {
			table.StyleSeparateRows = true
		}
		common.PrintPrettyTable(table, long)
	},
}
var typeShow = &cobra.Command{
	Use:   "show <id or name>",
	Short: "Show volume type",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()
		idOrName := args[0]
		volumeType, err := client.StorageClient().VolumeTypeFound(idOrName)
		utility.LogError(err, "get volume type failed", true)
		printVolumeType(*volumeType)
	},
}
var typeDefault = &cobra.Command{
	Use:   "default",
	Short: "Show default volume type",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()
		volumeType, err := client.StorageClient().VolumeTypeShow("default")
		utility.LogError(err, "get default volume type failed", true)
		printVolumeType(*volumeType)
	},
}
var typeCreate = &cobra.Command{
	Use:   "create <name>",
	Short: "Create volume type",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		public, _ := cmd.Flags().GetBool("public")
		private, _ := cmd.Flags().GetBool("private")
		if public && private {
			return fmt.Errorf("argument --private not allowed with argument --public")
		}
		properties, _ := cmd.Flags().GetStringArray("property")
		for _, property := range properties {
			if _, err := common.SplitKeyValue(property); err != nil {
				return err
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		public, _ := cmd.Flags().GetBool("public")
		private, _ := cmd.Flags().GetBool("private")
		properties, _ := cmd.Flags().GetStringArray("property")

		params := map[string]interface{}{
			"name": args[0],
		}
		if public {
			params["is_public"] = true
			params["os-volume-type-access:is_public"] = true
		}
		if private {
			params["is_public"] = false
			params["os-volume-type-access:is_public"] = false
		}
		if len(properties) > 0 {
			extraSpecs := map[string]interface{}{}
			for _, property := range properties {
				kv, _ := common.SplitKeyValue(property)
				extraSpecs[kv[0]] = kv[1]
			}
			params["extra_specs"] = extraSpecs
		}

		client := cli.GetClient()
		volume, err := client.StorageClient().VolumeTypeCreate(params)
		utility.LogError(err, "create volume type failed", true)
		printVolumeType(*volume)
	},
}
var typeDelete = &cobra.Command{
	Use:   "delete <type1> [<type2> ...]",
	Short: "Delete volume type",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()

		for _, idOrName := range args {
			volumeType, err := client.StorageClient().VolumeTypeFound(idOrName)
			if err != nil {
				utility.LogError(err, "get volume type failed", false)
				continue
			}
			err = client.StorageClient().VolumeTypeDelete(volumeType.Id)
			if err != nil {
				utility.LogError(err, fmt.Sprintf("delete volume type %s failed", idOrName), false)
			} else {
				fmt.Printf("Requested to delete volume type %s\n", idOrName)
			}
		}
	},
}

func init() {
	typeList.Flags().BoolP("long", "l", false, "List additional fields in output")
	typeList.Flags().Bool("public", false, "List only public types")
	typeList.Flags().Bool("private", false, "List only private types(admin only)")
	typeList.Flags().Bool("default", false, "List the default volume type")

	typeCreate.Flags().Bool("public", false, "Volume type is accessible to the public")
	typeCreate.Flags().Bool("private", false, "Volume type is not accessible to the public")
	typeCreate.Flags().StringArrayP("property", "p", []string{},
		"Set a property on this volume type (repeat option to set multiple properties)")

	VolumeType.AddCommand(typeList, typeShow, typeCreate, typeDelete, typeDefault)
	Volume.AddCommand(VolumeType)
}
