package secrets

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type SecretsManager interface {
	GetSecret(secret string) string
}

type JsonFileSecrets struct {
	FilePath string
	Secrets  map[string]string
}

func (jsonSecrets JsonFileSecrets) GetSecret(secret string) string {
	return jsonSecrets.Secrets[secret]
}

func (jsonSecrets *JsonFileSecrets) readSecrets() {
	jsonFile, err := os.Open(jsonSecrets.FilePath)
	if err != nil {
		panic("Secrets could not be read from JSON file path: " + jsonSecrets.FilePath)
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var secrets map[string]string
	json.Unmarshal([]byte(byteValue), &secrets)
	jsonSecrets.Secrets = secrets
}

func NewJsonFileSecrets(filePath string) JsonFileSecrets {
	jsonFileSecrets := JsonFileSecrets{
		FilePath: filePath,
	}
	jsonFileSecrets.readSecrets()
	return jsonFileSecrets
}
