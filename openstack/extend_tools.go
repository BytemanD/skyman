package openstack

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/BytemanD/easygo/pkg/syncutils"
	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/openstack/model/cinder"
	"github.com/BytemanD/skyman/openstack/model/neutron"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/utility"
	"github.com/samber/lo"
)

func (o Openstack) PruneServers(query url.Values, yes bool, waitDeleted bool) {
	c := o.NovaV2()
	console.Info("查询虚拟机: %v", query.Encode())
	servers, err := c.ListServer(query, true)
	if query.Get("host") == "" {
		console.Info("过滤虚拟机: No valid host was found")
		servers = lo.Filter(servers, func(item nova.Server, _ int) bool {
			if item.Host == "" && strings.Contains(item.Fault.Message, "No valid host was found") {
				return true
			}
			if item.Fault.Message == "" {
				if server, err := c.GetServer(item.Id); err == nil {
					return server.Host == "" && strings.Contains(server.Fault.Message, "No valid host was found")
				}
			}
			return false
		})
	}
	utility.LogError(err, "query servers failed", true)
	console.Info("需要清理的虚拟机数量: %d\n", len(servers))
	if len(servers) == 0 {
		return
	}
	if !yes {
		for _, server := range servers {
			fmt.Printf("    %s (name=%s, status=%s, host=%s), fault message: %s\n",
				server.Id, server.Name, server.Status, server.Host, server.Fault.Message)
		}
		fmt.Printf("即将删除 %d 个虚拟机\n", len(servers))
		yes = utility.DefaultScanComfirm("是否删除")
	}
	if !yes {
		return
	}
	console.Info("开始删除虚拟机")
	tg := syncutils.TaskGroup[nova.Server]{
		Items: servers,
		Title: fmt.Sprintf("delete %d server(s)", len(servers)),
		Func: func(server nova.Server) error {
			console.Info("delete %s", server.Id)
			err := c.DeleteServer(server.Id)
			if err != nil {
				return fmt.Errorf("delete %s failed: %v", server.Id, err)
			}
			if waitDeleted {
				c.WaitServerDeleted(server.Id)
			}
			return nil
		},
		ShowProgress: true,
	}
	tg.Start()
}
func (o Openstack) PruneVolumes(query url.Values, matchName string, volumeType string,
	yes bool) {
	c := o.CinderV2()
	if query == nil {
		query = url.Values{}
	}
	if query.Get("status") == "" {
		query.Add("status", "error")
	}
	console.Info("查询卷: %s", query.Encode())
	volumes, err := c.ListVolume(query, true)
	filterdVolumes := []cinder.Volume{}
	for _, vol := range volumes {
		if volumeType != "" && vol.VolumeType != volumeType {
			continue
		}
		if matchName != "" && !strings.Contains(vol.Name, matchName) {
			continue
		}
		filterdVolumes = append(filterdVolumes, vol)
	}
	volumes = filterdVolumes

	if err != nil {
		console.Error("get volumes failed, %s", err)
		return
	}
	console.Info("需要清理的卷数量: %d\n", len(volumes))
	if len(volumes) == 0 {
		return
	}
	console.Info("Last volume id: %s", volumes[len(volumes)-1].Id)
	if !yes {
		for _, volume := range volumes {
			fmt.Printf("%s 名称: %s\t创建时间: %s\n", volume.Id, volume.Name, volume.CreatedAt)
		}
		fmt.Printf("即将清理 %d 个卷:\n", len(volumes))
		yes = utility.DefaultScanComfirm("是否删除?")
		if !yes {
			return
		}
	}
	console.Info("开始清理")
	tg := syncutils.TaskGroup[cinder.Volume]{
		Items:        volumes,
		Title:        fmt.Sprintf("delete %d server(s)", len(volumes)),
		ShowProgress: true,
		Func: func(volume cinder.Volume) error {
			console.Debug("delete volume %s(%s)", volume.Id, volume.Name)
			err := c.DeleteVolume(volume.Id, true, true)
			if err != nil {
				return fmt.Errorf("delete volume %s failed: %v", volume.Id, err)
			}
			return nil
		},
	}
	err = tg.Start()
	if err != nil {
		console.Error("清理失败: %v", err)
	} else {
		console.Info("清理完成")
	}
}
func (o Openstack) PrunePorts(ports []neutron.Port) {
	c := o.NeutronV2()
	for _, port := range ports {
		fmt.Printf("%s (%s)\n", port.Id, port.Name)
	}
	fmt.Printf("即将清理 %d 个Port(s)\n", len(ports))
	yes := utility.DefaultScanComfirm("是否清理?")
	if !yes {
		return
	}
	tg := syncutils.TaskGroup[neutron.Port]{
		Title:        fmt.Sprintf("delete %d port(s)", len(ports)),
		Items:        ports,
		ShowProgress: true,
		Func: func(port neutron.Port) error {
			console.Debug("delete port %s(%s)", port.Id, port.Name)
			err := c.DeletePort(port.Id)
			if err != nil {
				return fmt.Errorf("delete port %s failed: %v", port.Id, err)
			}
			return nil
		},
	}
	err := tg.Start()
	if err != nil {
		console.Error("清理失败: %v", err)
	} else {
		console.Info("清理完成")
	}
}
