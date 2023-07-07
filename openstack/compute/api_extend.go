package compute

import (
	"fmt"
	"net/url"

	"github.com/BytemanD/easygo/pkg/global/logging"
)

func (client ComputeClientV2) ServerPrune(query url.Values, yes bool, waitDeleted bool) {
	if len(query) == 0 {
		query.Set("status", "error")
	}
	logging.Info("查询虚拟机: %v", query)
	servers := client.ServerList(query)
	logging.Info("需要清理的虚拟机数量: %d\n", len(servers))
	if len(servers) == 0 {
		return
	}
	if !yes {
		var confirm string
		fmt.Println("即将删除虚拟机:")
		for _, server := range servers {
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
	logging.Info("开始删除虚拟机")
	for _, server := range servers {
		logging.Info("删除虚拟机 %s(%s)", server.Id, server.Name)
		client.ServerDelete(server.Id)
		if waitDeleted {
			client.WaitServerDeleted(server.Id)
		}
	}
}
