package notify

import (
	"time"

	"github.com/f0resee/stdlib/apollo/component/remote"
	"github.com/f0resee/stdlib/apollo/env/config"
	"github.com/f0resee/stdlib/apollo/storage"
)

const (
	longPollInterval = 2 * time.Second
)

type ConfigComponent struct {
	appConfigFunc func() config.AppConfig
	cache         *storage.Cache
	stopCh        chan interface{}
}

func (c *ConfigComponent) SetAppConfig(appConfigFunc func() config.AppConfig) {
	c.appConfigFunc = appConfigFunc
}

func (c *ConfigComponent) SetCache(cache *storage.Cache) {
	c.cache = cache
}

func (c *ConfigComponent) Start() {
	if c.stopCh == nil {
		c.stopCh = make(chan interface{})
	}

	t2 := time.NewTimer(longPollInterval)
	instance := remote.CreateAsyncApolloConfig()
loop:
	for {
		select {
		case <-t2.C:
			configs := instance.Sync(c.appConfigFunc)
			for _, apolloConfig := range configs {
				c.cache.UpdateApolloConfig(apolloConfig, c.appConfigFunc)
			}
			t2.Reset(longPollInterval)
		case <-c.stopCh:
			break loop
		}
	}
}

func (c *ConfigComponent) Stop() {
	if c.stopCh != nil {
		close(c.stopCh)
	}
}
