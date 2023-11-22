package networking

import (
	"strings"

	"github.com/BytemanD/skyman/openstack/identity"
)

type NeutronClientV2 struct {
	identity.IdentityClientV3
	endpoint    string
	BaseHeaders map[string]string
}

func GetNeutronClientV2(authClient identity.IdentityClientV3) (*NeutronClientV2, error) {
	endpoint, err := authClient.GetServiceEndpoint(
		identity.TYPE_NETWORK, "", identity.INTERFACE_PUBLIC)

	if err != nil {
		return nil, err
	}
	if !strings.Contains(endpoint, "/v2") {
		endpoint += "/v2.0"
	}
	return &NeutronClientV2{
		IdentityClientV3: authClient,
		endpoint:         endpoint,
		BaseHeaders:      map[string]string{},
	}, nil
}
