module github.com/BytemanD/skyman

go 1.24.0

require (
	github.com/spf13/cobra v1.7.0
	gopkg.in/yaml.v3 v3.0.1
	libvirt.org/go/libvirt v1.10002.0
)

require (
	github.com/duke-git/lancet/v2 v2.3.4
	github.com/dustin/go-humanize v1.0.1
	github.com/samber/lo v1.49.1
)

require golang.org/x/exp v0.0.0-20221208152030-732eee02a75a // indirect

require (
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-resty/resty/v2 v2.13.1
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/howeyc/gopass v0.0.0-20210920133722-c8aef6fb66ef
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-runewidth v0.0.16
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.2
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/spf13/afero v1.9.5 // indirect
	github.com/spf13/cast v1.5.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/subosito/gotenv v1.4.2 // indirect
	golang.org/x/crypto v0.23.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/term v0.20.0 // indirect
	golang.org/x/text v0.21.0
	gopkg.in/ini.v1 v1.67.0 // indirect
)

require (
	github.com/BytemanD/easygo/pkg v0.1.2
	github.com/MichaelMure/go-term-markdown v0.1.4
	github.com/cheggaaa/pb/v3 v3.1.5
	github.com/jedib0t/go-pretty/v6 v6.5.8
	github.com/nicksnyder/go-i18n/v2 v2.2.1
	github.com/satori/go.uuid v1.2.0 // indirect
	github.com/spf13/viper v1.16.0
)

require (
	github.com/BytemanD/go-console v0.0.4
	github.com/MichaelMure/go-term-text v0.3.1 // indirect
	github.com/VividCortex/ewma v1.2.0 // indirect
	github.com/alecthomas/chroma v0.7.1 // indirect
	github.com/danwakefield/fnmatch v0.0.0-20160403171240-cbb64ac3d964 // indirect
	github.com/disintegration/imaging v1.6.2 // indirect
	github.com/dlclark/regexp2 v1.1.6 // indirect
	github.com/eliukblau/pixterm/pkg/ansimage v0.0.0-20191210081756-9fb6cf8c2f75 // indirect
	github.com/fatih/color v1.18.0 // direct
	github.com/gomarkdown/markdown v0.0.0-20230922112808-5421fefb8386 // indirect
	github.com/kyokomi/emoji/v2 v2.2.8 // indirect
	github.com/lucasb-eyer/go-colorful v1.0.3 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/wxnacy/wgo v1.0.4
	golang.org/x/image v0.0.0-20191206065243-da761ea9ff43 // indirect
	golang.org/x/net v0.25.0 // indirect
)

replace github.com/BytemanD/skyman => ./

replace github.com/BytemanD/skyman/openstack => ./openstack
