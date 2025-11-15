package config

import (
	"net/url"
	"strings"
	"sync"
)

var (
	defaultNotificationID = int64(-1)
	Comma                 = " "
)

type File interface {
	Load(fileName string, unmarshal func([]byte) (interface{}, error)) (interface{}, error)
	Write(content interface{}, configPath string) error
}

type AppConfig struct {
	AppID                   string `json:"appId"`
	Cluster                 string `json:"cluster"`
	NamespaceName           string `json:"namespaceName"`
	IP                      string `json:"ip"`
	IsBackupConfig          bool   `default:"true" json:"isBackupConfig"`
	BackupConfigPath        string `json:"backupConfigPath"`
	Secret                  string `json:"secret"`
	Label                   string `json:"label"`
	SyncServerTimeout       int    `json:"syncServerTimeout"`
	MustStart               bool   `default:"false"`
	notificationsMap        *notificationsMap
	currentConnApolloConfig *CurrentApolloConfig
}

func (a *AppConfig) GetHost() string {
	u, err := url.Parse(a.IP)
	if err != nil {
		return a.IP
	}
	if !strings.HasSuffix(u.Path, "/") {
		return u.String() + "/"
	}
	return u.String()
}

type ServerInfo struct {
	AppName     string `json:"appName"`
	InstanceID  string `json:"instanceId"`
	HomepageURL string `json:"homepageUrl"`
	IsDown      bool   `json:"-"`
}

type notificationsMap struct {
	notifications sync.Map
}
