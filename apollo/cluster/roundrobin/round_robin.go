package roundrobin

import "github.com/f0resee/stdlib/apollo/env/config"

type RoundRobin struct {
}

func (r *RoundRobin) Load(servers map[string]*config.ServerInfo) *config.ServerInfo {
	var returnServer *config.ServerInfo
	for _, server := range servers {
		if server.IsDown {
			continue
		}
		returnServer = server
		break
	}
	return returnServer
}
