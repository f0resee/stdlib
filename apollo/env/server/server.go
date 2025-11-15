package server

import (
	"strings"
	"sync"
	"time"

	"github.com/f0resee/stdlib/apollo/env/config"
)

var (
	ipMap                map[string]*Info
	serverLock           sync.Mutex
	nextTryConnectPeriod int64 = 30
)

func init() {
	ipMap = make(map[string]*Info)
}

type Info struct {
	serverMap       map[string]*config.ServerInfo
	nextTryConnTime int64
}

func IsConnectDirectly(configIp string) bool {
	serverLock.Lock()
	defer serverLock.Unlock()

	s := ipMap[configIp]
	if s == nil || len(s.serverMap) == 0 {
		return false
	}
	if s.nextTryConnTime >= 0 && s.nextTryConnTime > time.Now().Unix() {
		return true
	}
	return false
}

func GetServers(configIp string) map[string]*config.ServerInfo {
	serverLock.Lock()
	defer serverLock.Unlock()

	if ipMap[configIp] == nil {
		return nil
	}
	return ipMap[configIp].serverMap
}

func SetDownNode(configService string, serverHost string) {
	serverLock.Lock()
	defer serverLock.Unlock()

	s := ipMap[configService]
	if serverHost == "" {
		return
	}

	if s == nil || len(s.serverMap) == 0 {
		ipMap[configService] = &Info{
			serverMap: map[string]*config.ServerInfo{
				serverHost: {
					HomepageURL: serverHost,
				},
			},
		}
	}

	if serverHost == configService {
		s.nextTryConnTime = time.Now().Unix() + nextTryConnectPeriod
	}

	for k, server := range s.serverMap {
		if strings.Contains(k, serverHost) {
			server.IsDown = true
		}
	}
}
