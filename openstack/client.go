package openstack

import (
	"net/url"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack/compute"
	"github.com/BytemanD/skyman/openstack/identity"
	"github.com/BytemanD/skyman/openstack/image"
	"github.com/BytemanD/skyman/openstack/networking"
	"github.com/BytemanD/skyman/openstack/storage"
)

type OpenstackClient struct {
	AuthClient identity.V3AuthClient
	Identity   identity.IdentityClientV3
	Compute    compute.ComputeClientV2
	Image      image.ImageClientV2
	Storage    storage.StorageClientV2
	Networking networking.NeutronClientV2
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
	identityClient, err := identity.GetIdentityClientV3(*authClient)

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
	networkingClient, err := networking.GetNeutronClientV2(*authClient)
	if err != nil {
		return nil, err
	}
	return &OpenstackClient{
		AuthClient: *authClient,
		Identity:   *identityClient,
		Compute:    *computeClient,
		Image:      *imageClient,
		Storage:    *storageClient,
		Networking: *networkingClient,
	}, nil
}

func (client OpenstackClient) ServerInspect(serverId string, detail bool) (*compute.ServerInspect, error) {
	server, err := client.Compute.ServerShow(serverId)
	if err != nil {
		return nil, err
	}
	interfaceAttachmetns, err := client.Compute.ServerInterfaceList(serverId)
	if err != nil {
		return nil, err
	}
	volumeAttachments, err := client.Compute.ServerVolumeList(serverId)
	serverInspect := compute.ServerInspect{
		Server:          *server,
		Interfaces:      interfaceAttachmetns,
		Volumes:         volumeAttachments,
		InterfaceDetail: map[string]networking.Port{},
		VolumeDetail:    map[string]storage.Volume{},
	}
	if detail {
		portQuery := url.Values{}
		portQuery.Add("device_id", serverId)
		for _, port := range client.Networking.PortList(portQuery) {
			serverInspect.InterfaceDetail[port.Id] = port
		}

		for _, volume := range serverInspect.Volumes {
			vol, err := client.Storage.VolumeShow(volume.VolumeId)
			common.LogError(err, "get volume  failed", true)
			serverInspect.VolumeDetail[volume.VolumeId] = *vol
		}
	}
	return &serverInspect, nil
}
