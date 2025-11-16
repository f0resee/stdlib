package config

import (
	"encoding/json"
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

func (a *AppConfig) GetIsBackupConfig() bool {
	return a.IsBackupConfig
}

func (a *AppConfig) GetBackupConfigPath() string {
	return a.BackupConfigPath
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

func (a *AppConfig) Init() {
	a.currentConnApolloConfig = CreateCurrentApolloConfig()
	a.initAllNotifications(nil)
}

func (a *AppConfig) initAllNotifications(callback func(namespace string)) {
	ns := SplitNamespaces(a.NamespaceName, callback)
	a.notificationsMap = &notificationsMap{
		notifications: ns,
	}
}

func (a *AppConfig) GetCurrentApolloConfig() *CurrentApolloConfig {
	return a.currentConnApolloConfig
}

func SplitNamespaces(namespacesStr string, callback func(namesapce string)) sync.Map {
	namespaces := sync.Map{}
	split := strings.Split(namespacesStr, Comma)
	for _, namespace := range split {
		if callback != nil {
			callback(namespace)
		}
		namespaces.Store(namespace, defaultNotificationID)
	}
	return namespaces
}

func (a *AppConfig) GetNotificationsMap() *notificationsMap {
	return a.notificationsMap
}

func (a *AppConfig) SetCurrentApolloConfig(apolloConfig *ApolloConnConfig) {
	a.currentConnApolloConfig.Set(apolloConfig.NamespaceName, apolloConfig)
}

type ServerInfo struct {
	AppName     string `json:"appName"`
	InstanceID  string `json:"instanceId"`
	HomepageURL string `json:"homepageUrl"`
	IsDown      bool   `json:"-"`
}

type Notification struct {
	NamespaceName  string `json:"namespaceName"`
	NotificationID int64  `json:"notificationId"`
}

type notificationsMap struct {
	notifications sync.Map
}

func (n *notificationsMap) UpdateAllNotifications(remoteConfigs []*Notification) {
	for _, remoteConfig := range remoteConfigs {
		if remoteConfig.NamespaceName == "" {
			continue
		}
		if n.GetNotify(remoteConfig.NamespaceName) == 0 {
			continue
		}
		n.setNotify(remoteConfig.NamespaceName, remoteConfig.NotificationID)
	}
}

func (n *notificationsMap) UpdateNotify(namespaceName string, notificationID int64) {
	if namespaceName != "" {
		n.setNotify(namespaceName, notificationID)
	}
}

func (n *notificationsMap) setNotify(namespaceName string, notificationID int64) {
	n.notifications.Store(namespaceName, notificationID)
}

func (n *notificationsMap) GetNotify(namesapce string) int64 {
	value, ok := n.notifications.Load(namesapce)
	if !ok || value == nil {
		return 0
	}
	return value.(int64)
}

func (n *notificationsMap) GetNotifyLen() int {
	s := n.notifications
	l := 0
	s.Range(func(k, v interface{}) bool {
		l++
		return true
	})
	return l
}

func (n *notificationsMap) GetNotifications() sync.Map {
	return n.notifications
}

func (n *notificationsMap) GetNotifies(namespace string) string {
	notificationAddr := make([]*Notification, 0)
	if namespace == "" {
		n.notifications.Range(func(key, value any) bool {
			namespaceName := key.(string)
			notificationID := value.(int64)
			notificationAddr = append(notificationAddr, &Notification{
				NamespaceName:  namespaceName,
				NotificationID: notificationID,
			})
			return true
		})
	} else {
		notify, _ := n.notifications.LoadOrStore(namespace, defaultNotificationID)

		notificationAddr = append(notificationAddr, &Notification{
			NamespaceName:  namespace,
			NotificationID: notify.(int64),
		})
	}

	j, err := json.Marshal(notificationAddr)
	if err != nil {
		return ""
	}
	return string(j)
}
