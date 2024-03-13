package image

import (
	"strings"

	"github.com/BytemanD/skyman/openstack/common"
	"github.com/BytemanD/skyman/openstack/identity"
	"github.com/BytemanD/skyman/openstack/keystoneauth"
	"github.com/BytemanD/skyman/utility"
)

type ImageClientV2 struct {
	identity.IdentityClientV3
	endpoint    string
	BaseHeaders map[string]string
}

func (client *ImageClientV2) Index() (*utility.Response, error) {
	return client.Request(common.NewIndexRequest(client.endpoint, nil, client.BaseHeaders))
}
func (client *ImageClientV2) GetCurrentVersion() (*identity.ApiVersion, error) {
	resp, err := client.Index()
	if err != nil {
		return nil, err
	}
	versions := map[string]identity.ApiVersions{"versions": {}}
	resp.BodyUnmarshal(&versions)
	return versions["versions"].Current(), nil
}

func GetImageClientV2(session identity.IdentityClientV3) (*ImageClientV2, error) {
	url, err := session.GetServiceEndpoint(keystoneauth.TYPE_IMAGE, "", keystoneauth.INTERFACE_PUBLIC)
	if err != nil {
		return nil, err
	}
	if !strings.Contains(url, "/v") {
		url += "/v2"
	}
	return &ImageClientV2{
		IdentityClientV3: session,
		endpoint:         url,
		BaseHeaders:      map[string]string{},
	}, nil
}
