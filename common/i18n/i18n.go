package i18n

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/BytemanD/go-console/console"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pelletier/go-toml/v2"
	"golang.org/x/text/language"
)

const (
	ENV_SKYMAN_LANG = "SKYMAN_LANG"
)

var localizer *i18n.Localizer
var bundle *i18n.Bundle

func T(msgId string) string {
	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:      msgId,
		DefaultMessage: getDefaultMessage(msgId),
	})
	if err != nil {
		console.Warn("localize message %s failed, %v", msgId, err)
	}
	return msg
}

func InitLocalizer(lang string) {
	if lang != "" && lang != "en_US" {
		bundle.LoadMessageFile(path.Join("locale", fmt.Sprintf("%s.toml", lang)))
		localizer = i18n.NewLocalizer(bundle, lang, "en-US")
	}
}

// TODO: remove to easygo
func GetOsLang() string {
	osLang := os.Getenv(ENV_SKYMAN_LANG)
	if osLang == "" {
		osLang = os.Getenv("LANG")
	}
	if osLang != "" {
		osLangList := strings.Split(osLang, ".")
		return osLangList[0]
	}
	return ""
}

func init() {
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	osLang := GetOsLang()
	if osLang != "" && osLang != "en_US" {
		_, err := bundle.LoadMessageFile(
			path.Join("locale", fmt.Sprintf("%s.toml", osLang)))
		if err != nil {
			bundle.LoadMessageFile(
				path.Join("/usr/share/skyman/locale", fmt.Sprintf("%s.toml", osLang)))
		}
		localizer = i18n.NewLocalizer(bundle, osLang, "en-US")
	} else {
		localizer = i18n.NewLocalizer(bundle, "en-US")
	}
}
