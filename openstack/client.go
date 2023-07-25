package openstack

import (
	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/stackcrud/common"
	"github.com/BytemanD/stackcrud/openstack/compute"
	"github.com/BytemanD/stackcrud/openstack/identity"
	"github.com/BytemanD/stackcrud/openstack/image"
	"github.com/BytemanD/stackcrud/openstack/storage"
)

type OpenstackClient struct {
	Identity identity.V3AuthClient
	Compute  compute.ComputeClientV2
	Image    image.ImageClientV2
	Storage  storage.StorageClientV2
}

func getAuthClient() (*identity.V3AuthClient, error) {
	authClient, err := identity.GetV3AuthClient(
		common.CONF.Auth.Url, common.CONF.Auth.User,
		common.CONF.Auth.Project, common.CONF.Auth.RegionName,
	)
	if err != nil {
		return nil, err
	}
	if err := authClient.TokenIssue(); err != nil {
		logging.Fatal("获取 Token 失败, %s", err)
	}
	return authClient, nil
}

func GetClient(authUrl string, user identity.User, project identity.Project, regionName string,
) (*OpenstackClient, error) {
	authClient, err := identity.GetV3AuthClient(authUrl, user, project, regionName)
	if err != nil {
		return nil, err
	}
	return GetClientWithAuthToken(authClient)
}

func GetClientWithAuthToken(authClient *identity.V3AuthClient) (*OpenstackClient, error) {
	computeClient, err := compute.GetComputeClientV2(*authClient)
	if err != nil {
		return nil, err
	}
	imageClient, err := image.GetImageClientV2(*authClient)
	if err != nil {
		return nil, err
	}
	storageClient, err := storage.GetStorageClientV2(*authClient)
	if err != nil {
		return nil, err
	}
	return &OpenstackClient{
		Identity: *authClient,
		Compute:  *computeClient,
		Image:    *imageClient,
		Storage:  *storageClient,
	}, nil
}
