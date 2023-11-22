// Openstack 认证客户端

package identity

import (
	"bytes"
	"net/http"
	"net/url"
	"strings"

	"github.com/BytemanD/skyman/openstack/common"
	"github.com/BytemanD/skyman/openstack/keystoneauth"
)

const (
	ContentType string = "application/json"

	TYPE_COMPUTE   string = "compute"
	TYPE_VOLUME    string = "volume"
	TYPE_VOLUME_V2 string = "volumev2"
	TYPE_VOLUME_V3 string = "volumev3"
	TYPE_IDENTITY  string = "identity"
	TYPE_IMAGE     string = "image"
	TYPE_NETWORK   string = "network"

	INTERFACE_PUBLIC   string = "public"
	INTERFACE_ADMIN    string = "admin"
	INTERFACE_INTERVAL string = "internal"

	URL_AUTH_TOKEN string = "/auth/tokens"

	DEFAULT_TOKEN_EXPIRE_SECOND = 3600
	AUTH_TOKEN_HEADER           = "X-Auth-Token"
)

type IdentityClientV3 struct {
	BaseHeaders map[string]string
	Auth        *keystoneauth.PasswordAuthPlugin
	endpoint    string
}

func (client IdentityClientV3) rejectToken(req *http.Request) error {
	tokenId, err := client.Auth.GetTokenId()
	if err != nil {
		return err
	}
	req.Header.Set(AUTH_TOKEN_HEADER, tokenId)
	return nil
}
func (client IdentityClientV3) RequestWithoutToken(restfulReq common.RestfulRequest) (*common.Response, error) {
	return client.request(restfulReq, false)
}
func (client IdentityClientV3) Request(restfulReq common.RestfulRequest) (*common.Response, error) {
	return client.request(restfulReq, true)
}
func (client IdentityClientV3) request(restfulReq common.RestfulRequest, rejectToken bool) (*common.Response, error) {
	url, err := restfulReq.Url()
	if err != nil {
		return nil, err
	}
	if restfulReq.Body != nil {
		restfulReq.Body = []byte{}
	}
	body := bytes.NewBuffer(restfulReq.Body)
	if restfulReq.Method == "" {
		restfulReq.Method = "GET"
	}
	req, err := http.NewRequest(restfulReq.Method, url, body)
	if err != nil {
		return nil, err
	}
	if restfulReq.Headers != nil {
		for k, v := range restfulReq.Headers {
			req.Header.Set(k, v)
		}
	}
	if rejectToken {
		if err := client.rejectToken(req); err != nil {
			return nil, err
		}
	}
	req.URL.RawQuery = restfulReq.Query.Encode()
	return client.Auth.Request(req)
}
func (client IdentityClientV3) FoundByIdOrName(resource string, idOrName string,
	headers map[string]string,
	foundById func(resp common.BaseResponse),
	foundByName func(resp common.BaseResponse),
) (*common.Response, error) {
	req := client.newRequest(resource, idOrName, nil, nil)
	resp, err := client.Request(req)
	if err != nil {
		if resp.IsNotFound() {
			query := url.Values{}
			query.Set("name", idOrName)
			req = client.newRequest(resource, "", query, nil)
			resp, err = client.Request(req)
		}
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return resp, err
}

func (client IdentityClientV3) Index() (*common.Response, error) {
	req := common.NewIndexRequest(client.endpoint, nil, client.BaseHeaders)
	return client.RequestWithoutToken(req)
}

func (client IdentityClientV3) GetCurrentVersion() (*ApiVersion, error) {
	resp, err := client.Index()
	if err != nil {
		return nil, err
	}
	versions := map[string]ApiVersions{"versions": {}}
	resp.BodyUnmarshal(&versions)
	return versions["versions"].Current(), nil
}
func (client IdentityClientV3) GetStableVersion() (*ApiVersion, error) {
	resp, err := client.Index()
	if err != nil {
		return nil, err
	}
	type apiVersion struct {
		Values ApiVersions `json:"values"`
	}
	versions := map[string]apiVersion{"versions": {}}
	resp.BodyUnmarshal(&versions)
	return versions["versions"].Values.Stable(), nil
}

func GetIdentityClientV3(auth keystoneauth.PasswordAuthPlugin) *IdentityClientV3 {
	endpoint := auth.AuthUrl
	if !strings.HasSuffix(endpoint, "/v3") {
		endpoint += "/v3"
	}
	return &IdentityClientV3{
		endpoint:    endpoint,
		Auth:        &auth,
		BaseHeaders: map[string]string{"Content-Type": "application/json"},
	}
}
