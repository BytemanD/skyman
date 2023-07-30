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
		dataTable := cli.DataListTable{
			ShortHeaders: []string{
				"Id", "Name", "Vcpus", "Ram", "Disk", "Ephemeral", "IsPublic"},
			LongHeaders: []string{
				"Swap", "RXTXFactor", "ExtraSpecs"},
			HeaderLabel: map[string]string{
				"IsPublic":   "Is Public",
				"RXTXFactor": "RXTX Factor",
				"ExtraSpecs": "Extra Specs",
			},
			SortBy: []table.SortBy{{Name: "Name", Mode: table.Asc}},
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
		}
		for _, flavor := range filteredFlavors {
			dataTable.Items = append(dataTable.Items, flavor)
		}
		dataTable.Print(long)
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
		logging.Info("properties: %v", properties)
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
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		flavorId := args[0]
		newName := args[1]
		newId, _ := cmd.Flags().GetString("id")

		client := cli.GetClient()

		logging.Info("Show flavor")
		flavor, err := client.Compute.FlavorShow(flavorId)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		flavor.Name = newName
		flavor.Id = newId

		logging.Info("Show flavor extra specs")
		extraSpecs, err := client.Compute.FlavorExtraSpecsList(flavorId)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		logging.Info("Create new flavor")
		newFlavor, err := client.Compute.FlavorCreate(*flavor)
		if err != nil {
			fmt.Printf("create flavor failed, %v", err)
			os.Exit(1)
		}
		if len(extraSpecs) != 0 {
			logging.Info("Set new flavor extra specs")
			_, err = client.Compute.FlavorExtraSpecsCreate(newFlavor.Id, extraSpecs)
			if err != nil {
				fmt.Printf("set flavor extra specs failed, %v", err)
				os.Exit(1)
			}
			newFlavor.ExtraSpecs = extraSpecs
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

	Flavor.AddCommand(flavorList, flavorCreate, flavorCopy)
}
