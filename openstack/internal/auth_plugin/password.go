package auth_plugin

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/openstack/session"
	"github.com/go-resty/resty/v2"
)

const (
	DEFAUL_REGION  string = "RegionOne"
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
	X_AUTH_TOKEN   string = "X-Auth-Token"
)

type PasswordAuthPlugin struct {
	AuthUrl           string
	Username          string
	Password          string
	ProjectName       string
	UserDomainName    string
	ProjectDomainName string

	LocalTokenExpireSecond int
	token                  *model.Token
	expiredAt              time.Time

	tokenLock *sync.Mutex

	session *resty.Client
}

func (plugin *PasswordAuthPlugin) SetLocalTokenExpire(expireSeconds int) {
	plugin.LocalTokenExpireSecond = expireSeconds
}

func (plugin *PasswordAuthPlugin) SetTimeout(timeout time.Duration) {
	plugin.session.SetTimeout(timeout)
}
func (plugin *PasswordAuthPlugin) SetRetryCount(c int) {
	plugin.session.SetRetryCount(c)
}
func (plugin *PasswordAuthPlugin) SetRetryWaitTime(t time.Duration) {
	plugin.session.SetRetryWaitTime(t)
}
func (plugin *PasswordAuthPlugin) SetRetryMaxWaitTime(t time.Duration) {
	plugin.session.SetRetryMaxWaitTime(t)
}

func (plugin *PasswordAuthPlugin) IsTokenExpired() bool {
	if plugin.token == nil {
		return true
	}
	if plugin.expiredAt.Before(time.Now()) {
		console.Warn("token exipred, expired at: %s , now: %s", plugin.expiredAt, time.Now())
		return true
	}
	return false
}

func (plugin *PasswordAuthPlugin) makesureTokenValid() error {
	plugin.tokenLock.Lock()
	defer plugin.tokenLock.Unlock()

	if plugin.IsTokenExpired() {
		return plugin.TokenIssue()
	}
	return nil
}

func (plugin *PasswordAuthPlugin) GetToken() (*model.Token, error) {
	if err := plugin.makesureTokenValid(); err != nil {
		return nil, err
	}
	return plugin.token, nil
}

type AuthBody struct {
	Auth model.Auth `json:"auth"`
}

func (client PasswordAuthPlugin) newAuthReqBody() AuthBody {
	authData := model.Auth{
		Identity: model.Identity{
			Methods: []string{"password"},
			Password: model.Password{
				User: model.User{
					Name: client.Username, Password: client.Password,
					Domain: model.Domain{Name: client.UserDomainName}}},
		},
		Scope: model.Scope{Project: model.Project{
			Name:   client.ProjectName,
			Domain: model.Domain{Name: client.ProjectDomainName}},
		},
	}
	return AuthBody{Auth: authData}
}

func (plugin *PasswordAuthPlugin) TokenIssue() error {
	respBody := struct {
		Token model.Token `json:"token"`
	}{}
	resp, err := plugin.session.R().SetBody(plugin.newAuthReqBody()).
		SetResult(&respBody).
		Post(fmt.Sprintf("%s%s", plugin.AuthUrl, URL_AUTH_TOKEN))
	if err != nil || resp.Error() != nil {
		return fmt.Errorf("token issue failed, %s %s", err, resp.Error())
	}
	if resp.IsError() {
		return fmt.Errorf("token issue failed, [%d] %s", resp.StatusCode(), resp.Body())
	}
	plugin.token = &respBody.Token
	plugin.token.TokenId = resp.Header().Get("X-Subject-Token")
	plugin.expiredAt = time.Now().Add(time.Second * time.Duration(plugin.LocalTokenExpireSecond))
	return nil
}

func (plugin *PasswordAuthPlugin) GetEndpoint(region string, sType string, sName string, sInterface string) (string, error) {
	if err := plugin.makesureTokenValid(); err != nil {
		return "", fmt.Errorf("get token failed: %s", err)
	}
	if region == "" {
		console.Warn("user default region: %s", DEFAUL_REGION)
		region = DEFAUL_REGION
	}
	for _, catalog := range plugin.token.Catalogs {
		if catalog.Type != sType || (sName != "" && catalog.Name != sName) {
			continue
		}
		for _, endpoint := range catalog.Endpoints {
			if endpoint.Interface == sInterface && endpoint.Region == region {
				return endpoint.Url, nil
			}
		}
	}
	return "", fmt.Errorf("endpoint %s:%s:%s for region '%s' not found",
		sType, sName, sInterface, region)
}

func (plugin PasswordAuthPlugin) AuthRequest(req *resty.Request) error {
	token, err := plugin.GetToken()
	if err != nil {
		return err
	}
	if req.Header.Get(X_AUTH_TOKEN) == token.TokenId {
		return nil
	}
	console.Debug("set header %s: %s", X_AUTH_TOKEN, token.TokenId)
	req.Header.Set(X_AUTH_TOKEN, token.TokenId)
	return nil
}
func (plugin PasswordAuthPlugin) GetSafeHeader(header http.Header) http.Header {
	safeHeaders := http.Header{}
	for k, v := range header {
		if k == X_AUTH_TOKEN {
			safeHeaders[k] = []string{"<TOKEN>"}
		} else {
			safeHeaders[k] = v
		}
	}
	return safeHeaders
}
func (plugin PasswordAuthPlugin) GetProjectId() (string, error) {
	if err := plugin.makesureTokenValid(); err != nil {
		return "", err
	}
	return plugin.token.Project.Id, nil
}
func (plugin PasswordAuthPlugin) IsAdmin() bool {
	for _, role := range plugin.token.Roles {
		if role.Name == "admin" {
			return true
		}
	}
	return false
}

func NewPasswordAuthPlugin(authUrl string, user model.User, project model.Project) *PasswordAuthPlugin {
	return &PasswordAuthPlugin{
		session:           session.DefaultRestyClient(authUrl),
		AuthUrl:           authUrl,
		Username:          user.Name,
		Password:          user.Password,
		UserDomainName:    user.Domain.Name,
		ProjectName:       project.Name,
		ProjectDomainName: project.Domain.Name,
		tokenLock:         &sync.Mutex{},
	}
}
