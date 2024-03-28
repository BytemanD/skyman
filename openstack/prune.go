package openstack

import (
	"fmt"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/easygo/pkg/stringutils"
	"github.com/BytemanD/easygo/pkg/syncutils"
	"github.com/BytemanD/skyman/openstack/model/neutron"
)

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
