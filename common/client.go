package common

import (
	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/openstack"
)

func DefaultClient() *openstack.Openstack {
	conn, err := openstack.Connect()
	conn.SetRetryCount(CONF.RetryCount)
	if err != nil {
		console.Fatal("connect cloud failed: %s", err)
	}
	return conn
}
