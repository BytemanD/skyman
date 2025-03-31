package i18n

import (
	"github.com/BytemanD/go-console/console"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var DefaultMessages []i18n.Message
var defaultMessageMap map[string]i18n.Message

func getDefaultMessage(msgId string) *i18n.Message {
	value, ok := defaultMessageMap[msgId]
	if !ok {
		console.Fatal("get default message failed, %s", msgId)

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
		{ID: "allTenants",
			Other: "all tenants",
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
			Other: "Test server network QOS",
		},
		{ID: "testServerDiskIO",
			Other: "Test server disk io",
		},
		{ID: "logFile", Other: "log file"},
		{ID: "enableLogColor", Other: "enable log color"},
		{ID: "theNumOfTask", Other: "the num of task"},
		{ID: "theNumOfWorker", Other: "the num of worker"},
		{ID: "reportServerEvents", Other: "report server events"},
		{ID: "defineResourcesFromTempFile",
			Other: "define resources from template file"},
		{ID: "undefineResourcesFromTempFile",
			Other: "undefine resources from template file"},
		{ID: "cloudName", Other: "cloud name"},
	}
	defaultMessageMap = map[string]i18n.Message{}

	for _, message := range DefaultMessages {
		defaultMessageMap[message.ID] = message
	}
}
