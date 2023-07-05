module github.com/BytemanD/stackcrud

go 1.16

require (
	github.com/spf13/cobra v1.7.0
	gopkg.in/yaml.v3 v3.0.1
)

require github.com/BytemanD/stackcrud/openstack v0.0.0

require (
	github.com/BytemanD/easygo/pkg v0.0.3
	github.com/jedib0t/go-pretty/v6 v6.4.6 // indirect
)

replace github.com/BytemanD/stackcrud => ./

replace github.com/BytemanD/stackcrud/openstack => ./openstack
