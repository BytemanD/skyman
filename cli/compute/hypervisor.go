package compute

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/BytemanD/skyman/cli"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack/compute"
	"github.com/BytemanD/skyman/utility"
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
			query.Set("with_servers", "true")
		}
		if name != "" {
			query.Set("hypervisor_hostname_pattern", name)
		}
		hypervisors, err := client.ComputeClient().HypervisorListDetail(query)
		utility.LogError(err, "list hypervisors failed", true)
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
		if withServers {
			pt.StyleSeparateRows = true
			pt.ShortColumns = append(pt.ShortColumns,
				common.Column{Name: "servers", Slot: func(item interface{}) interface{} {
					p, _ := item.(compute.Hypervisor)
					hypervisorServers := []string{}
					for _, s := range p.Servers {
						hypervisorServers = append(hypervisorServers,
							fmt.Sprintf("%s(%s)", s.UUID, s.Name),
						)
					}
					return strings.Join(hypervisorServers, "\n")
				}})
		}
		pt.AddItems(hypervisors)
		common.PrintPrettyTable(pt, long)
	},
}
var hypervisorShow = &cobra.Command{
	Use:   "show <id or name>",
	Short: "Show hypervisor",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()
		hypervisor, err := client.ComputeClient().HypervisorFound(args[0])
		utility.LogError(err, "get hypervisor failed", true)

		pt := common.PrettyItemTable{
			ShortFields: []common.Column{
				{Name: "Id"}, {Name: "Hostname"}, {Name: "HostIp"},
				{Name: "Status"}, {Name: "State"},
				{Name: "Type"}, {Name: "Version"},
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
var hypervisorUptime = &cobra.Command{
	Use:   "uptime <id or name>",
	Short: "uptime hypervisor",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()
		hypervisor, err := client.ComputeClient().HypervisorFound(args[0])
		utility.LogError(err, "get hypervisor failed", true)

		hypervisor, err = client.ComputeClient().HypervisorUptime(hypervisor.Id)
		utility.LogError(err, "get hypervisor uptime failed", true)

		pt := common.PrettyItemTable{
			ShortFields: []common.Column{
				{Name: "Id"}, {Name: "Hostname"},
				{Name: "State"}, {Name: "Status"},
				{Name: "Uptime"},
			},
			Item: *hypervisor,
		}
		common.PrintPrettyItemTable(pt)
	},
}

func init() {
	// hypervisor list flags
	hypervisorList.Flags().StringP("name", "n", "", "Show hypervisors matched by name")
	hypervisorList.Flags().BoolP("long", "l", false, "List additional fields in output")
	hypervisorList.Flags().Bool("with-servers", false, "List hypervisors with servers")

	Hypervisor.AddCommand(hypervisorList, hypervisorShow, hypervisorUptime)
}
