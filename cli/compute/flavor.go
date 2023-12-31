package compute

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/cli"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack/compute"
)

var Flavor = &cobra.Command{Use: "flavor"}

func getExtraSpecsMap(extraSpecs []string) compute.ExtraSpecs {
	extraSpecsMap := compute.ExtraSpecs{}
	for _, property := range extraSpecs {
		kv := strings.Split(property, "=")
		extraSpecsMap[kv[0]] = kv[1]
	}
	return extraSpecsMap
}

var flavorList = &cobra.Command{
	Use:   "list",
	Short: "List flavors",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		query := url.Values{}
		public, _ := cmd.Flags().GetBool("public")
		if public {
			query.Set("public", "true")
		}
		long, _ := cmd.Flags().GetBool("long")
		name, _ := cmd.Flags().GetString("name")

		client := cli.GetClient()
		flavors, err := client.ComputeClient().FlavorListDetail(nil)
		common.LogError(err, "get server failed %s", true)

		filteredFlavors := []compute.Flavor{}
		if name != "" {
			for _, flavor := range flavors {
				if strings.Contains(flavor.Name, name) {
					filteredFlavors = append(filteredFlavors, flavor)
				}
			}
		} else {
			filteredFlavors = flavors
		}
		pt := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id"}, {Name: "Name"},
				{Name: "Vcpus"}, {Name: "Ram"}, {Name: "Disk"},
				{Name: "Ephemeral"}, {Name: "IsPublic"},
			},
			LongColumns: []common.Column{
				{Name: "Swap"}, {Name: "RXTXFactor", Text: "RXTX Factor"},
				{Name: "ExtraSpecs",
					Slot: func(item interface{}) interface{} {
						p, _ := item.(compute.Flavor)
						return strings.Join(p.ExtraSpecs.GetList(), "\n")
					},
				},
			},
		}

		if long {
			pt.StyleSeparateRows = true
			for i, flavor := range filteredFlavors {
				extraSpecs, err := client.ComputeClient().FlavorExtraSpecsList(flavor.Id)
				if err != nil {
					logging.Fatal("get flavor extra specs failed %s", err)
				}
				filteredFlavors[i].ExtraSpecs = extraSpecs
			}
		}
		pt.AddItems(filteredFlavors)
		common.PrintPrettyTable(pt, long)
	},
}
var flavorShow = &cobra.Command{
	Use:   "show <flavor id or name>",
	Short: "Show flavor",
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		client := cli.GetClient()
		idOrName := args[0]
		flavor, err := client.ComputeClient().FlavorShowWithExtraSpecs(idOrName)
		common.LogError(err, "Show flavor failed", true)
		printFlavor(*flavor)
	},
}
var flavorDelete = &cobra.Command{
	Use:   "delete <flavor1> [flavor2 ...]",
	Short: "Delete flavor(s)",

	Args: cobra.MinimumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		client := cli.GetClient()
		for _, flavorId := range args {
			flavor, err := client.ComputeClient().FlavorFound(flavorId)
			common.LogError(err, "Get flavor failed", false)
			err = client.ComputeClient().FlavorDelete(flavor.Id)
			common.LogError(err, "Delete flavor failed", false)

			fmt.Printf("Flavor %s deleted \n", flavorId)
		}
	},
}
var flavorCreate = &cobra.Command{
	Use:   "create <name> <vcpus> <ram>",
	Short: "Create flavor",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(3)(cmd, args); err != nil {
			return err
		}
		if _, err := strconv.Atoi(args[1]); err != nil {
			return fmt.Errorf("invalid vcpus %s, %v", args[1], err)
		}
		if _, err := strconv.Atoi(args[2]); err != nil {
			return fmt.Errorf("invalid ram %s, %v", args[2], err)
		}
		properties, _ := cmd.Flags().GetStringArray("property")
		for _, property := range properties {
			kv := strings.Split(property, "=")
			if len(kv) != 2 {
				return fmt.Errorf("invalid property '%s', it must be format: key1=value1", property)
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		vcpus, _ := strconv.Atoi(args[1])
		ram, _ := strconv.Atoi(args[2])

		flavorId, _ := cmd.Flags().GetString("id")
		disk, _ := cmd.Flags().GetUint("disk")
		swap, _ := cmd.Flags().GetUint("swap")
		ephemeral, _ := cmd.Flags().GetUint("ephemeral")
		private, _ := cmd.Flags().GetBool("private")
		rxtxFactor, _ := cmd.Flags().GetFloat32("rxtx-factor")
		properties, _ := cmd.Flags().GetStringArray("property")

		reqFlavor := compute.Flavor{
			Name:       name,
			Vcpus:      int(vcpus),
			Ram:        int(ram),
			Disk:       int(disk),
			Swap:       int(swap),
			Ephemeral:  int(ephemeral),
			IsPublic:   !private,
			RXTXFactor: rxtxFactor,
		}
		if flavorId != "" {
			reqFlavor.Id = flavorId
		}

		client := cli.GetClient()

		flavor, err := client.ComputeClient().FlavorCreate(reqFlavor)
		common.LogError(err, "create flavor failed", true)

		extraSpecs := getExtraSpecsMap(properties)
		createdExtraSpecs, err := client.ComputeClient().FlavorExtraSpecsCreate(flavor.Id, extraSpecs)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		flavor.ExtraSpecs = createdExtraSpecs
		printFlavor(*flavor)
	},
}
var flavorSet = &cobra.Command{
	Use:   "set <flavor id or name>",
	Short: "Set flavor properties",
	Args:  cobra.ExactArgs(1),
	Run: func(cmds *cobra.Command, args []string) {
		client := cli.GetClient()
		idOrName := args[0]

		extraSpecs := map[string]string{}
		properties, _ := cmds.Flags().GetStringArray("property")
		for _, property := range properties {
			splited := strings.Split(property, "=")
			if len(splited) != 2 {
				logging.Fatal("Invalid property %s, must be: key=value", property)
			}
			extraSpecs[splited[0]] = splited[1]
		}

		flavor, err := client.ComputeClient().FlavorFound(idOrName)
		common.LogError(err, "Get flavor failed", true)
		client.ComputeClient().FlavorExtraSpecsCreate(flavor.Id, extraSpecs)
	},
}
var flavorUnset = &cobra.Command{
	Use:   "unset <flavor id or name>",
	Short: "Unset flavor properties",
	Args:  cobra.ExactArgs(1),
	Run: func(cmds *cobra.Command, args []string) {
		client := cli.GetClient()

		properties, _ := cmds.Flags().GetStringArray("property")
		flavor, err := client.ComputeClient().FlavorFound(args[0])
		common.LogError(err, "Get flavor failed", true)
		for _, property := range properties {
			client.ComputeClient().FlavorExtraSpecsDelete(flavor.Id, property)
			common.LogError(err, "delete extra specs failed", false)

		}

	},
}
var flavorCopy = &cobra.Command{
	Use:   "copy <flavor id> <new flavor name>",
	Short: "Copy flavor",
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
		flavorId := args[0]
		newName := args[1]
		newId, _ := cmd.Flags().GetString("id")
		vcpus, _ := cmd.Flags().GetUint("vcpus")
		ram, _ := cmd.Flags().GetUint("ram")
		disk, _ := cmd.Flags().GetUint("disk")
		swap, _ := cmd.Flags().GetUint("swap")
		ephemeral, _ := cmd.Flags().GetUint("ephemeral")
		rxtxFactor, _ := cmd.Flags().GetFloat32("rxtx-factor")
		setProperties, _ := cmd.Flags().GetStringArray("set")
		unSetProperties, _ := cmd.Flags().GetStringArray("unset")

		client := cli.GetClient()

		newFlavor, err := client.ComputeClient().FlavorCopy(flavorId, newName, newId,
			int(vcpus), int(ram), int(disk), int(swap), int(ephemeral), rxtxFactor,
			getExtraSpecsMap(setProperties), unSetProperties,
		)
		if err != nil {
			fmt.Printf("Copy flavor faield, %v", err)
			os.Exit(1)
		}
		printFlavor(*newFlavor)
	},
}

func init() {
	// flavor list flags
	flavorList.Flags().Bool("public", false, "List public flavors")
	flavorList.Flags().StringP("name", "n", "", "Show flavors matched by name")
	flavorList.Flags().BoolP("long", "l", false, "List additional fields in output")

	flavorCreate.Flags().String("id", "", "Unique flavor ID, creates a UUID if empty")
	flavorCreate.Flags().Uint("disk", 0, "Disk size in GB")
	flavorCreate.Flags().Uint("swap", 0, "Swap space size in MB")
	flavorCreate.Flags().Uint("ephemeral", 0, "Swap space size in MB")
	flavorCreate.Flags().Float32("rxtx-factor", 0, "RX/TX factor")
	flavorCreate.Flags().Bool("private", false, "Flavor is not available to other projects")
	flavorCreate.Flags().StringArrayP("property", "p", []string{},
		"Property to add for this flavor (repeat option to set multiple properties)")

	flavorSet.Flags().StringArrayP("property", "p", []string{},
		"Property to add or modify for this flavor (repeat option to set multiple properties)")
	flavorUnset.Flags().StringArrayP("property", "p", []string{},
		"Property to add or modify for this flavor (repeat option to set multiple properties)")

	flavorCopy.Flags().String("id", "", "New flavor ID, creates a UUID if empty")
	flavorCopy.Flags().Uint("vcpus", 0, "Number of vcpus")
	flavorCopy.Flags().Uint("ram", 0, "Memory size in MB")
	flavorCopy.Flags().Uint("disk", 0, "Disk size in GB")
	flavorCopy.Flags().Uint("swap", 0, "Swap space size in MB")
	flavorCopy.Flags().Uint("ephemeral", 0, "Swap space size in MB")
	flavorCopy.Flags().Float32("rxtx-factor", 0, "RX/TX factor")
	flavorCopy.Flags().StringArray("set", []string{},
		"Set property to for new flavor (repeat option to set multiple properties)")
	flavorCopy.Flags().StringArray("unset", []string{},
		"Unset property for new flavor (repeat option to set multiple properties)")
	Flavor.AddCommand(flavorList, flavorShow, flavorCreate, flavorDelete,
		flavorSet, flavorUnset, flavorCopy)
}
