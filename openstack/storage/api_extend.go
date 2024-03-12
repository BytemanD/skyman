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
	utility.GoroutineMap(
		func(i interface{}) {
			volume := i.(Volume)
			logging.Info("删除卷 %s(%s)", volume.Id, volume.Name)
			client.VolumeDelete(volume.Id)
		}, volumes,
	)
	logging.Info("开始完成")
}
