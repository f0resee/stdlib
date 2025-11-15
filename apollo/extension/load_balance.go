package extension

import "github.com/f0resee/stdlib/apollo/cluster"

var defaultLoadBalance cluster.LoadBalance

func SetLoadBalance(loadbalance cluster.LoadBalance) {
	defaultLoadBalance = loadbalance
}

func GetLoadBalance() cluster.LoadBalance {
	return defaultLoadBalance
}
