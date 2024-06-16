/*
OpenStack Client with Golang
*/
package openstack

import (
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack/auth"
)

type Openstack struct {
	keystoneClient *KeystoneV3
	glanceClient   *Glance
	neutronClient  *NeutronV2
	cinderClient   *CinderV2
	novaClient     *NovaV2
	AuthPlugin     auth.AuthPlugin
}

func NewClient(authUrl string, user auth.User, project auth.Project,
	regionName string,
) *Openstack {
	passwordAuth := auth.NewPasswordAuth(authUrl, user, project, regionName)
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
	c := NewClient(common.CONF.Auth.Url, user, project, region)
	c.AuthPlugin.SetLocalTokenExpire(common.CONF.Auth.TokenExpireTime)
	return c
}

func (o Openstack) Region() string {
	return o.AuthPlugin.Region()
}

func DefaultClient() *Openstack {
	return Client(common.CONF.Auth.Region.Id)
}
