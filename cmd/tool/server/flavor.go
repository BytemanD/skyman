package server

import (
	"fmt"
	"strings"

	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/cmd/views"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/utility"
	"github.com/spf13/cobra"
)

var FlavorCommand = &cobra.Command{Use: "flavor", Short: "flavor tools"}

var flavorClone = &cobra.Command{
	Use:     "clone <flavor id> <new flavor name>",
	Aliases: []string{"copy"},
	Short:   "clone flavor",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(2)(cmd, args); err != nil {
			return err
		}
		properties, _ := cmd.Flags().GetStringArray("set")
		for _, property := range properties {
			kv := strings.Split(property, "=")
			if len(kv) != 2 {
				return fmt.Errorf("invalid property '%s', it must be format: key1=value1", property)
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		flavorId, newName := args[0], args[1]
		newId, _ := cmd.Flags().GetString("id")
		vcpus, _ := cmd.Flags().GetUint("vcpus")
		ram, _ := cmd.Flags().GetUint("ram")
		disk, _ := cmd.Flags().GetUint("disk")
		swap, _ := cmd.Flags().GetUint("swap")
		ephemeral, _ := cmd.Flags().GetUint("ephemeral")
		rxtxFactor, _ := cmd.Flags().GetFloat32("rxtx-factor")
		setProperties, _ := cmd.Flags().GetStringArray("set")
		unsetProperties, _ := cmd.Flags().GetStringArray("unset")

		cmdent := openstack.DefaultClient()
		novaClient := cmdent.NovaV2()
		console.Info("show flavor")
		flavor, err := novaClient.Flavor().Show(flavorId)
		utility.LogError(err, "show flavor failed", true)

		flavor.Name = newName
		flavor.Id = newId
		if vcpus != 0 {
			flavor.Vcpus = int(vcpus)
		}
		if ram != 0 {
			flavor.Ram = int(ram)
		}
		if disk != 0 {
			flavor.Disk = int(disk)
		}
		if swap != 0 {
			flavor.Swap = int(swap)
		} else {
			if _, ok := flavor.Swap.(string); ok {
				flavor.Swap = 0
			}
		}
		if ephemeral != 0 {
			flavor.Ephemeral = int(ephemeral)
		}
		if rxtxFactor != 0 {
			flavor.RXTXFactor = rxtxFactor
		}
		console.Info("show flavor extra specs")
		extraSpecs, err := novaClient.Flavor().ListExtraSpecs(flavorId)
		utility.LogError(err, "show flavor extra specs failed", true)

		newProperties := nova.ParseExtraSpecsMap(setProperties)
		for k, v := range newProperties {
			extraSpecs[k] = v
		}
		for _, k := range unsetProperties {
			delete(extraSpecs, k)
		}
		console.Info("create new flavor")
		newFlavor, err := novaClient.Flavor().Create(*flavor)
		utility.LogError(err, "create flavor failed", true)

		if len(extraSpecs) != 0 {
			console.Info("set new flavor extra specs")
			_, err = novaClient.Flavor().SetExtraSpecs(newFlavor.Id, extraSpecs)
			utility.LogError(err, "set new flavor extra specs failed", true)
		}

		newFlavor, err = novaClient.Flavor().ShowWithExtraSpecs(newFlavor.Id)
		utility.LogError(err, "show new flavor", true)
		views.PrintFlavor(*newFlavor)
	},
}

func init() {
	flavorClone.Flags().String("id", "", "New flavor ID, creates a UUID if empty")
	flavorClone.Flags().Uint("vcpus", 0, "Number of vcpus")
	flavorClone.Flags().Uint("ram", 0, "Memory size in MB")
	flavorClone.Flags().Uint("disk", 0, "Disk size in GB")
	flavorClone.Flags().Uint("swap", 0, "Swap space size in MB")
	flavorClone.Flags().Uint("ephemeral", 0, "Swap space size in MB")
	flavorClone.Flags().Float32("rxtx-factor", 0, "RX/TX factor")
	flavorClone.Flags().StringArray("set", []string{},
		"Set property to for new flavor (repeat option to set multiple properties)")
	flavorClone.Flags().StringArray("unset", []string{},
		"Unset property for new flavor (repeat option to set multiple properties)")

	FlavorCommand.AddCommand(flavorClone)
}
