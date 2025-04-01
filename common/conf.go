package common

import (
	"errors"
	"fmt"

	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/common/i18n"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/samber/lo"
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
	Cloud               string `yaml:"cloud"`
	Language            string `yaml:"language"`
	RetryCount          int    `yaml:"retryCount"`
	LogFile             string `yaml:"logFile"`
	BarChar             string `yaml:"barchar"`
	HttpTimeoutSecond   int    `yaml:"httpTimeoutSecond"`
	RetryWaitTimeSecond int    `yaml:"retryWaitTimeSecond"`
}

var ErrInvalidConf = errors.New("invalid format in config file")

func LoadConfig(configFile string) (err error) {
	if err = openstack.LoadConfig(configFile); err != nil {
		return
	}
	if err = viper.Unmarshal(&CONF); err != nil {
		return fmt.Errorf("unmarshal config file failed: %w", err)
	}

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
	if CONF.Debug {
		console.EnableLogDebug()
	}
	if CONF.LogFile != "" {
		console.SetLogFile(CONF.LogFile)
	}
	if !lo.Contains(GetOutputFormats(), CONF.Format) {
		return fmt.Errorf("invalid foramt '%s'", CONF.Format)
	}
	openstack.SetCloudName(CONF.Cloud)

	return nil
}
