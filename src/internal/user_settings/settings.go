package user_settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type userSettings struct {
	DownloadPath       string `json:"download_path"`
}

type UserSettingsManager struct {
	fpath string
	Settings userSettings
}

func NewUserSettingsManager(userSettingsFname string) (UserSettingsManager, error) {
	fpath, err := getUserSettingsFpath(userSettingsFname)
	if err != nil {
		return UserSettingsManager{}, fmt.Errorf("failed to get user settings file path: %w", err)
	}

	// No user settings have been saved, just return an empty one
	if _, err := os.Stat(fpath); errors.Is(err, os.ErrNotExist) {
		return UserSettingsManager{
			fpath: fpath,
			Settings: userSettings{
				DownloadPath: "",
			},
		}, nil
	}

	settings, err := getUserSettingsFromFile(fpath)
	if err != nil {
		return UserSettingsManager{}, fmt.Errorf("failed to read user settings file: %w", err)
	}
	return UserSettingsManager{
		fpath: fpath,
		Settings: settings,
	}, nil
}

func getUserSettingsFpath(userSettingsFname string) (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	exPath := filepath.Dir(ex)
	return filepath.Join(exPath, userSettingsFname), nil
}

func getUserSettingsFromFile(fpath string) (userSettings, error) {
	file, err := os.Open(fpath)
	if err != nil {
		return userSettings{}, err
	}
	defer file.Close()
	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		return userSettings{}, err
	}
	var settings userSettings
	err = json.Unmarshal(byteValue, &settings)
	if err != nil {
		return userSettings{}, fmt.Errorf("failed to unmarshall user settings json file: %v", err)
	}
	return settings, nil
}

func (us *UserSettingsManager) SaveNewUserSettings(downloadPath string) error {
	us.Settings.DownloadPath = downloadPath
	return us.saveUserSettings()
}

func (us *UserSettingsManager) saveUserSettings() error {
	out, err := json.Marshal(us.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshall user settings to json: %v", err)
	}
	err = ioutil.WriteFile(us.fpath, out, 0644)
	return err
}
