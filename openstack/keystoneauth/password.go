package keystoneauth

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/BytemanD/easygo/pkg/global/logging"
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

	tokenCache TokenCache
}

func (plugin PasswordAuthPlugin) GetToken() Token {
	if plugin.isTokenExpired() {
		plugin.TokenIssue()
	}
	return plugin.tokenCache.token
}

func (client PasswordAuthPlugin) GetTokenId() (string, error) {
	if client.isTokenExpired() {
		if err := client.TokenIssue(); err != nil {
			return "", err
		}
	}
	return client.tokenCache.TokenId, nil
}
func (client PasswordAuthPlugin) isTokenExpired() bool {
	if client.tokenCache.TokenId == "" {
		return true
	}
	if client.tokenCache.expiredAt.Before(time.Now()) {
		logging.Debug("token is exipred, expire second is %d", client.tokenCache.TokenExpireSecond)
		return true
	}
	return false
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

func (plugin *PasswordAuthPlugin) TokenIssue() error {
	body, _ := json.Marshal(plugin.getAuthReqBody())
	resp, err := plugin.restfulClient.Post(
		fmt.Sprintf("%s%s", plugin.AuthUrl, URL_AUTH_TOKEN), body)
	if err != nil {
		return fmt.Errorf("token issue failed, %v", err)
	}
	var resToken RespToken
	resp.BodyUnmarshal(&resToken)

	plugin.tokenCache = TokenCache{
		token:     resToken.Token,
		TokenId:   resp.GetHeader("X-Subject-Token"),
		expiredAt: time.Now().Add(time.Second * time.Duration(plugin.tokenCache.TokenExpireSecond)),
	}
	return nil
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
