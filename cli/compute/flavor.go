package compute

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/stackcrud/cli"
	libCommon "github.com/BytemanD/stackcrud/common"
	"github.com/BytemanD/stackcrud/openstack/common"
	"github.com/BytemanD/stackcrud/openstack/compute"
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
		flavors, err := client.Compute.FlavorListDetail(nil)
		if err != nil {
			logging.Fatal("%s", err)
		}
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
		dataListTable := libCommon.DataListTable{
			ShortHeaders: []string{
				"Id", "Name", "Vcpus", "Ram", "Disk", "Ephemeral", "IsPublic"},
			LongHeaders: []string{
				"Swap", "RXTXFactor", "ExtraSpecs"},
			HeaderLabel: map[string]string{"RXTXFactor": "RXTX Factor"},
			SortBy:      []table.SortBy{{Name: "Name", Mode: table.Asc}},
			Slots: map[string]func(item interface{}) interface{}{
				"ExtraSpecs": func(item interface{}) interface{} {
					p, _ := (item).(compute.Flavor)
					return strings.Join(p.ExtraSpecs.GetList(), "\n")
				},
			},
		}
		if long {
			for i, flavor := range filteredFlavors {
				extraSpecs, err := client.Compute.FlavorExtraSpecsList(flavor.Id)
				if err != nil {
					logging.Fatal("get flavor extra specs failed %s", err)
				}
				filteredFlavors[i].ExtraSpecs = extraSpecs
			}
			dataListTable.StyleSeparateRows = true
		}
		for _, flavor := range filteredFlavors {
			dataListTable.Items = append(dataListTable.Items, flavor)
		}
		libCommon.PrintDataListTable(dataListTable, long)
	},
}
var flavorShow = &cobra.Command{
	Use:   "show <flavor id>",
	Short: "Show flavor",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()
		flavorId := args[0]
		flavor, err := client.Compute.FlavorShowWithExtraSpecs(flavorId)
		if err != nil {
			if httpError, ok := err.(*common.HttpError); ok {
				logging.Fatal("Show flavor %s failed, %s", flavorId, httpError.Message)
			} else {
				logging.Fatal("Show flavor %s failed, %v", flavorId, err)
			}
		}
		printFlavor(*flavor)
	},
}
var flavorDelete = &cobra.Command{
	Use:   "delete <flavor1> [flavor2 ...]",
	Short: "Delete flavor(s)",

	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()
		for _, flavorId := range args {
			err := client.Compute.FlavorDelete(flavorId)
			if err != nil {
				if httpError, ok := err.(*common.HttpError); ok {
					logging.Fatal("Delete flavor %s failed, %s", flavorId, httpError.Message)
				} else {
					logging.Fatal("Delete flavor %s failed, %v", flavorId, err)
				}
			} else {
				fmt.Printf("Delete flavor success: %s\n", flavorId)
			}
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

		flavor, err := client.Compute.FlavorCreate(reqFlavor)
		if err != nil {
			logging.Fatal("%s", err)
		}
		extraSpecs := getExtraSpecsMap(properties)
		createdExtraSpecs, err := client.Compute.FlavorExtraSpecsCreate(flavor.Id, extraSpecs)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		flavor.ExtraSpecs = createdExtraSpecs
		printFlavor(*flavor)
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

		newFlavor, err := client.Compute.FlavorCopy(flavorId, newName, newId,
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
		flavorCopy)
}
