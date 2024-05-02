package i18n

import (
	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var DefaultMessages []i18n.Message
var defaultMessageMap map[string]i18n.Message

func getDefaultMessage(msgId string) *i18n.Message {
	value, ok := defaultMessageMap[msgId]
	if !ok {
		logging.Fatal("get default message failed, %s", msgId)
	}
	return &value
}

func init() {
	DefaultMessages = []i18n.Message{
		{ID: "thePathOfConfigFile",
			Other: "the path of config file",
		},
		{ID: "showDebug",
			Other: "show debug messages",
		},
		{ID: "formatAndSupported",
			Other: "output format, supported: %v",
		},
		{ID: "answerYes",
			Other: "automatically answer yes for all questions",
		},
		{ID: "listServers",
			Other: "List Servers",
		},
		{ID: "showServerDetails",
			Other: "Show server details",
		},
		{ID: "localFuzzySearch",
			Other: "Local fuzzy search",
		},
		{ID: "testServerNetworkQOS",
			Other: "Test Server Network QOS",
		},
		{ID: "logFile", Other: "log file"},
	}
	defaultMessageMap = map[string]i18n.Message{}

	for _, message := range DefaultMessages {
		defaultMessageMap[message.ID] = message
	}
}
