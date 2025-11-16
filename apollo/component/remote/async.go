package remote

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"

	"github.com/f0resee/stdlib/apollo/component/log"
	"github.com/f0resee/stdlib/apollo/constant"
	"github.com/f0resee/stdlib/apollo/env"
	"github.com/f0resee/stdlib/apollo/env/config"
	"github.com/f0resee/stdlib/apollo/extension"
	"github.com/f0resee/stdlib/apollo/protocol/http"
	"github.com/f0resee/stdlib/apollo/utils"
)

func CreateAsyncApolloConfig() ApolloConfig {
	a := &asyncApolloConfig{}
	a.remoteApollo = a
	return a
}

type asyncApolloConfig struct {
	AbsApolloConfig
}

// CallBack implements ApolloConfig.
func (a *asyncApolloConfig) CallBack(namespace string) http.CallBack {
	return http.CallBack{
		SuccessCallBack:   createApolloConfigWithJSON,
		NotModifyCallBack: touchApolloConfigCache,
		Namespace:         namespace,
	}
}

// GetNotifyURLSuffix implements ApolloConfig.
func (a *asyncApolloConfig) GetNotifyURLSuffix(notifications string, config config.AppConfig) string {
	return fmt.Sprintf("notifications/v2?appId=%s&cluster=%s&notifications=%s",
		url.QueryEscape(config.AppID),
		url.QueryEscape(config.Cluster),
		url.QueryEscape(notifications),
	)
}

// GetSyncURI implements ApolloConfig.
func (a *asyncApolloConfig) GetSyncURI(config config.AppConfig, namespaceName string) string {
	return fmt.Sprintf("configs/%s/%s/%s?releaseKey=%s&ip=%s&label=%s",
		url.QueryEscape(config.AppID),
		url.QueryEscape(config.Cluster),
		url.QueryEscape(namespaceName),
		url.QueryEscape(config.GetCurrentApolloConfig().GetReleaseKey(namespaceName)),
		utils.GetInternal(),
		url.QueryEscape(config.Label),
	)
}

// Sync implements ApolloConfig.
func (a *asyncApolloConfig) Sync(appConfigFunc func() config.AppConfig) []*config.ApolloConfig {
	appConfig := appConfigFunc()
	remoteConfigs, err := a.notifyRemoteConfig(appConfigFunc, utils.Empty)

	var apolloConfigs []*config.ApolloConfig
	if err != nil {
		apolloConfigs = loadBackupConfig(appConfig.NamespaceName, appConfig)
	}

	if len(remoteConfigs) == 0 || len(apolloConfigs) > 0 {
		return apolloConfigs
	}

	for _, notifyConfig := range remoteConfigs {
		apolloConfig, err := a.SyncWithNamespace(notifyConfig.NamespaceName, appConfigFunc)
		if err == nil {
			appConfig.GetNotificationsMap().UpdateNotify(notifyConfig.NamespaceName, notifyConfig.NotificationID)
		}
		if apolloConfig != nil {
			apolloConfigs = append(apolloConfigs, apolloConfig)
		}
	}
	return apolloConfigs
}

func (a *asyncApolloConfig) notifyRemoteConfig(appConfigFunc func() config.AppConfig, namespace string) ([]*config.Notification, error) {
	if appConfigFunc == nil {
		panic("can not find apollo config! please confirm!")
	}
	appConfig := appConfigFunc()
	notificationsMap := appConfig.GetNotificationsMap()
	urlSuffix := a.GetNotifyURLSuffix(notificationsMap.GetNotifies(namespace), appConfig)

	connectConfig := &env.ConnectConfig{
		URI:    urlSuffix,
		AppID:  appConfig.AppID,
		Secret: appConfig.Secret,
	}
	connectConfig.Timeout = notifyConnectTimeout
	notifies, err := http.RequestRecovery(appConfig, connectConfig, &http.CallBack{
		SuccessCallBack: func(responseBody []byte, callback http.CallBack) (interface{}, error) {
			return toApolloConfig(responseBody)
		},
		NotModifyCallBack: touchApolloConfigCache,
		Namespace:         namespace,
	})
	if notifies == nil {
		return nil, err
	}
	return notifies.([]*config.Notification), err
}

// SyncWithNamespace implements ApolloConfig.
// Subtle: this method shadows the method (AbsApolloConfig).SyncWithNamespace of asyncApolloConfig.AbsApolloConfig.
func (a *asyncApolloConfig) SyncWithNamespace(namespace string, appConfigFunc func() config.AppConfig) (*config.ApolloConfig, error) {
	panic("unimplemented")
}

func touchApolloConfigCache() error {
	return nil
}

func toApolloConfig(resBody []byte) ([]*config.Notification, error) {
	remoteConfig := make([]*config.Notification, 0)
	err := json.Unmarshal(resBody, &remoteConfig)

	if err != nil {
		log.Errorf("Unmarshal Msg Fail, error: %v", err)
		return nil, err
	}
	return remoteConfig, nil
}

func loadBackupConfig(namespace string, appConfig config.AppConfig) []*config.ApolloConfig {
	apolloConfigs := make([]*config.ApolloConfig, 0)
	config.SplitNamespaces(namespace, func(namesapce string) {
		c, err := extension.GetFileHandler().LoadConfigFile(appConfig.BackupConfigPath, appConfig.AppID, namesapce)
		if err != nil {
			log.Errorf("LoadConfigFile error, error: %v", err)
			return
		}
		if c == nil {
			return
		}
		apolloConfigs = append(apolloConfigs, c)
	})
	return apolloConfigs
}

func createApolloConfigWithJSON(b []byte, callback http.CallBack) (o interface{}, err error) {
	apolloConfig := &config.ApolloConfig{}
	err = json.Unmarshal(b, apolloConfig)
	if utils.IsNotNil(err) {
		return nil, err
	}

	parser := extension.GetFormatParser(constant.ConfigFileFormat(path.Ext(apolloConfig.NamespaceName)))
	if parser == nil {
		parser = extension.GetFormatParser(constant.DEFAULT)
	}

	if parser == nil {
		return apolloConfig, nil
	}

	content, ok := apolloConfig.Configurations[defaultContentKey]
	if !ok {
		content = string(b)
	}

	m, err := parser.Parse(content)
	if err != nil {
		log.Debugf("GetContent fail! error: %v", err)
	}

	if len(m) > 0 {
		apolloConfig.Configurations = m
	}
	return apolloConfig, nil
}
