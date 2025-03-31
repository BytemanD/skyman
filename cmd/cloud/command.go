package cloud

import (
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/common/datatable"
	"github.com/BytemanD/skyman/openstack"
	"github.com/fatih/color"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var CloudsCmd = &cobra.Command{
	Use:   "clouds",
	Short: "Show clouds",
	Run: func(cmd *cobra.Command, args []string) {
		println("配置文件:", lo.CoalesceOrEmpty(viper.ConfigFileUsed(), "无"))

		type CloudView struct {
			openstack.Cloud
			Name string
		}
		items := lo.MapToSlice(openstack.CONF.Clouds, func(k string, v openstack.Cloud) CloudView {
			return CloudView{v, k}
		})
		println("云环境:")
		common.PrintItems(
			[]datatable.Column[CloudView]{
				{Name: "Name"},
				{Name: "AuthUrl", RenderFunc: func(item CloudView) any { return item.Auth.AuthUrl }},
				{Name: "RegionName"},
				{Name: "ProjectName", RenderFunc: func(item CloudView) any { return item.Auth.ProjectName }},
				{Name: "Username", RenderFunc: func(item CloudView) any { return item.Auth.Username }},
			},
			[]datatable.Column[CloudView]{},
			items,
			common.TableOptions{},
		)
		if cloudName := openstack.CloudName(); cloudName != "" {
			if lo.HasKey(openstack.CONF.Clouds, cloudName) {
				println(color.CyanString("使用云环境: %s", cloudName))
			} else {
				println(color.RedString("云环境 '%s' 不存在, 请检查配置或者环境变量", cloudName))
			}
		} else {
			println(color.YellowString("云名称未配置, 使用环境变量"))
		}
	},
}
