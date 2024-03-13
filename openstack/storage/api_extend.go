package storage

import (
	"fmt"
	"net/url"

	"github.com/BytemanD/skyman/utility"

	"github.com/BytemanD/easygo/pkg/global/logging"
)

func (client StorageClientV2) VolumePrune(query url.Values, yes bool) {
	if query == nil {
		query = url.Values{}
	}
	if len(query) == 0 {
		query.Add("status", "error")
	}
	logging.Info("查询卷: %v", query)
	volumes, err := client.VolumeList(query)
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
		yes = utility.ScanfComfirm("是否删除?", []string{"yes", "y"}, []string{"no", "n"})
		if !yes {
			return
		}
	}
	logging.Info("开始清理")
	tg := utility.TaskGroup{
		Func: func(i interface{}) error {
			volume := i.(Volume)
			logging.Debug("delete volume %s(%s)", volume.Id, volume.Name)
			err := client.VolumeDelete(volume.Id)
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
