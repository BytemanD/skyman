package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/BytemanD/skyman/utility/httpclient"
	"github.com/go-resty/resty/v2"
)

const (
	CONTENT_TYPE string = "application/json"

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
	session                *httpclient.RESTClient
	AuthUrl                string
	Username               string
	Password               string
	ProjectName            string
	UserDomainName         string
	ProjectDomainName      string
	RegionName             string
	LocalTokenExpireSecond int
	tokenCache             TokenCache
}

func (plugin PasswordAuthPlugin) Region() string {
	return plugin.RegionName
}

func (plugin *PasswordAuthPlugin) SetLocalTokenExpire(expireSeconds int) {
	plugin.LocalTokenExpireSecond = expireSeconds
}

func (plugin PasswordAuthPlugin) GetToken() (*Token, error) {
	if plugin.tokenCache.IsTokenExpired() {
		if err := plugin.TokenIssue(); err != nil {
			return nil, err
		}
	}
	return &plugin.tokenCache.token, nil
}

func (plugin *PasswordAuthPlugin) GetTokenId() (string, error) {
	if plugin.tokenCache.IsTokenExpired() {
		if err := plugin.TokenIssue(); err != nil {
			return "", err
		}
	}
	return plugin.tokenCache.TokenId, nil
}

func (client PasswordAuthPlugin) getAuthReqBody() map[string]Auth {
	auth := Auth{}
	auth.Identity.Methods = []string{"password"}

	auth.Identity.Password.User.Name = client.Username
	auth.Identity.Password.User.Password = client.Password
	auth.Identity.Password.User.Domain.Name = client.UserDomainName
	auth.Scope.Project.Name = client.ProjectName
	auth.Scope.Project.Domain.Name = client.ProjectDomainName

	return map[string]Auth{"auth": auth}
}

func (plugin *PasswordAuthPlugin) TokenIssue() error {
	body := plugin.getAuthReqBody()
	resp, err := plugin.session.Post(
		fmt.Sprintf("%s%s", plugin.AuthUrl, URL_AUTH_TOKEN), body, nil,
	)
	if err != nil {
		return fmt.Errorf("token issue failed, %v", err)
	}
	var respToken RespToken
	if err := json.Unmarshal(resp.Body(), &respToken); err != nil {
		return err
	}
	plugin.tokenCache = TokenCache{
		token:     respToken.Token,
		TokenId:   resp.Header().Get("X-Subject-Token"),
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
	return "", fmt.Errorf("endpoint %s:%s:%s for region '%s' not found",
		serviceType, serviceName, serviceInterface,
		plugin.RegionName)
}
func (plugin PasswordAuthPlugin) SetHttpTimeout(timeout int) *PasswordAuthPlugin {
	plugin.session.Timeout = time.Second * time.Duration(timeout)
	return &plugin
}

func (plugin PasswordAuthPlugin) AuthRequest(req *resty.Request) error {
	tokenId, err := plugin.GetTokenId()
	if err != nil {
		return err
	}
	req.Header.Set("X-Auth-Token", tokenId)
	return nil
}
func (plugin PasswordAuthPlugin) GetSafeHeader(header http.Header) http.Header {
	safeHeaders := http.Header{}
	for k, v := range header {
		if k == "X-Auth-Token" {
			safeHeaders[k] = []string{"<TOKEN>"}
		} else {
			safeHeaders[k] = v
		}
	}
	return safeHeaders
}
func (plugin PasswordAuthPlugin) GetProjectId() (string, error) {
	_, err := plugin.GetTokenId()
	if err != nil {
		return "", err
	}
	return plugin.tokenCache.token.Project.Id, nil
}
func (plugin PasswordAuthPlugin) IsAdmin() bool {
	for _, role := range plugin.tokenCache.token.Roles {
		if role.Name == "admin" {
			return true
		}
	}
	return false
}
func NewPasswordAuth(authUrl string, user User, project Project, regionName string) *PasswordAuthPlugin {
	return &PasswordAuthPlugin{
		session:           httpclient.New(),
		AuthUrl:           authUrl,
		Username:          user.Name,
		Password:          user.Password,
		UserDomainName:    user.Domain.Name,
		ProjectName:       project.Name,
		ProjectDomainName: project.Domain.Name,
		RegionName:        regionName,
	}
}
