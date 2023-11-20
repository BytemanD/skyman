package cli

import (
	"fmt"
	"os"

	"github.com/BytemanD/easygo/pkg/global/logging"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/identity"
)

func getAuthClient() (*identity.V3AuthClient, error) {
	authClient, err := identity.GetV3AuthClient(
		common.CONF.Auth.Url, common.CONF.Auth.User,
		common.CONF.Auth.Project, common.CONF.Auth.RegionName.Name,
	)
	if err != nil {
		return nil, fmt.Errorf("获取认证客户端失败, %s", err)
	}
	if err := authClient.TokenIssue(); err != nil {
		return nil, fmt.Errorf("获取 Token 失败, %s", err)
	}
	return authClient, nil
}

func GetClient() *openstack.OpenstackClient {
	authClient, err := getAuthClient()
	if err != nil {
		logging.Fatal("get auth client failed %s", err)
	}
	client, err := openstack.GetClientWithAuthToken(authClient)
	if err == nil {
		client.Compute.UpdateVersion()
	}
	if err != nil {
		logging.Fatal("get openstack client failed %s", err)
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
