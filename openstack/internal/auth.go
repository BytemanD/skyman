package internal

import (
	"github.com/BytemanD/skyman/openstack/internal/auth_plugin"
	"github.com/BytemanD/skyman/openstack/model"
)

func NewPasswordAuth(authUrl string, user model.User, project model.Project, regionName string) *auth_plugin.PasswordAuthPlugin {
	return auth_plugin.NewPasswordAuthPlugin(authUrl, user, project, regionName)
}
