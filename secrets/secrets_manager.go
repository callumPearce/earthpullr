package secrets

import (
	"earthpullr/file_readers"
)

type SecretsManager interface {
	GetSecret(secret string) (string, error)
}

type FlatJsonFileSecretManagerAdaptor struct {
	FlatJsonFile file_readers.FlatJsonFile
}

func (adapter FlatJsonFileSecretManagerAdaptor) GetSecret(secret string) (string, error) {
	return adapter.FlatJsonFile.GetKey(secret)
}
