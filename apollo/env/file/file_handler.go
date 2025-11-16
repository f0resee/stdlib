package file

import "github.com/f0resee/stdlib/apollo/env/config"

type FileHandler interface {
	WriteConfigFile(config *config.ApolloConfig, configPath string) error
	GetConfigFile(configDir string, appID string, namespace string) string
	LoadConfigFile(configDir string, appID string, namespace string) (*config.ApolloConfig, error)
}
