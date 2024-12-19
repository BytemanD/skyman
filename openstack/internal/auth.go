package internal

import (
	"github.com/BytemanD/skyman/openstack/auth"
	"github.com/BytemanD/skyman/openstack/internal/auth_plugin"
)

func NewPasswordAuth(authUrl string, user auth.User, project auth.Project, regionName string) *auth_plugin.PasswordAuthPlugin {
	return auth_plugin.NewPasswordAuthPlugin(authUrl, user, project, regionName)
}
