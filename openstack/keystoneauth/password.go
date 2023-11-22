package keystoneauth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/BytemanD/skyman/openstack/common"
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
)

type PasswordAuthPlugin struct {
	restfulClient common.RestfulClient

	AuthUrl           string
	Username          string
	Password          string
	ProjectName       string
	UserDomainName    string
	ProjectDomainName string
	RegionName        string
	TokenExpireSecond int
	tokenCache        TokenCache
}

func (plugin *PasswordAuthPlugin) SetTokenExpireSecond(expire int) {
	plugin.TokenExpireSecond = expire
}

func (plugin PasswordAuthPlugin) GetToken() (*Token, error) {
	if plugin.tokenCache.IsTokenExpired() {
		if err := plugin.TokenIssue(); err != nil {
			return nil, err
		}
	}
	return &plugin.tokenCache.token, nil
}
func (plugin *PasswordAuthPlugin) GetAuthTokenId() string {
	return plugin.tokenCache.TokenId
}

func (plugin *PasswordAuthPlugin) GetTokenId() (string, error) {
	if plugin.tokenCache.IsTokenExpired() {
		if err := plugin.TokenIssue(); err != nil {
			return "", err
		}
	}
	return plugin.tokenCache.TokenId, nil
}

func (client *PasswordAuthPlugin) getAuthReqBody() map[string]Auth {
	auth := Auth{}
	auth.Identity.Methods = []string{"password"}

	auth.Identity.Password.User.Name = client.Username
	auth.Identity.Password.User.Password = client.Password
	auth.Identity.Password.User.Domain.Name = client.UserDomainName
	auth.Scope.Project.Name = client.ProjectName
	auth.Scope.Project.Domain.Name = client.ProjectDomainName

	return map[string]Auth{"auth": auth}
}
func (plugin *PasswordAuthPlugin) Get(url string, query url.Values,
	headers map[string]string) (*common.Response, error) {
	return plugin.restfulClient.Get(url, query, headers)
}
func (plugin *PasswordAuthPlugin) Post(url string, body []byte,
	headers map[string]string) (*common.Response, error) {
	return plugin.restfulClient.Post(url, body, headers)
}
func (plugin *PasswordAuthPlugin) Put(url string, body []byte,
	headers map[string]string) (*common.Response, error) {
	return plugin.restfulClient.Put(url, body, headers)
}
func (plugin *PasswordAuthPlugin) Delete(url string,
	headers map[string]string) (*common.Response, error) {
	return plugin.restfulClient.Delete(url, headers)
}

func (plugin *PasswordAuthPlugin) Request(req *http.Request) (*common.Response, error) {
	return plugin.restfulClient.Request(req)
}

func (plugin *PasswordAuthPlugin) TokenIssue() error {
	body, _ := json.Marshal(plugin.getAuthReqBody())
	resp, err := plugin.Post(
		fmt.Sprintf("%s%s", plugin.AuthUrl, URL_AUTH_TOKEN), body, nil)
	if err != nil {
		return fmt.Errorf("token issue failed, %v", err)
	}
	var resToken RespToken
	resp.BodyUnmarshal(&resToken)

	plugin.tokenCache = TokenCache{
		token:     resToken.Token,
		TokenId:   resp.GetHeader("X-Subject-Token"),
		expiredAt: time.Now().Add(time.Second * time.Duration(3600)),
	}
	return nil
}

func (plugin *PasswordAuthPlugin) GetServiceEndpoint(
	serviceType string, serviceName string, serviceInterface string,
) (string, error) {
	if _, err := plugin.GetTokenId(); err != nil {
		return "", err
	}
	endpoints := plugin.tokenCache.GetServiceEndpoints(serviceType, serviceName)
	for _, endpoint := range endpoints {
		if endpoint.Interface == serviceInterface && endpoint.Region == plugin.RegionName {
			return endpoint.Url, nil
		}
	}
	return "", fmt.Errorf("endpoint not found")
}
func (plugin *PasswordAuthPlugin) SetHttpTimeout(timeout int) {
	plugin.restfulClient.Timeout = time.Second * time.Duration(timeout)
}
func NewPasswordAuth(authUrl string, user User, project Project, regionName string) PasswordAuthPlugin {
	return PasswordAuthPlugin{
		AuthUrl:           authUrl,
		Username:          user.Name,
		Password:          user.Password,
		UserDomainName:    user.Domain.Name,
		ProjectName:       project.Name,
		ProjectDomainName: project.Domain.Name,
		RegionName:        regionName,
	}
}
