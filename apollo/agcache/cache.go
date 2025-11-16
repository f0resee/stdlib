package agcache

type CacheInterface interface {
	Set(key string, value interface{}, expireSeconds int) (err error)
	EntryCount() (entryCount int64)
	Get(key string) (value interface{}, err error)
	Del(key string) (affected bool)
	Range(f func(key, value interface{}) bool)
	Clear()
}

type CacheFactory interface {
	Create() CacheInterface
}
