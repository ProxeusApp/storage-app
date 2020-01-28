package updater

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
)

func Download(updateUrls []string) error {
	v, err := fetchVersionInfo(updateUrls)
	if err != nil {
		return err
	}
	url := ""
	switch runtime.GOOS {
	case "darwin":
		url = v.Darwin
	case "linux":
		url = v.Linux
	case "windows":
		url = v.Windows
	}
	if url == "" {
		return errors.New("empty url")
	}

	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status code %d url %s", resp.StatusCode, url)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()
	err = ioutil.WriteFile(updatedExecPath(), b, 0644)
	if err != nil {
		return err
	}
	log.Printf("downloaded update %d bytes from %s", len(b), url)
	return nil
}

func updatedExecPath() string {
	p, _ := os.Executable()
	return filepath.Join(filepath.Dir(p), "Proxeus.update")
}

var applyMu sync.Mutex

func Apply() error {
	applyMu.Lock()
	defer applyMu.Unlock()

	p, err := os.Executable()
	if err != nil {
		return err
	}
	pStat, err := os.Stat(p)
	if err != nil {
		return err
	}
	pMode := pStat.Mode()

	pu := updatedExecPath()

	err = os.Rename(p, p+".old")
	defer func() {
		if err != nil {
			// bring it back!
			os.Rename(p+".old", p)
		}
	}()
	if err != nil {
		log.Println("[Update] Rename 1 failed: ", err)
		return err
	}
	err = os.Rename(pu, p)
	if err != nil {
		log.Println("[Update] Rename 2 failed: ", err)
		return err
	}
	// bring back mod
	err = os.Chmod(p, pMode)
	if err != nil {
		log.Println("[Update] Chmod back failed: ", err)
		return err
	}

	os.Remove(p + ".old")
	err = Restart(p)
	if err != nil {
		log.Println("[Update] Restarting failed: ", err)
		return err
	}

	return nil
}

// only need on Windows
func Cleanup() {
	if runtime.GOOS == "windows" {
		p, _ := os.Executable()
		os.Remove(p + ".old")
	}
}

var onCloseHook func()

// allows us to avoid crash dialog on Mac and error box on Windows
func OnClose(f func()) {
	onCloseHook = f
}

func Restart(p string) error {
	err := exec.Command(p, os.Args[1:]...).Start()
	if err != nil {
		return err
	}
	log.Println("[Update] about to self exit")
	if onCloseHook != nil {
		onCloseHook()
	}
	os.Exit(0)
	return nil
}
