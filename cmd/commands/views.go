package commands

import (
	"os"
	"strings"

	"github.com/BytemanD/stackcrud/openstack/compute"
	"github.com/BytemanD/stackcrud/openstack/image"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

type ResourceTable interface {
	Print()
}

type ServerTable struct {
	Server compute.Server
}

func (t ServerTable) Print() {
	header := table.Row{"Property", "Value"}

	tableWriter := table.NewWriter()
	tableWriter.AppendHeader(header)
	tableWriter.AppendRows([]table.Row{
		{"ID", t.Server.Id}, {"name", t.Server.Name},
		{"description", t.Server.Description},

		{"flavor:original_name", t.Server.Flavor.OriginalName},
		{"flavor:ram", t.Server.Flavor.Ram},
		{"flavor:vcpus", t.Server.Flavor.Vcpus},
		{"flavor:disk", t.Server.Flavor.Disk},
		{"flavor:swap", t.Server.Flavor.Swap},
		{"flavor:extra_specs", t.Server.GetFlavorExtraSpecsString()},

		{"image", t.Server.Image.Id},

		{"availability_zone  ", t.Server.AZ}, {"host", t.Server.Host},

		{"status", t.Server.Status}, {"task_state", t.Server.TaskState},
		{"power_state", t.Server.PowerState}, {"vm_state", t.Server.VmState},

		{"root_bdm_type", t.Server.RootBdmType},

		{"created", t.Server.Created}, {"updated", t.Server.Updated},
		{"terminated_at", t.Server.TerminatedAt}, {"launched_at", t.Server.LaunchedAt},

		{"user_id", t.Server.UserId},
		{"fault:code", t.Server.Fault.Code},
		{"fault:message", t.Server.Fault.Message},
		{"fault:details", t.Server.Fault.Details},
	})
	// tableWriter.SetStyle(table.StyleLight)
	tableWriter.Style().Format.Header = text.FormatDefault
	tableWriter.SetOutputMirror(os.Stdout)
	tableWriter.Render()
}

type ServersTable struct {
	Servers []compute.Server
}

func (t ServersTable) Print(long bool, verbose bool) {
	header := table.Row{
		"ID", "Name", "Status", "Task State", "Power State", "Networks",
	}
	var networksJoinSep string
	if long {
		networksJoinSep = "\n"
		if verbose {
			header = append(header, "Flavor:ram")
			header = append(header, "Flavor:vcpus")
		} else {
			header = append(header, "Flavor:Name")
		}
		header = append(header, "AZ")
		header = append(header, "Host")
		header = append(header, "Instance Name")
	} else {
		networksJoinSep = "; "
	}
	tableWriter := table.NewWriter()

	for _, server := range t.Servers {
		row := table.Row{
			server.Id, server.Name, server.Status,
			server.GetTaskState(), server.GetPowerState(),
			strings.Join(server.GetNetworks(), networksJoinSep),
		}
		if long {
			if verbose {
				row = append(row, server.Flavor.Ram, server.Flavor.Vcpus)
			} else {
				row = append(row, server.Flavor.OriginalName)
			}
			row = append(row, server.AZ, server.Host, server.InstanceName)
		}
		tableWriter.SortBy([]table.SortBy{
			{Name: "Name", Mode: table.Asc},
		})
		tableWriter.AppendRow(row)
	}

	// tableWriter.SetStyle(table.StyleLight)
	tableWriter.AppendHeader(header)
	tableWriter.Style().Format.Header = text.FormatDefault
	tableWriter.SetOutputMirror(os.Stdout)
	tableWriter.Render()

}

type ImagesTable struct {
	Images []image.Image
}

func (t ImagesTable) Print(long bool, verbose bool) {
	header := table.Row{
		"ID", "Name", "Disk Format", "Container Format",
		"Size", "Status",
	}
	tableWriter := table.NewWriter()
	for _, image := range t.Images {
		row := table.Row{image.ID, image.Name, image.DiskFormat,
			image.ContainerFormat, image.Size, image.Status}
		tableWriter.SortBy([]table.SortBy{
			{Name: "Name", Mode: table.Asc},
		})
		tableWriter.AppendRow(row)
	}

	// tableWriter.SetStyle(table.StyleLight)
	tableWriter.AppendHeader(header)
	tableWriter.Style().Format.Header = text.FormatDefault
	tableWriter.SetOutputMirror(os.Stdout)
	tableWriter.Render()
}
