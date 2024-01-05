package cli

import (
	"fmt"
	"os"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/keystoneauth"
)

func GetClient() *openstack.OpenstackClient {
	return GetClientWithRegion(common.CONF.Auth.Region.Name)
}
func GetClientWithRegion(regionName string) *openstack.OpenstackClient {
	user := keystoneauth.User{
		Name:     common.CONF.Auth.User.Name,
		Password: common.CONF.Auth.User.Password,
		Domain:   keystoneauth.Domain{Name: common.CONF.Auth.User.Domain.Name},
	}
	project := keystoneauth.Project{
		Name: common.CONF.Auth.Project.Name,
		Domain: keystoneauth.Domain{
			Name: common.CONF.Auth.Project.Domain.Name,
		},
	}

	client, _ := openstack.NewOpenstackClient(
		common.CONF.Auth.Url, user, project, regionName,
		common.CONF.Auth.TokenExpireTime,
	)

	if common.CONF.HttpTimeout > 0 {
		client.Identity.Auth.SetHttpTimeout(common.CONF.HttpTimeout)
	}
	if err := client.Identity.Auth.TokenIssue(); err != nil {
		logging.Fatal("获取 Token 失败, %s", err)
	}
	return client
}

func ExitIfError(err error) {
	if err == nil {
		return
	}
	fmt.Println(err)
	os.Exit(1)
}
