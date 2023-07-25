package storage

import (
	"fmt"
	"net/url"

	"github.com/BytemanD/easygo/pkg/global/logging"
)

func (client StorageClientV2) VolumePrune(query url.Values, yes bool, waitDeleted bool) {
	if len(query) == 0 {
		query.Set("status", "error")
	}
	logging.Info("查询卷: %v", query)
	volumes := client.VolumeList(query)
	logging.Info("需要清理的卷数量: %d\n", len(volumes))
	if len(volumes) == 0 {
		return
	}
	if !yes {
		var confirm string
		fmt.Println("即将删除卷:")
		for _, server := range volumes {
			fmt.Printf("%s (%s)\n", server.Id, server.Name)
		}
		for {
			fmt.Printf("是否删除[yes/no]: ")
			fmt.Scanf("%s %d %f", &confirm)
			if confirm == "yes" || confirm == "y" {
				break
			} else if confirm == "no" || confirm == "n" {
				return
			} else {
				fmt.Print("输入错误, 请重新输入!")
			}
		}
	}
	logging.Info("开始删除卷")
	for _, volume := range volumes {
		logging.Info("删除卷 %s(%s)", volume.Id, volume.Name)
		client.VolumeDelete(volume.Id)
	}
}
