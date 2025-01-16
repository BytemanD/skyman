package nova

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"

	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/cmd/flags"
	"github.com/BytemanD/skyman/cmd/views"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/common/datatable"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/utility"
)

var (
	flavorListFlags   flags.FlavorListFlags
	flavorCreateFlags flags.FlavorCreateFlags
	flavorSetFlags    flags.FlavorSetFlags
)

var Flavor = &cobra.Command{Use: "flavor"}

func getExtraSpecsMap(extraSpecs []string) nova.ExtraSpecs {
	extraSpecsMap := nova.ExtraSpecs{}
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

		if *flavorListFlags.Public {
			query.Set("public", "true")
		}
		if *flavorListFlags.MinRam > 0 {
			query.Set("minRam", strconv.FormatUint(*flavorListFlags.MinRam, 10))
		}
		if *flavorListFlags.MinDisk > 0 {
			query.Set("minDisk", strconv.FormatUint(*flavorListFlags.MinDisk, 10))
		}
		client := openstack.DefaultClient()
		flavors, err := client.NovaV2().Flavor().Detail(query)
		utility.LogError(err, "get server failed %s", true)

		filteredFlavors := []nova.Flavor{}
		for _, flavor := range flavors {
			if *flavorListFlags.Name != "" && !strings.Contains(flavor.Name, *flavorListFlags.Name) {
				continue
			}
			if *flavorListFlags.MinVcpu > 0 && flavor.Vcpus < int(*flavorListFlags.MinVcpu) {
				continue
			}
			if *flavorListFlags.MinRam > 0 && flavor.Ram < int(*flavorListFlags.MinRam) {
				continue
			}
			if *flavorListFlags.MinDisk > 0 && flavor.Disk < int(*flavorListFlags.MinDisk) {
				continue
			}
			filteredFlavors = append(filteredFlavors, flavor)
		}
		if *flavorListFlags.Long {
			// 提前查询，否则输出json、yaml 格式时缺失
			for i, item := range filteredFlavors {
				extraSpecs, err := client.NovaV2().Flavor().ListExtraSpecs(item.Id)
				if err != nil {
					console.Fatal("get flavor extra specs failed %s", err)
				}
				filteredFlavors[i].ExtraSpecs = extraSpecs
			}
		}
		common.PrintItems(
			[]datatable.Column[nova.Flavor]{
				{Name: "Id"}, {Name: "Name"},
				{Name: "Vcpus", Align: text.AlignRight},
				{Name: "Ram", Align: text.AlignRight,
					RenderFunc: func(item nova.Flavor) interface{} { return item.HumanRam() },
				},
				{Name: "Disk", Align: text.AlignRight},
				{Name: "Ephemeral", Align: text.AlignRight},
				{Name: "IsPublic"},
			},
			[]datatable.Column[nova.Flavor]{
				{Name: "Swap"}, {Name: "RXTXFactor", Text: "RXTX Factor"},
				{Name: "ExtraSpecs",
					RenderFunc: func(item nova.Flavor) interface{} {
						return strings.Join(item.ExtraSpecs.GetList(), "\n")
					},
				},
			},
			filteredFlavors,
			common.TableOptions{
				SeparateRows: *flavorListFlags.Long,
				More:         *flavorListFlags.Long},
		)
	},
}
var flavorShow = &cobra.Command{
	Use:   "show <flavor id or name>",
	Short: "Show flavor",
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		client := openstack.DefaultClient()
		idOrName := args[0]
		flavor, err := client.NovaV2().Flavor().ShowWithExtraSpecs(idOrName)
		utility.LogError(err, "Show flavor failed", true)
		views.PrintFlavor(*flavor)
	},
}
var flavorDelete = &cobra.Command{
	Use:   "delete <flavor1> [flavor2 ...]",
	Short: "Delete flavor(s)",

	Args: cobra.MinimumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		client := openstack.DefaultClient()
		flavorApi := client.NovaV2().Flavor()
		for _, flavorId := range args {
			flavor, err := flavorApi.Find(flavorId, false)
			if err != nil {
				utility.LogError(err, "Get flavor failed", false)
				continue
			}
			err = flavorApi.Delete(flavor.Id)
			utility.LogError(err, "Delete flavor failed", false)

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

		reqFlavor := nova.Flavor{
			Name:       name,
			Vcpus:      int(vcpus),
			Ram:        int(ram),
			Disk:       int(*flavorCreateFlags.Disk),
			Swap:       int(*flavorCreateFlags.Swap),
			Ephemeral:  int(*flavorCreateFlags.Ephemeral),
			IsPublic:   !*flavorCreateFlags.Private,
			RXTXFactor: *flavorCreateFlags.RxtxFactor,
		}
		if *flavorCreateFlags.Id != "" {
			reqFlavor.Id = *flavorCreateFlags.Id
		}

		client := openstack.DefaultClient()

		flavor, err := client.NovaV2().Flavor().Create(reqFlavor)
		utility.LogError(err, "create flavor failed", true)

		if len(*flavorCreateFlags.Properties) > 0 {
			extraSpecs := getExtraSpecsMap(*flavorCreateFlags.Properties)
			createdExtraSpecs, err := client.NovaV2().Flavor().SetExtraSpecs(flavor.Id, extraSpecs)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			flavor.ExtraSpecs = createdExtraSpecs
		}
		views.PrintFlavor(*flavor)
	},
}
var flavorSet = &cobra.Command{
	Use:   "set <flavor id or name>",
	Short: "Set flavor properties",
	Args:  cobra.ExactArgs(1),
	Run: func(cmds *cobra.Command, args []string) {
		client := openstack.DefaultClient()
		idOrName := args[0]

		extraSpecs := map[string]string{}

		for _, property := range *flavorSetFlags.Properties {
			splited := strings.Split(property, "=")
			if len(splited) != 2 {
				console.Fatal("Invalid property %s, must be: key=value", property)
			}
			extraSpecs[splited[0]] = splited[1]
		}

		flavor, err := client.NovaV2().Flavor().Find(idOrName, true)

		utility.LogError(err, "Get flavor failed", true)
		client.NovaV2().Flavor().SetExtraSpecs(flavor.Id, extraSpecs)
		for _, property := range *flavorSetFlags.NoProperties {
			if _, ok := flavor.ExtraSpecs[property]; !ok {
				continue
			}
			err := client.NovaV2().Flavor().DeleteExtraSpec(flavor.Id, property)
			if err != nil {
				utility.LogError(err, "delete extra spec failed", false)
			}
		}
	},
}

func init() {
	flavorListFlags = flags.FlavorListFlags{
		Public:  flavorList.Flags().Bool("public", false, "List public flavors"),
		Name:    flavorList.Flags().StringP("name", "n", "", "Show flavors matched by name (local)"),
		MinVcpu: flavorList.Flags().Uint64("min-vcpu", 0, "Filters the flavors by a minimum vcpu (local)"),
		MinRam:  flavorList.Flags().Uint64("min-ram", 0, "Filters the flavors by a minimum RAM, in MB."),
		MinDisk: flavorList.Flags().Uint64("min-disk", 0, "Filters the flavors by a minimum disk space, in GiB."),
		Long:    flavorList.Flags().BoolP("long", "l", false, "List additional fields in output"),
		Human:   flavorList.Flags().Bool("human", false, " Print ram like 1M 2G etc"),
	}
	flavorCreateFlags = flags.FlavorCreateFlags{
		Id:         flavorCreate.Flags().String("id", "", "Unique flavor ID, creates a UUID if empty"),
		Disk:       flavorCreate.Flags().Uint("disk", 0, "Disk size in GB"),
		Swap:       flavorCreate.Flags().Uint("swap", 0, "Swap space size in MB"),
		Ephemeral:  flavorCreate.Flags().Uint("ephemeral", 0, "Swap space size in MB"),
		RxtxFactor: flavorCreate.Flags().Float32("rxtx-factor", 0, "RX/TX factor"),
		Private:    flavorCreate.Flags().Bool("private", false, "Flavor is not available to other projects"),
		Properties: flavorCreate.Flags().StringArrayP("property", "p", []string{},
			"Property to add for this flavor (repeat option to set multiple properties)"),
	}
	flavorSetFlags = flags.FlavorSetFlags{
		Properties: flavorSet.Flags().StringArrayP("property", "p", []string{},
			"Property to add or modify for this flavor (repeat option to set multiple properties)"),
		NoProperties: flavorSet.Flags().StringArrayP("no-property", "r", []string{},
			"Property to remove for this flavor (repeat option to set multiple properties)"),
	}

	Flavor.AddCommand(flavorList, flavorShow, flavorCreate, flavorDelete,
		flavorSet)
}
