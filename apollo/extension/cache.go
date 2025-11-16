package extension

import "github.com/f0resee/stdlib/apollo/agcache"

var (
	globalCachefactory agcache.CacheFactory
)

func GetCacheFactory() agcache.CacheFactory {
	return globalCachefactory
}

func SetCacheFactory(cacheFactory agcache.CacheFactory) {
	globalCachefactory = cacheFactory
}
