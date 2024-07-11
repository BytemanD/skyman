package openstack

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/easygo/pkg/stringutils"
	"github.com/BytemanD/easygo/pkg/syncutils"
	"github.com/BytemanD/skyman/openstack/model/cinder"
	"github.com/BytemanD/skyman/openstack/model/neutron"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/utility"
)

func (o Openstack) PruneServers(query url.Values, yes bool, waitDeleted bool) {
	c := o.NovaV2()
	if len(query) == 0 {
		query.Set("status", "error")
	}
	logging.Info("查询虚拟机: %v", query.Encode())
	servers, err := c.Servers().List(query)
	utility.LogError(err, "query servers failed", true)
	logging.Info("需要清理的虚拟机数量: %d\n", len(servers))
	if len(servers) == 0 {
		return
	}
	if !yes {
		fmt.Printf("即将删除 %d 个虚拟机:\n", len(servers))
		for _, server := range servers {
			fmt.Printf("    %s(%s)\n", server.Id, server.Name)
		}
		yes = stringutils.ScanfComfirm("是否删除", []string{"yes", "y"}, []string{"no", "n"})
	}
	if !yes {
		return
	}
	logging.Info("开始删除虚拟机")
	tg := syncutils.TaskGroup{
		Items: servers,
		Func: func(o interface{}) error {
			s := o.(nova.Server)
			logging.Info("delete %s", s.Id)
			err := c.Servers().Delete(s.Id)
			if err != nil {
				return fmt.Errorf("delete %s failed: %v", s.Id, err)
			}
			if waitDeleted {
				c.Servers().WaitDeleted(s.Id)
			}
			return nil
		},
		ShowProgress: true,
	}
	tg.Start()
}
func (o Openstack) PruneVolumes(query url.Values, matchName string, yes bool) {
	c := o.CinderV2()
	if query == nil {
		query = url.Values{}
	}
	if len(query) == 0 {
		query.Add("status", "error")
	}
	logging.Info("查询卷: %v", query)
	volumes, err := c.Volumes().List(query)
	if matchName != "" {
		filterdVolumes := []cinder.Volume{}
		for _, vol := range volumes {
			if strings.Contains(vol.Name, matchName) {
				filterdVolumes = append(filterdVolumes, vol)
			}
		}
		volumes = filterdVolumes
	}
	if err != nil {
		logging.Error("get volumes failed, %s", err)
		return
	}
	logging.Info("需要清理的卷数量: %d\n", len(volumes))
	if len(volumes) == 0 {
		return
	}
	if !yes {
		fmt.Printf("即将清理 %d 个卷:\n", len(volumes))
		for _, server := range volumes {
			fmt.Printf("%s (%s)\n", server.Id, server.Name)
		}
		yes = stringutils.ScanfComfirm("是否删除?", []string{"yes", "y"}, []string{"no", "n"})
		if !yes {
			return
		}
	}
	logging.Info("开始清理")
	tg := syncutils.TaskGroup{
		Func: func(i interface{}) error {
			volume := i.(cinder.Volume)
			logging.Debug("delete volume %s(%s)", volume.Id, volume.Name)
			err := c.Volumes().Delete(volume.Id, true, true)
			if err != nil {
				return fmt.Errorf("delete volume %s failed: %v", volume.Id, err)
			}
			return nil
		},
		Items:        volumes,
		ShowProgress: true,
	}
	err = tg.Start()
	if err != nil {
		logging.Error("清理失败: %v", err)
	} else {
		logging.Info("清理完成")
	}
}
func (o Openstack) PrunePorts(ports []neutron.Port) {
	c := o.NeutronV2()
	for _, port := range ports {
		fmt.Printf("%s (%s)\n", port.Id, port.Name)
	}
	fmt.Printf("即将清理 %d 个Port(s)\n", len(ports))
	yes := stringutils.ScanfComfirm("是否清理?", []string{"yes", "y"}, []string{"no", "n"})
	if !yes {
		return
	}
	tg := syncutils.TaskGroup{
		Func: func(i interface{}) error {
			port := i.(neutron.Port)
			logging.Debug("delete port %s(%s)", port.Id, port.Name)
			err := c.Ports().Delete(port.Id)
			if err != nil {
				return fmt.Errorf("delete port %s failed: %v", port.Id, err)
			}
			return nil
		},
		Items:        ports,
		ShowProgress: true,
	}
	err := tg.Start()
	if err != nil {
		logging.Error("清理失败: %v", err)
	} else {
		logging.Info("清理完成")
	}
}
