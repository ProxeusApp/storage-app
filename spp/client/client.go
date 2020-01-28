package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	cache "github.com/ProxeusApp/memcache"

	"github.com/ProxeusApp/storage-app/spp/client/models"
)

var (
	ErrFileNotFound        = errors.New("file not found")                    // File doesn't exist at all, it has never been registered
	ErrForbidden           = errors.New("no permissions to access the file") // User doesn't have the appropriate credentials to access the file
	ErrFileRemoved         = errors.New("file doesn't exist anymore")
	ErrFileNotReady        = errors.New("file not ready yet. Try again") // The file isn't ready to be served yet
	ErrNotSatisfiable      = errors.New("request not satisfiable")
	ErrFilePaymentNotFound = errors.New("file payment not found")
)

var (
	serviceProviderInfoCache = cache.New(3 * time.Minute)
)

func Challenge(urlPath string) (resp *http.Response, err error) {
	client := &http.Client{}
	resp, err = client.Get(urlPath + "/challenge")
	if resp == nil && err == nil {
		err = os.ErrInvalid
		return
	}
	if err == nil && resp != nil && resp.StatusCode != http.StatusOK {
		err = os.ErrInvalid
	}
	return
}

func Input(urlPath, fileHash, token, signature string, reader io.Reader, durationDays int) (resp *http.Response, err error) {
	return InputWithContext(urlPath, fileHash, token, signature, reader, nil, 0, nil, durationDays)
}

func InputWithContext(urlPath, fileHash, token, signature string, reader io.Reader, ctx context.Context,
	filesize int64, transferProgressCallback func(float32), durationDays int) (resp *http.Response, err error) {
	// Wrap the reader into a ProgressReader in order to be able to read percentage of upload
	progressReader := &ProgressReader{
		Reader:     reader,
		Size:       filesize,
		Delay:      200 * time.Millisecond,
		UpdateFunc: transferProgressCallback,
	}
	var req *http.Request
	url := fmt.Sprintf("%s/%s/%s/%s", urlPath, fileHash, token, signature)
	req, err = http.NewRequest("POST", url, progressReader)
	q := req.URL.Query()
	q.Add("duration", strconv.Itoa(durationDays))
	req.URL.RawQuery = q.Encode()
	if err != nil {
		return
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}
	client := &http.Client{}

	log.Printf("About to upload file to spp: fileHash %v | fileSize: %v | duration in days: %v | upload url path: %s",
		fileHash, filesize, durationDays, urlPath)

	resp, err = client.Do(req)
	if resp == nil && err == nil {
		err = os.ErrInvalid
		return
	}
	if err == nil && resp != nil && resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusPaymentRequired {
			err = ErrFilePaymentNotFound
		} else {
			err = os.ErrInvalid
		}
	}
	return
}

type PercentageCallback func(float32)

func Output(urlPath, fileHash, token, signature string, force bool, writer io.Writer) (resp *http.Response, err error) {
	return OutputWithContext(urlPath, fileHash, token, signature, force, writer, nil, nil)
}

func OutputWithContext(urlPath, fileHash, token, signature string, force bool, writer io.Writer, ctx context.Context, transferProgressCallback func(float32)) (resp *http.Response, err error) {
	urlStr := fmt.Sprintf("%s/%s/%s/%s", urlPath, fileHash, token, signature)
	if force {
		urlStr += "?force"
	}
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}
	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		return
	}
	headers := resp.Header
	contentLength, err := strconv.Atoi(headers.Get("Content-Length"))
	if err != nil {
		return resp, errors.New("can't get 'Content-Length' from headers " + err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusAccepted:
			err = ErrFileNotReady
		case http.StatusNotFound:
			err = ErrFileNotFound
		case http.StatusForbidden:
			err = ErrForbidden // No permissions or file removed
		case http.StatusRequestedRangeNotSatisfiable:
			err = ErrNotSatisfiable
		case http.StatusGone:
			err = ErrFileRemoved
			return
		default:
			log.Println("[SPP Client] Unhandled error " + resp.Status + ". Can't download file " + fileHash)
			err = errors.New(resp.Status)
		}
		return
	}
	progressReader := &ProgressReader{
		Reader:     resp.Body,
		Size:       int64(contentLength),
		UpdateFunc: transferProgressCallback,
		Delay:      200 * time.Millisecond,
	}

	if writer != nil {
		_, err = io.Copy(writer, progressReader)
	}
	return
}

// Makes a request to /info and returns the info
func ProviderInfo(urlPath string) (models.StorageProviderInfo, error) {
	spi := models.StorageProviderInfo{}
	err := serviceProviderInfoCache.Get(urlPath, &spi)
	if err == nil {
		log.Println("Returning cached provider info")
		return spi, err
	}
	resp, err := http.Get(urlPath + "/info")
	if err != nil {
		return spi, err
	}
	if resp.StatusCode != http.StatusOK {
		return spi, errors.New("call to /info failed")
	}
	spi.Online = true
	err = json.NewDecoder(resp.Body).Decode(&spi)
	if err != nil {
		return spi, nil
	}
	serviceProviderInfoCache.Put(urlPath, spi)
	return spi, nil
}
