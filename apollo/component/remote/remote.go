package remote

import (
	"github.com/f0resee/stdlib/apollo/env/config"
	"github.com/f0resee/stdlib/apollo/protocol/http"
)

type ApolloConfig interface {
	GetNotifyURLSuffix(notifications string, config config.AppConfig) string
	GetSyncURI(config config.AppConfig, namespaceName string) string
	Sync(appConfigFunc func() config.AppConfig) []*config.ApolloConfig
	CallBack(namespace string) http.CallBack
	SyncWithNamespace(namespace string, appConfigFunc func() config.AppConfig) (*config.ApolloConfig, error)
}
