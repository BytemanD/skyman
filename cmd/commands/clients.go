package commands

import (
	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/stackcrud/common"
	"github.com/BytemanD/stackcrud/openstack/compute"
	"github.com/BytemanD/stackcrud/openstack/identity"
	"github.com/BytemanD/stackcrud/openstack/image"
)

func getAuthClient() identity.V3AuthClient {
	authClient, err := identity.GetV3Client(
		common.CONF.Auth.Url, common.CONF.Auth.User,
		common.CONF.Auth.Project, common.CONF.Auth.RegionName,
	)
	if err != nil {
		logging.Fatal("获取认证客户端失败, %s", err)
	}
	if err := authClient.TokenIssue(); err != nil {
		logging.Fatal("获取 Token 失败, %s", err)
	}
	return authClient
}

func getComputeClient() compute.ComputeClientV2 {
	authClient := getAuthClient()
	computeClient, err := compute.GetComputeClientV2(authClient)
	if err != nil {
		logging.Fatal("获取计算客户端失败, %s", err)
	}
	computeClient.UpdateVersion()
	return computeClient
}

func getImageClient() image.ImageClientV2 {
	authClient := getAuthClient()
	client, err := image.GetImageClientV2(authClient)
	if err != nil {
		logging.Fatal("获取计算客户端失败, %s", err)
	}
	client.UpdateVersion()
	return client
}
