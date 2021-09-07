package file_readers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type FlatJsonFile struct {
	FilePath string
	Data     map[string]string
}

func (jsonFile FlatJsonFile) GetKey(key string) (value string, err error) {
	value, ok := jsonFile.Data[key]
	if !ok {
		err = fmt.Errorf("key '%s' does not exist within JSON file at %s", key, jsonFile.FilePath)
	}
	return value, err
}

func (jsonFile FlatJsonFile) GetMultiKeys(keys []string) (values map[string]string, err error) {
	values = make(map[string]string)
	for _, key := range keys {
		value, getErr := jsonFile.GetKey(key)
		if getErr != nil {
			if err != nil {
				err = fmt.Errorf("%v, %v", err, getErr)
			} else {
				err = getErr
			}
		} else {
			values[key] = value
		}
	}
	return values, err
}

func (jsonFile *FlatJsonFile) readJsonFile() (err error) {
	file, err := os.Open(jsonFile.FilePath)
	if err != nil {
		return err
	}
	defer file.Close()
	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	var secrets map[string]string
	json.Unmarshal([]byte(byteValue), &secrets)
	jsonFile.Data = secrets
	return err
}

func NewFlatJsonFile(filePath string) (FlatJsonFile, error) {
	jsonFile := FlatJsonFile{
		FilePath: filePath,
	}
	err := jsonFile.readJsonFile()
	return jsonFile, err
}
