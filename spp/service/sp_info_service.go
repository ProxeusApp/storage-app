package service

import (
	"encoding/json"
	"os"
	"path/filepath"

	"errors"

	"github.com/ProxeusApp/storage-app/spp/client/models"
)

/**
ProviderInfoService takes care of managing information about the provider in all its details.
Service Provider's info have not to be confused with configuration.
An instance should be reused over the application
*/
type ProviderInfoService interface {
	Get() models.StorageProviderInfo
}

/*
	Implements ProviderInfoService and is supposed to manage provider info on a file
	Usage:

		pis := NewProviderInfoService("settings.json")
		pis.Get()
*/
type FileProviderInfoService struct {
	filename            string                      // JSON file where to read the settings from
	serviceProviderInfo *models.StorageProviderInfo // Cached settings
}

func NewProviderInfoService(filename string) (ProviderInfoService, error) {
	providerInfoService := &FileProviderInfoService{
		filename: filename,
	}

	wd, err := os.Getwd()
	if err != nil {
		return providerInfoService, err
	}
	relFilePath := filepath.Join(wd, "settings.json")

	configFile, err := os.Open(relFilePath)
	defer configFile.Close()
	if err != nil {
		return providerInfoService, errors.New("Can't open " + filename + ". " + err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&providerInfoService.serviceProviderInfo)
	return providerInfoService, err
}

// Returns Settings by value
func (me *FileProviderInfoService) Get() models.StorageProviderInfo {
	return *me.serviceProviderInfo
}
