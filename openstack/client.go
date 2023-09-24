package openstack

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/common"
	osCommon "github.com/BytemanD/skyman/openstack/common"
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

func (client OpenstackClient) WaitServerStatus(serverId string, status string, taskState string) (*compute.Server, error) {
	for {
		server, err := client.Compute.ServerShow(serverId)
		if err != nil {
			return nil, err
		}
		logging.Info("server %s status: %s", serverId, server.AllStatus())

		if strings.ToUpper(server.Status) == "ERROR" {
			return nil, fmt.Errorf("server %s status is ERROR", serverId)
		}
		if (status == "" || strings.ToUpper(server.Status) == strings.ToUpper(status)) &&
			(strings.ToUpper(server.TaskState) == strings.ToUpper(taskState)) {
			return server, nil
		}
		time.Sleep(time.Second * 2)
	}
}
func (client OpenstackClient) WaitServerCreated(serverId string) (*compute.Server, error) {
	return client.WaitServerStatus(serverId, "ACTIVE", "")
}

func (client OpenstackClient) WaitServerRebooted(serverId string) (*compute.Server, error) {
	return client.WaitServerStatus(serverId, "ACTIVE", "")
}

func (client OpenstackClient) WaitServerDeleted(serverId string) error {
	_, err := client.WaitServerStatus(serverId, "", "")
	if httpError, ok := err.(*osCommon.HttpError); ok {
		if httpError.Status == 404 {
			return nil
		}
	}
	return err
}
func (client OpenstackClient) WaitServerResized(serverId string, newFlavorName string) error {
	server, err := client.WaitServerStatus(serverId, "", "")
	if err != nil {
		return err
	}
	if server.Flavor.OriginalName == newFlavorName {
		return nil
	} else {
		return fmt.Errorf("server %s not resized", serverId)
	}
}
