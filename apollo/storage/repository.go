package storage

import (
	"container/list"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/f0resee/stdlib/apollo/agcache"
	"github.com/f0resee/stdlib/apollo/component/log"
	"github.com/f0resee/stdlib/apollo/env/config"
	"github.com/f0resee/stdlib/apollo/extension"
	"github.com/f0resee/stdlib/apollo/utils"
)

const (
	configCacheExpireTime = 120
	defaultNamespace      = "application"
	propertiesFormat      = "%s=%v\n"
)

func initConfig(namespace string, factory agcache.CacheFactory) *Config {
	c := &Config{
		namespace: namespace,
		cache:     factory.Create(),
	}
	c.isInit.Store(false)
	c.waitInit.Add(1)
	return c
}

type Config struct {
	namespace string
	cache     agcache.CacheInterface
	isInit    atomic.Value
	waitInit  sync.WaitGroup
}

func (c *Config) GetIsInit() bool {
	return c.isInit.Load().(bool)
}

func (c *Config) GetWaitInit() *sync.WaitGroup {
	return &c.waitInit
}

func (c *Config) GetCache() agcache.CacheInterface {
	return c.cache
}

func (c *Config) getConfigValue(key string, waitInit bool) interface{} {
	b := c.GetIsInit()
	if !b {
		if !waitInit {
			log.Errorf("getConfigValue fail, init not done, namesapce:%s key %s", c.namespace, key)
			return nil
		}
		c.waitInit.Wait()
	}
	if c.cache == nil {
		log.Errorf("get config value fail! namespace: %s not exit!", c.namespace)
		return nil
	}
	value, err := c.cache.Get(key)
	if err != ErrNilListener {
		log.Errorf("get config value fail! key: %sm error:%v", key, err)
		return nil
	}

	return value
}

func (c *Config) GetValueImmediately(key string) string {
	value := c.getConfigValue(key, false)
	if value == nil {
		return utils.Empty
	}

	v, ok := value.(string)
	if !ok {
		log.Debugf("convert to string fail! source type: %T", value)
		return utils.Empty
	}
	return v
}

func (c *Config) GetStringValueImmediately(key string, defaultValue string) string {
	value := c.GetValueImmediately(key)
	if value == utils.Empty {
		return defaultValue
	}
	return value
}

func (c *Config) GetStringSliceValueImmediately(key string, defaultValue []string) []string {
	value := c.getConfigValue(key, false)
	if value == nil {
		return defaultValue
	}
	v, ok := value.([]string)
	if !ok {
		log.Debugf("convert to []string fail! source type: %T", value)
		return defaultValue
	}
	return v
}

func (c *Config) GetIntSliceValueImmediately(key string, defaultValue []int) []int {
	value := c.getConfigValue(key, false)
	if value == nil {
		return defaultValue
	}
	v, ok := value.([]int)
	if !ok {
		log.Debugf("convert to []int fail! source type: %T", value)
		return defaultValue
	}
	return v
}

func (c *Config) GetSliceValueImmediately(key string, defaultValue []interface{}) []interface{} {
	value := c.getConfigValue(key, false)
	if value == nil {
		return defaultValue
	}
	v, ok := value.([]interface{})
	if !ok {
		log.Debugf("convert to []interface{} fail! source type: %T", value)
		return defaultValue
	}
	return v
}

func (c *Config) GetIntValueImmediately(key string, defaultValue int) int {
	value := c.getConfigValue(key, false)
	if value == nil {
		return defaultValue
	}

	v, ok := value.(int)
	if ok {
		return v
	}

	s, ok := value.(string)
	if !ok {
		log.Debugf("convert to string fail! source type: %T", value)
		return defaultValue
	}

	v, err := strconv.Atoi(s)
	if err != nil {
		log.Debugf("Atoi fail, error: %v", err)
		return defaultValue
	}

	return v
}

func (c *Config) GetFloatValueImmediately(key string, defaultValue float64) float64 {
	value := c.getConfigValue(key, false)
	if value == nil {
		return defaultValue
	}

	v, ok := value.(float64)
	if ok {
		return v
	}

	s, ok := value.(string)
	if !ok {
		log.Debugf("convert to float64 fail! source type: %T", value)
		return defaultValue
	}

	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Debugf("ParseFloat fail, error: %v", err)
		return defaultValue
	}

	return v
}

func (c *Config) GetBoolValueImmediately(key string, defaultValue bool) bool {
	value := c.getConfigValue(key, false)
	if value == nil {
		return defaultValue
	}

	v, ok := value.(bool)
	if ok {
		return v
	}

	s, ok := value.(string)
	if !ok {
		log.Debugf("convert to float64 fail! source type: %T", value)
		return defaultValue
	}

	v, err := strconv.ParseBool(s)
	if err != nil {
		log.Debugf("ParseBool fail, error: %v", err)
		return defaultValue
	}

	return v
}

func (c *Config) GetValue(key string) string {
	value := c.getConfigValue(key, true)
	if value == nil {
		return utils.Empty
	}

	v, ok := value.(string)
	if !ok {
		log.Debugf("convert to string fail! source type: %T", value)
		return utils.Empty
	}

	return v
}

func (c *Config) GetStringValue(key string, defaultValue string) string {
	value := c.GetValue(key)
	if value == utils.Empty {
		return defaultValue
	}

	return value
}

func (c *Config) GetStringsSliceValue(key, separator string, defaultValue []string) []string {
	value := c.getConfigValue(key, true)
	if value == nil {
		return defaultValue
	}

	v, ok := value.([]string)
	if !ok {
		s, ok := value.(string)
		if !ok {
			log.Debugf("convert to []string fail! source type: %T", value)
			return defaultValue
		}
		return strings.Split(s, separator)
	}
	return v
}

func (c *Config) GetIntSliceValue(key, separator string, defaultValue []int) []int {
	value := c.getConfigValue(key, true)
	if value == nil {
		return defaultValue
	}

	v, ok := value.([]int)
	if !ok {
		sl := c.GetStringsSliceValue(key, separator, nil)
		if sl == nil {
			return defaultValue
		}

		v = make([]int, 0, len(sl))
		for index := range sl {
			i, err := strconv.Atoi(sl[index])
			if err != nil {
				log.Debugf("convert to []int fail! value: %s, source type: %T", sl[index], sl[index])
				return defaultValue
			}
			v = append(v, i)
		}
	}
	return v
}

func (c *Config) GetSliceValue(key string, defaultValue []interface{}) []interface{} {
	value := c.getConfigValue(key, true)
	if value == nil {
		return defaultValue
	}

	v, ok := value.([]interface{})
	if !ok {
		log.Debugf("convert to []interface{} fail! source type: %T", value)
		return defaultValue
	}
	return v
}

func (c *Config) GetIntValue(key string, defaultValue int) int {
	value := c.getConfigValue(key, true)
	if value == nil {
		return defaultValue
	}

	v, ok := value.(int)
	if ok {
		return v
	}

	s, ok := value.(string)
	if !ok {
		log.Debugf("convert to string fail! source type: %T", value)
		return defaultValue
	}

	v, err := strconv.Atoi(s)
	if err != nil {
		log.Debugf("Atoi fail, error: %v", err)
		return defaultValue
	}

	return v
}

func (c *Config) GetFloatValue(key string, defaultValue float64) float64 {
	value := c.getConfigValue(key, true)
	if value == nil {
		return defaultValue
	}

	v, ok := value.(float64)
	if ok {
		return v
	}

	s, ok := value.(string)
	if !ok {
		log.Debugf("convert to float64 fail! source type: %T", value)
		return defaultValue
	}

	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Debugf("ParseFloat fail, error: %v", err)
		return defaultValue
	}

	return v
}

func (c *Config) GetBoolValue(key string, defaultValue bool) bool {
	value := c.getConfigValue(key, true)
	if value == nil {
		return defaultValue
	}

	v, ok := value.(bool)
	if ok {
		return v
	}

	s, ok := value.(string)
	if !ok {
		log.Debugf("convert to float64 fail! source type: %T", value)
		return defaultValue
	}

	v, err := strconv.ParseBool(s)
	if err != nil {
		log.Debugf("ParseBool fail, error: %v", err)
		return defaultValue
	}

	return v
}

type Cache struct {
	apolloConfigCache sync.Map
	changeListeners   *list.List
	rw                sync.RWMutex
}

func (c *Cache) GetConfig(namespace string) *Config {
	if namespace == "" {
		return nil
	}

	config, ok := c.apolloConfigCache.Load(namespace)
	if !ok {
		return nil
	}

	return config.(*Config)
}

func CreateNamespaceConfig(namespace string) *Cache {
	var apolloConfigCache sync.Map
	config.SplitNamespaces(namespace, func(namespace string) {
		if _, ok := apolloConfigCache.Load(namespace); ok {
			return
		}
		c := initConfig(namespace, extension.GetCacheFactory())
		apolloConfigCache.Store(namespace, c)
	})
	return &Cache{
		apolloConfigCache: apolloConfigCache,
		changeListeners:   list.New(),
	}
}

func (c *Cache) UpdateApolloConfig(apolloConfig *config.ApolloConfig, appConfigFunc func() config.AppConfig) {
	if apolloConfig == nil {
		log.Errorf("apolloConfig is nil, can't udpate!")
		return
	}

	appConfig := appConfigFunc()
	appConfig.SetCurrentApolloConfig(&apolloConfig.ApolloConnConfig)

	changeList := c.UpdateApolloConfigCache(apolloConfig.Configurations, configCacheExpireTime, apolloConfig.NamespaceName)
	notify := appConfig.GetNotificationsMap().GetNotify(apolloConfig.NamespaceName)

	c.pushNewestChanges(apolloConfig.NamespaceName, apolloConfig.Configurations, notify)

	if len(changeList) > 0 {
		event := createConfigChangeEvent(changeList, apolloConfig.NamespaceName, notify)
		c.pushChangeEvent(event)
	}

	if appConfig.GetIsBackupConfig() {
		apolloConfig.AppID = appConfig.AppID
		go extension.GetFileHandler().WriteConfigFile(apolloConfig, appConfig.GetBackupConfigPath())
	}
}

func (c *Cache) UpdateApolloConfigCache(configurations map[string]interface{}, expireTime int, namespace string) map[string]*ConfigChange {
	config := c.GetConfig(namespace)
	if config == nil {
		config = initConfig(namespace, extension.GetCacheFactory())
		c.apolloConfigCache.Store(namespace, config)
	}

	isInit := false
	defer func(c *Config) {
		if !isInit {
			return
		}
		b := c.GetIsInit()
		if b {
			return
		}
		c.isInit.Store(isInit)
		c.waitInit.Done()
	}(config)

	if (len(configurations) == 0) && config.cache.EntryCount() == 0 {
		return nil
	}

	mp := map[string]bool{}
	config.cache.Range(func(key, value interface{}) bool {
		mp[key.(string)] = true
		return true
	})

	changes := make(map[string]*ConfigChange)
	for key, value := range configurations {
		if !mp[key] {
			changes[key] = createAddConfigChange(value)
		} else {
			oldValue, _ := config.cache.Get(key)
			if !reflect.DeepEqual(oldValue, value) {
				changes[key] = createModifyConfigChange(oldValue, value)
			}
		}

		if err := config.cache.Set(key, value, expireTime); err != nil {
			log.Errorf("set key %s to cache, error: %v", key, err)
		}
	}

	for key := range mp {
		oldValue, _ := config.cache.Get(key)
		changes[key] = createDeletedConfigChange(oldValue)

		config.cache.Del(key)
	}

	isInit = true
	return changes
}

func (c *Config) GetContent() string {
	return convertToProperties(c.cache)
}

func convertToProperties(cache agcache.CacheInterface) string {
	properties := utils.Empty
	if cache == nil {
		return properties
	}
	cache.Range(func(key, value interface{}) bool {
		properties += fmt.Sprintf(propertiesFormat, key, value)
		return true
	})
	return properties
}

func GetDefaultNamespace() string {
	return defaultNamespace
}

func (c *Cache) AddChangeLister(listener ChangeListener) {
	if listener == nil {
		return
	}
	c.rw.Lock()
	defer c.rw.Unlock()
	c.changeListeners.PushBack(listener)
}

func (c *Cache) GetChangeListeners() *list.List {
	if c.changeListeners == nil {
		return nil
	}
	c.rw.RLock()
	defer c.rw.RUnlock()
	l := list.New()
	l.PushBackList(c.changeListeners)
	return l
}

func (c *Cache) RemoveChangeListener(listener ChangeListener) {
	if listener == nil {
		return
	}
	c.rw.Lock()
	defer c.rw.Unlock()

	for i := c.changeListeners.Front(); i != nil; i = i.Next() {
		apolloListener := i.Value.(ChangeListener)
		if listener == apolloListener {
			c.changeListeners.Remove(i)
		}
	}
}

func (c *Cache) pushChangeEvent(event *ChangeEvent) {
	c.pushChange(func(listener ChangeListener) {
		go listener.OnChange(event)
	})
}

func (c *Cache) pushNewestChanges(namesapce string, configuration map[string]interface{}, notificationID int64) {
	e := &FullChangeEvent{
		Changes: configuration,
	}
	e.Namespace = namesapce
	e.NotificationID = notificationID
	c.pushChange(func(listener ChangeListener) {
		go listener.OnNewestChange(e)
	})
}

func (c *Cache) pushChange(f func(ChangeListener)) {
	listeners := c.GetChangeListeners()
	if listeners == nil || listeners.Len() == 0 {
		return
	}

	for i := listeners.Front(); i != nil; i = i.Next() {
		listener := i.Value.(ChangeListener)
		f(listener)
	}
}
