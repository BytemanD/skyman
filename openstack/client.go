package openstack

import (
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack/auth"
)

type OpenstackClientInterface interface {
	Servers() *RestClient
}

type Openstack struct {
	keystoneClient *NeutronV2
	glanceClient   *Glance
	neutronClient  *NeutronV2
	cinderClient   *CinderV2
	novaClient     *NovaV2
	AuthPlugin     auth.AuthPlugin
}

func NewClient(authUrl string, user auth.User, project auth.Project,
	regionName string, tokenExpireSecond int,
) *Openstack {
	// if authUrl == "" {
	// 	return nil, fmt.Errorf("authUrl is required")
	// }
	passwordAuth := auth.NewPasswordAuth(authUrl, user, project, regionName)
	// passwordAuth.TokenIssue()
	passwordAuth.SetTokenExpireSecond(tokenExpireSecond)

	return &Openstack{AuthPlugin: &passwordAuth}
}

func Client(region string) *Openstack {
	user := auth.User{
		Name:     common.CONF.Auth.User.Name,
		Password: common.CONF.Auth.User.Password,
		Domain:   auth.Domain{Name: common.CONF.Auth.User.Domain.Name},
	}
	project := auth.Project{
		Name: common.CONF.Auth.Project.Name,
		Domain: auth.Domain{
			Name: common.CONF.Auth.Project.Domain.Name,
		},
	}
	return NewClient(
		common.CONF.Auth.Url, user, project, region,
		common.CONF.Auth.TokenExpireTime,
	)
}

func DefaultClient() *Openstack {
	return Client(common.CONF.Auth.Region.Id)
}
