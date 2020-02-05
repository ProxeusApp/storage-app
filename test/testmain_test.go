package test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	"gopkg.in/gavv/httpexpect.v2"
)

var (
	serviceURL string
	ethAddress string
)

type session struct {
	base string
	id   string
	t    *testing.T
	e    *httpexpect.Expect
	c    *http.Client
}

func TestApi(t *testing.T) {
	url := os.Getenv("STORAGE_APP_URL")
	if len(url) == 0 {
		url = "http://127.0.0.1:8081"
	}
	if !isAvailable(url) {
		t.Fatal("Service not online")
	}

	s := newSession(t, url)

	initStorageDapp(s)
	if s.t.Failed() {
		return
	}

	tests := []struct {
		name string
		f    func(s *session)
	}{
		{"UserTestCreate", testCreateUser},
		{"UserTestLogout", testLogout},
		{"UserTestLogin", testLogin},
		{"FileTestUpload", testFileUpload},
		{"FileTestDownload", testGetDownload},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) { test.f(cloneSession(t, s)) })
	}
}

func newSession(t *testing.T, serverURL string) *session {
	id, err := uuid.NewV4()
	if err != nil {
		t.Error(err)
	}

	return &session{
		base: serverURL,
		id:   id.String(),
		t:    t,
		e: httpexpect.WithConfig(httpexpect.Config{
			BaseURL:  serverURL,
			Reporter: httpexpect.NewAssertReporter(t),
			Printers: []httpexpect.Printer{
				httpexpect.NewCompactPrinter(t),
			},
		}),
	}
}

func cloneSession(t *testing.T, s *session) *session {
	id, err := uuid.NewV4()
	if err != nil {
		t.Error(err)
	}

	return &session{
		base: s.base,
		id:   id.String(),
		t:    s.t,
		e: httpexpect.WithConfig(httpexpect.Config{
			BaseURL:  s.base,
			Reporter: httpexpect.NewAssertReporter(s.t),
			Printers: []httpexpect.Printer{
				httpexpect.NewCompactPrinter(s.t),
			},
		}),
	}
}

func isAvailable(url string) bool {
	for i := 0; i < 10; i++ {
		r, err := http.Get(url)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		defer r.Body.Close()
		_, err = ioutil.ReadAll(r.Body)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		return true
	}
	return false
}

func waitForHttpStatusCode(url string, expectedStatusCode, sleepTime int) bool {
	for i := 0; i < 10; i++ {
		fmt.Sprintf("waitForHttpStatusCode for url: %s called. Expecting: %d", url, expectedStatusCode)

		r, err := http.Get(url)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		defer r.Body.Close()

		_, err = ioutil.ReadAll(r.Body)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		if r.StatusCode != expectedStatusCode {
			fmt.Sprintf("waitForHttpStatusCode not matching. expect: %d, but got: %d",
				expectedStatusCode, r.StatusCode)
			time.Sleep(time.Second * time.Duration(sleepTime))
			continue
		}
		return true
	}
	return false
}

func trimResponseString(response string) string {
	response = strings.Trim(response, "\n")
	response = strings.Trim(response, "\"")
	return response
}

func initStorageDapp(s *session) {

}
