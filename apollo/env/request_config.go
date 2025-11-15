package env

import "time"

type ConnectConfig struct {
	Timeout time.Duration
	URI     string
	IsRetry bool
	AppID   string
	Secret  string
}
