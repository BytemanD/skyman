package nova

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/BytemanD/skyman/cmd/flags"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/utility"
)

var (
	hypervisorListFlags flags.HypervisorListFlags
)
var Hypervisor = &cobra.Command{Use: "hypervisor"}

var hypervisorList = &cobra.Command{
	Use:   "list",
	Short: "List hypervisors",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		client := openstack.DefaultClient()

		query := url.Values{}
		if *hypervisorListFlags.WithServers {
			query.Set("with_servers", "true")
		}
		if *hypervisorListFlags.Name != "" {
			query.Set("hypervisor_hostname_pattern", *hypervisorListFlags.Name)
		}
		hypervisors, err := client.NovaV2().Hypervisor().Detail(query)
		utility.LogError(err, "list hypervisors failed", true)
		pt := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id"}, {Name: "HypervisorHostname"}, {Name: "HostIp"},
				{Name: "Status", AutoColor: true},
				{Name: "State", AutoColor: true},
			},
			LongColumns: []common.Column{
				{Name: "Type"}, {Name: "Version"},
				{Name: "Vcpus"}, {Name: "VcpusUsed"},
				{Name: "MemoryMB", Text: "Memory(MB)"},
				{Name: "MemoryMBUsed", Text: "Memory Used(MB)"},
			},
			Filters: map[string]string{},
		}
		if *hypervisorListFlags.Type != "" {
			filterHypervisors := []nova.Hypervisor{}
			for _, hypervisor := range hypervisors {
				if hypervisor.Type != *hypervisorListFlags.Type {
					continue
				}
				filterHypervisors = append(filterHypervisors, hypervisor)
			}
			hypervisors = filterHypervisors
		}
		if *hypervisorListFlags.WithServers {
			pt.StyleSeparateRows = true
			pt.ShortColumns = append(pt.ShortColumns,
				common.Column{Name: "servers", Slot: func(item interface{}) interface{} {
					p, _ := item.(nova.Hypervisor)
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
		common.PrintPrettyTable(pt, *hypervisorListFlags.Long)
	},
}
var hypervisorShow = &cobra.Command{
	Use:   "show <id or name>",
	Short: "Show hypervisor",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := openstack.DefaultClient()
		hypervisor, err := client.NovaV2().Hypervisor().Find(args[0])
		utility.LogError(err, "get hypervisor failed", true)

		pt := common.PrettyItemTable{
			ShortFields: []common.Column{
				{Name: "Id"}, {Name: "HypervisorHostname"}, {Name: "HostIp"},
				{Name: "Status", AutoColor: true}, {Name: "State", AutoColor: true},
				{Name: "Type"}, {Name: "Version"},
				{Name: "Vcpus"}, {Name: "VcpusUsed"},
				{Name: "MemoryMB", Text: "Memory MB"},
				{Name: "MemoryMBUsed", Text: "Memory Used MB"},
				{Name: "ExtraResources", Slot: func(item interface{}) interface{} {
					p, _ := item.(nova.Hypervisor)
					return p.ExtraResourcesMarshal(true)
				}},
				{Name: "CpuInfoArch",
					Slot: func(item interface{}) interface{} {
						p, _ := item.(nova.Hypervisor)
						return p.CpuInfo.Arch
					}},
				{Name: "CpuInfoModel",
					Slot: func(item interface{}) interface{} {
						p, _ := item.(nova.Hypervisor)
						return p.CpuInfo.Model
					}},
				{Name: "CpuInfoVendor",
					Slot: func(item interface{}) interface{} {
						p, _ := item.(nova.Hypervisor)
						return p.CpuInfo.Vendor
					},
				},
				{Name: "CpuInfoFeature",
					Slot: func(item interface{}) interface{} {
						p, _ := item.(nova.Hypervisor)
						return p.CpuInfo.Features
					},
				},
				{Name: "NumaNodes", Slot: func(item interface{}) interface{} {
					p, _ := item.(nova.Hypervisor)
					if common.CONF.Format == common.FORMAT_TABLE_LIGHT {
						return p.NumaNodesBar()
					} else {
						return p.NumaNodesLine()
					}
				}},
			},
			Item: *hypervisor,
		}
		common.PrintPrettyItemTable(pt)
	},
}
var hypervisorUptime = &cobra.Command{
	Use:   "uptime <id or name>",
	Short: "uptime hypervisor",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := openstack.DefaultClient()
		hypervisor, err := client.NovaV2().Hypervisor().Find(args[0])
		utility.LogError(err, "get hypervisor failed", true)

		hypervisor, err = client.NovaV2().Hypervisor().Uptime(hypervisor.Id)
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
	hypervisorListFlags = flags.HypervisorListFlags{
		Name:        hypervisorList.Flags().StringP("name", "n", "", "Show hypervisors matched by name"),
		Type:        hypervisorList.Flags().StringP("type", "t", "", "Filte hypervisors by type"),
		WithServers: hypervisorList.Flags().Bool("with-servers", false, "List hypervisors with servers"),
		Long:        hypervisorList.Flags().BoolP("long", "l", false, "List additional fields in output"),
	}
	Hypervisor.AddCommand(hypervisorList, hypervisorShow, hypervisorUptime)
}
