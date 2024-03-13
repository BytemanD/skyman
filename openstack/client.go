package openstack

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/openstack/compute"
	"github.com/BytemanD/skyman/openstack/identity"
	"github.com/BytemanD/skyman/openstack/image"
	"github.com/BytemanD/skyman/openstack/keystoneauth"
	"github.com/BytemanD/skyman/openstack/networking"
	"github.com/BytemanD/skyman/openstack/storage"
	"github.com/BytemanD/skyman/utility"
)

type OpenstackClient struct {
	Identity   *identity.IdentityClientV3
	Compute    *compute.ComputeClientV2
	Image      *image.ImageClientV2
	Storage    *storage.StorageClientV2
	Networking *networking.NeutronClientV2
}

func (c *OpenstackClient) ComputeClient() *compute.ComputeClientV2 {
	if c.Compute == nil {
		client, err := compute.GetComputeClientV2(*c.Identity)
		if err != nil {
			logging.Error("get compute client failed, %s", err)
			return nil
		}
		c.Compute = client
		c.Compute.UpdateVersion()
	}
	return c.Compute
}

func (c *OpenstackClient) MustGenerateComputeClient() *compute.ComputeClientV2 {
	client := c.ComputeClient()
	if client == nil {
		fmt.Println("can not get compute client")
		os.Exit(1)
	}
	return client
}

func (c *OpenstackClient) ImageClient() *image.ImageClientV2 {
	if c.Image == nil {
		c.Image, _ = image.GetImageClientV2(*c.Identity)
	}
	return c.Image
}

func (c *OpenstackClient) StorageClient() *storage.StorageClientV2 {
	if c.Storage == nil {
		c.Storage, _ = storage.GetStorageClientV2(*c.Identity)
	}
	return c.Storage
}

func (c *OpenstackClient) NetworkingClient() *networking.NeutronClientV2 {
	if c.Networking == nil {
		c.Networking, _ = networking.GetNeutronClientV2(*c.Identity)
	}
	return c.Networking
}

func NewOpenstackClient(authUrl string, user keystoneauth.User, project keystoneauth.Project,
	regionName string, tokenExpireSecond int,
) (*OpenstackClient, error) {
	if authUrl == "" {
		return nil, fmt.Errorf("authUrl is missing")
	}
	passwordAuth := keystoneauth.NewPasswordAuth(authUrl, user, project, regionName)
	// passwordAuth.TokenIssue()
	passwordAuth.SetTokenExpireSecond(tokenExpireSecond)
	return GetClientWithAuthToken(&passwordAuth), nil
}

func GetClientWithAuthToken(passwordAuth *keystoneauth.PasswordAuthPlugin) *OpenstackClient {
	identityClient := identity.GetIdentityClientV3(*passwordAuth)

	return &OpenstackClient{Identity: identityClient}
}

func (client OpenstackClient) ServerInspect(serverId string) (*ServerInspect, error) {
	server, err := client.ComputeClient().ServerShow(serverId)
	if err != nil {
		return nil, err
	}
	interfaceAttachmetns, err := client.ComputeClient().ServerInterfaceList(serverId)
	if err != nil {
		return nil, err
	}
	volumeAttachments, err := client.ComputeClient().ServerVolumeList(serverId)
	actions, err := client.ComputeClient().ServerActionList(serverId)
	if err != nil {
		return nil, err
	}
	serverInspect := ServerInspect{
		Server:          *server,
		Interfaces:      interfaceAttachmetns,
		Volumes:         volumeAttachments,
		InterfaceDetail: map[string]networking.Port{},
		VolumeDetail:    map[string]storage.Volume{},
		Actions:         actions,
	}

	portQuery := url.Values{}
	ports, err := client.NetworkingClient().PortList(portQuery)
	if err != nil {
		return nil, err
	}
	for _, port := range ports {
		serverInspect.InterfaceDetail[port.Id] = port
	}

	for _, volume := range serverInspect.Volumes {
		vol, err := client.StorageClient().VolumeShow(volume.VolumeId)
		utility.LogError(err, "get volume failed", true)
		serverInspect.VolumeDetail[volume.VolumeId] = *vol
	}
	return &serverInspect, nil
}

func (client OpenstackClient) WaitServerStatus(serverId string, status string, taskState string) (*compute.Server, error) {
	for {
		server, err := client.ComputeClient().ServerShow(serverId)
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
	if httpError, ok := err.(*utility.HttpError); ok {
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
