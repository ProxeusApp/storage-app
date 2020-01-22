package pgpService

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

type (
	ChallengeMsg struct {
		Token     string `json:"token"`
		Challenge string `json:"challenge"`
	}
	Client struct {
		url string
	}
)

func NewClient(url string) (*Client, error) {
	if strings.HasPrefix(url, "http") {
		c := &Client{}
		for strings.HasSuffix(url, "/") {
			url = url[:len(url)-1]
		}
		c.url = url
		fmt.Println(c.url)
		return c, nil
	}
	return nil, os.ErrInvalid
}

func (me *Client) GetURL() string {
	return me.url
}

func (me *Client) Challenge() (*ChallengeMsg, error) {
	client := &http.Client{}
	client.Timeout = time.Duration(time.Second * 10)
	resp, err := client.Get(me.url + "/pks/challenge")
	if err != nil {
		return nil, err
	}
	if resp == nil && err == nil {
		err = os.ErrInvalid
		return nil, err
	}
	if err == nil && resp != nil && resp.StatusCode != 200 {
		err = os.ErrInvalid
	}
	cm := ChallengeMsg{}
	bts, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bts, &cm)
	return &cm, err
}

func (me *Client) Add(pgpPublicKey, token, signature string) (bool, error) {
	params := struct {
		Pubkey    string `json:"pubkey"`
		Token     string `json:"token"`
		Signature string `json:"signature"`
	}{Pubkey: pgpPublicKey, Token: token, Signature: signature}
	reader := &bytes.Buffer{}
	bts, err := json.Marshal(params)
	if err != nil {
		return false, err
	}
	reader.Write(bts)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/pks/add", me.url), reader)
	if err == nil {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Connection", "Keep-Alive")
		req.Header.Set("Cache-Control", "no-cache")
		req.Header.Set("Accept-Charset", "utf-8")
		req.Header.Set("charset", "utf-8")

		client := &http.Client{}
		client.Timeout = time.Duration(time.Second * 10)
		resp, err := client.Do(req)
		if resp == nil && err == nil {
			err = os.ErrInvalid
			return false, err
		}
		if err == nil && resp != nil {
			if resp.StatusCode == 200 {
				return true, err
			} else if resp.StatusCode == 208 {
				return false, os.ErrExist
			}
			err = os.ErrInvalid
		}
	}
	return false, err
}

func (me *Client) Lookup(ethAddress string) (pgpPublicKey string, err error) {
	client := &http.Client{}
	client.Timeout = time.Duration(time.Second * 10)
	var resp *http.Response
	resp, err = client.Get(fmt.Sprintf("%s/pks/lookup?search=%s", me.url, ethAddress))
	if resp == nil && err == nil {
		err = os.ErrInvalid
		return
	}
	if err == nil && resp != nil {
		if resp.StatusCode == 200 {
			var bts []byte
			bts, err = ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				return
			}
			pgpPublicKey = string(bts)
			return
		} else if resp.StatusCode == 404 {
			err = os.ErrNotExist
			return
		}
	}
	err = os.ErrInvalid
	return
}
