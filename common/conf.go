package common

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/BytemanD/easygo/pkg/global/logging"
)

var CONF_FILES = []string{
	"etc/stackcrud.yaml",
	"/etc/stackcrud/stackcrud.yaml",
}

var (
	CONF      ConfGroup
	CONF_FILE string
)

type ConfGroup struct {
	Debug bool `yaml:"debug"`
	Auth  Auth `yaml:"auth"`
}
type Auth struct {
	Url             string            `yaml:"url"`
	RegionName      string            `yaml:"regionName"`
	User            map[string]string `yaml:"user"`
	Project         map[string]string `yaml:"project"`
	TokenExpireTime int               `yaml:"tokenExpireTime"`
}

func fileExists(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !fi.IsDir()
}

func LoadConf(confFiles []string) error {
	if len(confFiles) == 0 {
		confFiles = CONF_FILES
	}
	for _, file := range confFiles {
		if !fileExists(file) {
			continue
		}
		CONF_FILE = file
		logging.Debug("load conf from %s", file)
		bytes, err := ioutil.ReadFile(file)
		if err != nil {
			return fmt.Errorf("load config %s failed: %s", file, err)
		}
		yaml.Unmarshal(bytes, &CONF)
		break
	}
	if CONF_FILE == "" {
		return fmt.Errorf("config file not found, find paths: %v", confFiles)
	}
	return nil
}
