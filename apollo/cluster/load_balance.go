package cluster

import "github.com/f0resee/stdlib/apollo/env/config"

type LoadBalance interface {
	Load(servers map[string]*config.ServerInfo) *config.ServerInfo
}
