package config

import "sync"

type CurrentApolloConfig struct {
	l       sync.RWMutex
	configs map[string]*ApolloConnConfig
}

type ApolloConnConfig struct {
}

type ApolloConfig struct {
	ApolloConnConfig
	Configurations map[string]interface{} `json:"configurations"`
}
