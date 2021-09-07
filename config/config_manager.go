package config

import "earthpullr/file_readers"

type ConfigManager interface {
	GetConfig(key string) (string, error)
	GetMultiConfig(keys []string) (map[string]string, error)
}

type FlatJsonFileConfigManagerAdaptor struct {
	FlatJsonFile file_readers.FlatJsonFile
}

func (adapter FlatJsonFileConfigManagerAdaptor) GetConfig(key string) (string, error) {
	return adapter.FlatJsonFile.GetKey(key)
}

func (adapter FlatJsonFileConfigManagerAdaptor) GetMultiConfig(keys []string) (map[string]string, error) {
	return adapter.FlatJsonFile.GetMultiKeys(keys)
}
