package reddit_cli

import (
	"earthpullr/pkg/file_readers"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"os"
	"path/filepath"
)

type ExistingBackgrounds struct {
	logger *zap.Logger
	fpath string
	existingBackgrounds *map[string]string
}

func NewExistingBackgrounds(downloadPath string, existingImagesFname string, logger *zap.Logger) *ExistingBackgrounds {
	ebs := ExistingBackgrounds {
		logger: logger,
		fpath: filepath.Join(downloadPath, existingImagesFname),
	}
	ebs.existingBackgrounds = ebs.getExistingBackgroundsMap()
	return &ebs
}

func (eb *ExistingBackgrounds) getExistingBackgroundsMap() *map[string]string {
	existingBackgrounds := map[string]string{}
	if _, err := os.Stat(eb.fpath); errors.Is(err, os.ErrNotExist) {
		// Existing backgrounds file doesn't exist
		return &existingBackgrounds
	}
	existingBackgroundsJson, err := file_readers.NewFlatJsonFile(eb.fpath)
	if err != nil {
		eb.logger.Error("failed to read existing images json file", zap.Error(err))
		return &existingBackgrounds
	}
	return &existingBackgroundsJson.Data
}

func (eb *ExistingBackgrounds) AddBackground(backgroundFname string) {
	(*eb.existingBackgrounds)[backgroundFname] = "s"
}

func (eb *ExistingBackgrounds) HasBackground(backgroundFname string) bool {
	if _, ok := (*eb.existingBackgrounds)[backgroundFname]; ok {
		return true
	}
	return false
}

func (eb *ExistingBackgrounds) SaveExistingBackgrounds() error {
	err := file_readers.SaveMapAsJson(*eb.existingBackgrounds, eb.fpath)
	if err != nil {
		eb.logger.Error(fmt.Sprintf("failed to save existing background file to '%s'", eb.fpath))
		return err
	}
	eb.logger.Info(fmt.Sprintf("saved existing backgrounds file to '%s'", eb.fpath))
	return nil
}