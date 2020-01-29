package service

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/ProxeusApp/storage-app/spp/client/models"
)

var tmpPath string

// FileSettingsService
func TestMain(m *testing.M) {
	tmpPath, _ = ioutil.TempDir("", "settings-test")
	ret := m.Run()
	os.RemoveAll(tmpPath)
	os.Exit(ret)
}

func TestFileIsReadCorrectly(t *testing.T) {
	file, err := ioutil.TempFile(tmpPath, "settings-test-file")
	if err != nil {
		t.Error(err)
	}
	jsonSettings := models.StorageProviderInfo{
		Name: "Test settings",
	}
	b, _ := json.Marshal(jsonSettings)
	file.Write(b)

	settingsService, err := NewProviderInfoService(file.Name())
	if err != nil {
		t.Error(err)
	}
	settings := settingsService.Get()
	if settings.Name != jsonSettings.Name {
		t.Error("Name isn't correct")
	}
}

func TestFileNonExistingReturnsAppropriateError(t *testing.T) {
	_, err := NewProviderInfoService("anyNonExistingFile.txt")
	if err == nil {
		t.Error("An error should be returned")
	}
}
