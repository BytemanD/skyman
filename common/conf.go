package common

import (
	"github.com/BytemanD/skyman/common/i18n"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/spf13/viper"
)

var (
	CONF      ConfGroup
	CONF_FILE string
)

var (
	FORMAT_TABLE_LIGHT        = "table-light"
	FORMAT_TABLE              = "table"
	DEFAULT_TOKEN_EXPIRE_TIME = 60 * 30
)

type ConfGroup struct {
	Debug               bool   `yaml:"debug"`
	Format              string `yaml:"format"`
	Language            string `yaml:"language"`
	RetryCount          int    `yaml:"retryCount"`
	LogFile             string `yaml:"logFile"`
	BarChar             string `yaml:"barchar"`
	HttpTimeoutSecond   int    `yaml:"httpTimeoutSecond"`
	RetryWaitTimeSecond int    `yaml:"retryWaitTimeSecond"`
}

func LoadConfig(configFile string) error {
	if err := openstack.LoadConfig(configFile); err != nil {
		return err
	}
	viper.Unmarshal(&CONF)

	i18n.InitLocalizer(CONF.Language)
	if CONF.BarChar != "" {
		nova.BAR_CHAR = CONF.BarChar
	} else {
		switch i18n.GetOsLang() {
		case "en_US":
			nova.BAR_CHAR = "▄"
		default:
			nova.BAR_CHAR = "*"
		}
	}
	return nil
}
