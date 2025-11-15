package remote

import (
	"time"

	"github.com/f0resee/stdlib/apollo/component/log"
	"github.com/f0resee/stdlib/apollo/env"
	"github.com/f0resee/stdlib/apollo/env/config"
	"github.com/f0resee/stdlib/apollo/protocol/http"
)

const (
	notifyConnectTimeout = 10 * time.Minute
	defaultContentKey    = "content"
)

type AbsApolloConfig struct {
	remoteApollo ApolloConfig
}

func (a *AbsApolloConfig) SyncWithNamespace(namespace string, appConfigFunc func() config.AppConfig) (*config.ApolloConfig, error) {
	if appConfigFunc == nil {
		panic("can not find apollo config! please confirm!")
	}

	appConfig := appConfigFunc()
	urlSuffix := a.remoteApollo.GetSyncURI(appConfig, namespace)

	c := &env.ConnectConfig{
		URI:     urlSuffix,
		AppID:   appConfig.AppID,
		Secret:  appConfig.Secret,
		Timeout: notifyConnectTimeout,
		IsRetry: true,
	}
	if appConfig.SyncServerTimeout > 0 {
		c.Timeout = time.Duration(appConfig.SyncServerTimeout) * time.Second
	}

	callback := a.remoteApollo.CallBack(namespace)
	apolloConfig, err := http.RequestRecovery(appConfig, c, &callback)
	if err != nil {
		log.Errorf("request %s fail, error: %v", urlSuffix, err)
		return nil, err
	}

	if apolloConfig == nil {
		log.Debug("apolloConfig is nil")
		return nil, nil
	}

	return apolloConfig.(*config.ApolloConfig), nil
}
