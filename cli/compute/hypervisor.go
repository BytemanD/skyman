package compute

import (
	"net/url"

	"github.com/spf13/cobra"

	"github.com/BytemanD/skyman/cli"
	"github.com/BytemanD/skyman/common"
	openstackCommon "github.com/BytemanD/skyman/openstack/common"
	"github.com/BytemanD/skyman/openstack/compute"
)

var Hypervisor = &cobra.Command{Use: "hypervisor"}

var hypervisorList = &cobra.Command{
	Use:   "list",
	Short: "List hypervisors",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		client := cli.GetClient()

		long, _ := cmd.Flags().GetBool("long")
		name, _ := cmd.Flags().GetString("name")
		withServers, _ := cmd.Flags().GetBool("with-servers")

		query := url.Values{}
		if withServers {
			// TODO
			query.Set("with_servers", "true")
		}
		if name != "" {
			query.Set("hypervisor_hostname_pattern", name)
		}
		hypervisors, err := client.ComputeClient().HypervisorListDetail(query)
		common.LogError(err, "list hypervisors failed", true)
		pt := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id"}, {Name: "Hostname"}, {Name: "HostIp"},
				{Name: "Status", AutoColor: true}, {Name: "State", AutoColor: true},
			},
			LongColumns: []common.Column{
				{Name: "Type"}, {Name: "Version"},
				{Name: "Vcpus"}, {Name: "VcpusUsed"},
				{Name: "MemoryMB", Text: "Memory(MB)"},
				{Name: "MemoryMBUsed", Text: "Memory Used(MB)"},
			},
		}
		pt.AddItems(hypervisors)
		common.PrintPrettyTable(pt, long)
	},
}
var hypervisorShow = &cobra.Command{
	Use:   "show <id>",
	Short: "Show hypervisor",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()

		var (
			hypervisor *compute.Hypervisor
			err        error
		)
		if !openstackCommon.IsUUID(args[0]) {
			hypervisor, err = client.ComputeClient().HypervisorShowByHostname(args[0])
		} else {
			hypervisor, err = client.ComputeClient().HypervisorShow(args[0])
			if httpError, ok := err.(*openstackCommon.HttpError); ok {
				if httpError.Status == 404 {
					hypervisor, err = client.ComputeClient().HypervisorShowByHostname(args[0])
				}
			}
		}
		common.LogError(err, "get hypervisor failed", true)

		pt := common.PrettyItemTable{
			ShortFields: []common.Column{
				{Name: "Id"}, {Name: "Hostname"}, {Name: "HostIp"},
				{Name: "Status", AutoColor: true},
				{Name: "State", AutoColor: true},
				{Name: "Type"}, {Name: "Version"}, {Name: "Uptime"},
				{Name: "Vcpus"}, {Name: "VcpusUsed"},
				{Name: "MemoryMB", Text: "Memory MB"},
				{Name: "MemoryMBUsed", Text: "Memory Used MB"},
				{Name: "ExtraResources", Slot: func(item interface{}) interface{} {
					p, _ := item.(compute.Hypervisor)
					return p.ExtraResourcesMarshal(true)
				}},
				{Name: "CpuInfoArch",
					Slot: func(item interface{}) interface{} {
						p, _ := item.(compute.Hypervisor)
						return p.CpuInfo.Arch
					}},
				{Name: "CpuInfoModel",
					Slot: func(item interface{}) interface{} {
						p, _ := item.(compute.Hypervisor)
						return p.CpuInfo.Model
					}},
				{Name: "CpuInfoVendor",
					Slot: func(item interface{}) interface{} {
						p, _ := item.(compute.Hypervisor)
						return p.CpuInfo.Vendor
					},
				},
				{Name: "CpuInfoFeature",
					Slot: func(item interface{}) interface{} {
						p, _ := item.(compute.Hypervisor)
						return p.CpuInfo.Features
					},
				},
				// {Name: "Servers"},
			},
			Item: *hypervisor,
		}
		if len(hypervisor.NumaNode0Hugepages) > 0 {
			pt.ShortFields = append(pt.ShortFields,
				common.Column{Name: "NumaNode0Hugepages", Marshal: true},
				common.Column{Name: "NumaNode0Cpuset", Marshal: true},
			)
		}
		if len(hypervisor.NumaNode1Hugepages) > 0 {
			pt.ShortFields = append(pt.ShortFields,
				common.Column{Name: "NumaNode1Hugepages", Marshal: true},
				common.Column{Name: "NumaNode1Cpuset", Marshal: true},
			)
		}
		if len(hypervisor.NumaNode2Hugepages) > 0 {
			pt.ShortFields = append(pt.ShortFields,
				common.Column{Name: "NumaNode2Hugepages", Marshal: true},
				common.Column{Name: "NumaNode2Cpuset", Marshal: true},
			)
		}
		if len(hypervisor.NumaNode3Hugepages) > 0 {
			pt.ShortFields = append(pt.ShortFields,
				common.Column{Name: "NumaNode3Hugepages", Marshal: true},
				common.Column{Name: "NumaNode3Cpuset", Marshal: true},
			)
		}
		common.PrintPrettyItemTable(pt)
	},
}

func init() {
	// hypervisor list flags
	hypervisorList.Flags().StringP("name", "n", "", "Show hypervisors matched by name")
	hypervisorList.Flags().BoolP("long", "l", false, "List additional fields in output")
	hypervisorList.Flags().Bool("with-servers", false, "List hypervisors with servers")

	Hypervisor.AddCommand(hypervisorList, hypervisorShow)
}
