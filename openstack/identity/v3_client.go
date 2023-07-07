package identity

import (
	"encoding/json"
	"fmt"
	netUrl "net/url"
	"time"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/stackcrud/openstack"
)

const (
	ContentType string = "application/json"

	TYPE_COMPUTE  string = "compute"
	TYPE_IDENTITY string = "identity"
	TYPE_IMAGE    string = "image"

	INTERFACE_PUBLIC   string = "public"
	INTERFACE_ADMIN    string = "admin"
	INTERFACE_INTERVAL string = "internal"

	URL_AUTH_TOKEN string = "/auth/tokens"
)

type V3AuthClient struct {
	AuthUrl           string
	Username          string
	Password          string
	ProjectName       string
	UserDomainName    string
	ProjectDomainName string
	TokenExpireSecond int
	RegionName        string
	token             Token
	expiredAt         time.Time

	session openstack.Session
}

func (client *V3AuthClient) TokenIssue() error {
	authBody := GetAuthReqBody(client.Username, client.Password, client.ProjectName)

	body, _ := json.Marshal(authBody)

	url := fmt.Sprintf("%s%s", client.AuthUrl, URL_AUTH_TOKEN)
	resp, err := client.session.Post(url, body, nil)
	if err != nil {
		logging.Error("token issue failed, %s", err)
		return err
	}
	var resToken RespToken
	json.Unmarshal(resp.Body, &resToken)
	resToken.Token.tokenId = resp.GetHeader("X-Subject-Token")
	client.token = resToken.Token
	client.expiredAt = time.Now().Add(time.Second * time.Duration(client.TokenExpireSecond))
	return nil
}
func (client *V3AuthClient) isTokenExpired() bool {
	if client.token.tokenId == "" {
		return true
	}
	if client.expiredAt.Before(time.Now()) {
		logging.Debug("token is exipred, expire second is %d", client.TokenExpireSecond)
		return true
	}
	return false
}

func (client *V3AuthClient) getTokenId() string {
	if client.isTokenExpired() {
		client.TokenIssue()
	}
	return client.token.tokenId
}

func (client *V3AuthClient) rejectToken(headers map[string]string) map[string]string {
	if headers == nil {
		headers = map[string]string{}
	}
	if headers["X-Auth-Token"] == "" {
		headers["X-Auth-Token"] = client.getTokenId()
	}
	return headers
}
func (client *V3AuthClient) Get(url string, query netUrl.Values, headers map[string]string) (*openstack.Response, error) {
	headers = client.rejectToken(headers)
	return client.session.Get(url, query, headers)
}
func (client *V3AuthClient) Post(url string, body []byte, headers map[string]string) (*openstack.Response, error) {
	headers = client.rejectToken(headers)
	return client.session.Post(url, body, headers)
}
func (client *V3AuthClient) Delete(url string, headers map[string]string) (*openstack.Response, error) {
	headers = client.rejectToken(headers)
	return client.session.Delete(url, headers)
}

func (client *V3AuthClient) ServiceList() (*openstack.Response, error) {
	url := fmt.Sprintf("%s%s", client.AuthUrl, "/services")
	return client.Get(url, nil, map[string]string{})
}

func (client *V3AuthClient) UserList() (*openstack.Response, error) {
	url := fmt.Sprintf("%s%s", client.AuthUrl, "/users")
	return client.Get(url, nil, map[string]string{})
}

func (client *V3AuthClient) GetEndpointFromCatalog(serviceType string, endpointInterface string, region string) (string, error) {
	if len(client.token.Catalogs) == 0 {
		if err := client.TokenIssue(); err != nil {
			return "", err
		}
	}
	endpoints := client.token.GetEndpoints(OptionCatalog{
		Type:      serviceType,
		Interface: endpointInterface,
		Region:    region,
	})
	if (len(endpoints)) == 0 {
		return "", fmt.Errorf("endpoints not found")
	} else {
		return endpoints[0].Url, nil
	}
}
