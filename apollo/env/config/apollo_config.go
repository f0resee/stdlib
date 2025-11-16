package config

import (
	"sync"

	"github.com/f0resee/stdlib/apollo/utils"
)

type CurrentApolloConfig struct {
	l       sync.RWMutex
	configs map[string]*ApolloConnConfig
}

func (c *CurrentApolloConfig) Set(namespace string, connConfig *ApolloConnConfig) {
	c.l.Lock()
	defer c.l.Unlock()

	c.configs[namespace] = connConfig
}

func (c *CurrentApolloConfig) Get() map[string]*ApolloConnConfig {
	c.l.RLock()
	defer c.l.RUnlock()

	return c.configs
}

func (c *CurrentApolloConfig) GetReleaseKey(namespace string) string {
	c.l.RLock()
	defer c.l.RUnlock()

	config := c.configs[namespace]
	if config == nil {
		return utils.Empty
	}

	return config.ReleaseKey
}

type ApolloConnConfig struct {
	AppID         string `json:"appId"`
	Cluster       string `json:"cluster"`
	NamespaceName string `json:"namespaceName"`
	ReleaseKey    string `json:"releaseKey"`
	sync.RWMutex
}

type ApolloConfig struct {
	ApolloConnConfig
	Configurations map[string]interface{} `json:"configurations"`
}

func (a *ApolloConfig) Init(appID string, cluster string, namespace string) {
	a.AppID = appID
	a.Cluster = cluster
	a.NamespaceName = namespace
}

func CreateCurrentApolloConfig() *CurrentApolloConfig {
	return &CurrentApolloConfig{
		configs: make(map[string]*ApolloConnConfig, 1),
	}
}
