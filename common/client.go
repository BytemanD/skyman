package common

import (
	"fmt"
	"time"

	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/utility"
)

func ClientWithRegion(region string) *openstack.Openstack {
	user := model.User{
		Name:     CONF.Auth.User.Name,
		Password: CONF.Auth.User.Password,
		Domain:   model.Domain{Name: CONF.Auth.User.Domain.Name},
	}
	project := model.Project{
		Name: CONF.Auth.Project.Name,
		Domain: model.Domain{
			Name: CONF.Auth.Project.Domain.Name,
		},
	}
	authUrl := utility.VersionUrl(CONF.Auth.Url, fmt.Sprintf("v%s", CONF.Identity.Api.Version))
	c := openstack.NewClient(authUrl, user, project, region)
	return c
}

func DefaultClient() *openstack.Openstack {
	console.Debug("new openstack client, HttpTimeoutSecond=%d RetryWaitTimeSecond=%d RetryCount=%d",
		CONF.HttpTimeoutSecond, CONF.RetryWaitTimeSecond, CONF.RetryCount,
	)
	c := ClientWithRegion(CONF.Auth.Region.Id)
	c.AuthPlugin.SetLocalTokenExpire(CONF.Auth.TokenExpireTime)
	if CONF.Neutron.Endpoint == "" {
		c.SetNeutronEndpoint(CONF.Neutron.Endpoint)
	}
	if CONF.HttpTimeoutSecond > 0 {
		c.SetHttpTimeout(time.Second * time.Duration(CONF.HttpTimeoutSecond))
	}
	if CONF.RetryWaitTimeSecond > 0 {
		c.SetRetryWaitTime(time.Second * time.Duration(CONF.RetryWaitTimeSecond))
	}
	if CONF.RetryCount > 0 {
		c.SetRetryCount(CONF.RetryCount)
	}
	return c
}
