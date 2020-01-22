package updater

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/coreos/go-semver/semver"
)

func Versions(currentVer, contractVer string, updateUrls []string) (interface{}, error) {
	update := "none"
	remoteInfo, err := fetchVersionInfo(updateUrls)
	if err != nil {
		return nil, err
	}
	if currentVer != remoteInfo.Version {
		v, err := semver.NewVersion(currentVer)
		if err != nil {
			fmt.Print("fix ReleaseVersion string")
			os.Exit(-1)
		}
		rv, err := semver.NewVersion(remoteInfo.Version)
		if err != nil {
			return nil, err
		}
		if v.Major != rv.Major {
			update = "block"
		} else {
			update = "info"
		}
	}
	build := fmt.Sprintf("%s", currentVer) // TODO(mmal): add ldflags BuildVersion
	return map[string]string{
		"contract": contractVer,
		"build":    build,
		"update":   update,
	}, nil
}

type VerInfo struct {
	Version string
	Darwin  string
	Linux   string
	Windows string
}

func fetchVersionInfo(updateUrls []string) (VerInfo, error) {
	if len(updateUrls) < 1 {
		return VerInfo{}, errors.New("no update url given")
	}
	var errs []error
	for _, u := range updateUrls {
		v, err := tryFetchWithUrl(u)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		return v, nil
	}
	return VerInfo{}, fmt.Errorf("%v", errs)
}

func tryFetchWithUrl(updateUrl string) (VerInfo, error) {
	resp, err := http.DefaultClient.Get(updateUrl)
	if err != nil {
		return VerInfo{}, err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return VerInfo{}, err
	}
	resp.Body.Close()

	var v VerInfo
	err = json.Unmarshal(b, &v)
	return v, err
}
