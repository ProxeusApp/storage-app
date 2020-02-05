package main

import (
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/ProxeusApp/storage-app/dapp/core/util"

	"log"

	"github.com/ethereum/go-ethereum/crypto"
	"gopkg.in/gavv/httpexpect.v2"

	"github.com/ProxeusApp/storage-app/dapp/core/ethereum"
	"github.com/ProxeusApp/storage-app/spp/config"
	"github.com/ProxeusApp/storage-app/spp/fs"
	"github.com/ProxeusApp/storage-app/spp/service"
)

func serverURL() string {
	u := os.Getenv("SERVER_URL")
	if u != "" {
		return u
	}
	server := httptest.NewServer(newEcho())
	return server.URL
}

var challenge *httpexpect.String
var token *httpexpect.String

func TestChallenge(t *testing.T) {
	//e := httpexpect.New(t, serverURL())
	//j := e.GET("/challenge").Expect().Status(http.StatusOK).JSON()
	//challenge = j.Object().Value("challenge").String()
	//token = j.Object().Value("token").String()
	//challenge.Match("^0x[0-9a-f]{188}$")
	//token.Match("^[0-9a-f\\-]{36}$")

	eventSignature := []byte("Deleted(bytes32)")
	esHash := crypto.Keccak256Hash(eventSignature)
	log.Println(esHash.Bytes(), string(esHash.Bytes()), esHash.String())
}

// func TestPostFile(t *testing.T) {
// 	e := httpexpect.New(t, serverURL())
// 	signature := "0x00000000000000000000000000000000000000000"
// 	path := fmt.Sprintf("/xxxxxxxxxxxxxxxxxxxxxxx/%s/%s", token.Raw(), signature)
// 	e.POST(path).Expect().Status(http.StatusOK).NoContent()
// }

func BenchmarkChallenge(b *testing.B) {
	e := httpexpect.New(b, serverURL())
	for i := 0; i < b.N; i++ {
		e.GET("/challenge").Expect().Status(http.StatusOK)
	}
}

func TestCheckForExpiredFiles(t *testing.T) {
	config.Setup()
	cfg := &config.Config

	providerInfoService, err := service.NewProviderInfoService("./settings.json")
	if err != nil {
		t.Fatal(err)
	}
	graceSeconds := int64(providerInfoService.Get().GraceSeconds)
	t.Log("Grace seconds:", graceSeconds)

	proxeusFSFileMocks := []*fs.FileInfoMock{
		{ // expired (should delete, expecting an error while trying to remove the file from disk)
			Id:     util.StrHexToBytes32("0x822ac138637485893a49980082b5dfbb020c14f20017f4ae69c7f350c06fe8c1"),
			Expiry: big.NewInt(time.Now().Add(-1*time.Hour).Unix() - graceSeconds),
		},
		{ // expired but in grace period (shouldn't delete)
			Id:     util.StrHexToBytes32("0x7df17c9bc5c29772556b178c5df73bdd8d3991943196ea8e0e331ecc1c43a908"),
			Expiry: big.NewInt(time.Now().Unix() - (graceSeconds / 2)),
		},
		{ // not expired (shouldn't delete)
			Id:     util.StrHexToBytes32("0x9fccf1e7169d31dc8e19473891e19fb8eff3c580b82e72f00c04c8864541cae3"),
			Expiry: big.NewInt(time.Now().Add(1 * time.Hour).Unix()),
		},
		{ // not expired, but will be manipulated to be (shouldn't delete)
			Id:     util.StrHexToBytes32("0xcd9afd2b4c1c663b96a19c1405d437ee760ccb6eda35eb7d81ccf2b9faa05ae9"),
			Expiry: big.NewInt(time.Now().Add(1 * time.Hour).Unix()),
		},
		//{
		//	Id:     ethereum.StrHexToBytes32("0xfd62d66896b2e571df8830cd455da537737ead7108cf841b49fc14fcafa13a8c"),
		//	Expiry: big.NewInt(time.Now().Add(-1 * time.Hour).Unix()),
		//},
		//{
		//	Id:     ethereum.StrHexToBytes32("0x74432e345d6f947400effefe72519370ae532b9b26528e50005e9f214eb70acc"),
		//	Expiry: big.NewInt(time.Now().Add(-1 * time.Hour).Unix()),
		//},
		//{
		//	Id:     ethereum.StrHexToBytes32("0x523b2792b0a90020a4f37e9218411e4d45fc65452c3e70cab45e804899374ad8"),
		//	Expiry: big.NewInt(time.Now().Add(-1 * time.Hour).Unix()),
		//},
		//{
		//	Id:     ethereum.StrHexToBytes32("0x735bc371ffd418780b1b5fab7b88e369b56e2c73af22971f6874532f0cb849ee"),
		//	Expiry: big.NewInt(time.Now().Add(-1 * time.Hour).Unix()),
		//},
		//{
		//	Id:     ethereum.StrHexToBytes32("0xf3ace7644cf4436ad030ea89ad7854a49a8e8814bb2c987b9868bb9c471eeb30"),
		//	Expiry: big.NewInt(time.Now().Add(-1 * time.Hour).Unix()),
		//},
	}

	fsFileMocks := fs.NewFileMetaClientMock(proxeusFSFileMocks)
	ethClientMock := ethereum.NewClientMock(fsFileMocks)

	proxeusFS, err := fs.NewProxeusFS(cfg, ethClientMock, fsFileMocks, providerInfoService)
	if err != nil {
		t.Fatal(err)
	}

	manipulatedFileInfo := fs.FileInfoMock{
		Id:     util.StrHexToBytes32("0xcd9afd2b4c1c663b96a19c1405d437ee760ccb6eda35eb7d81ccf2b9faa05ae9"),
		Expiry: big.NewInt(time.Now().Add(-1 * time.Hour).Unix()),
	}
	fsFileMocks.Manipulate(manipulatedFileInfo)

	proxeusFS.CheckForExpiredFiles()

	_, err = fsFileMocks.Get(util.StrHexToBytes32("0x822ac138637485893a49980082b5dfbb020c14f20017f4ae69c7f350c06fe8c1"))
	if err != fs.ErrSppFileMetaNotFound {
		t.Error("0x822ac138637485893a49980082b5dfbb020c14f20017f4ae69c7f350c06fe8c1 was expired, but not deleted. -", err)
	}

	fileInGracePeriod, err := fsFileMocks.Get(util.StrHexToBytes32("0x7df17c9bc5c29772556b178c5df73bdd8d3991943196ea8e0e331ecc1c43a908"))
	if fileInGracePeriod == nil || err != nil {
		t.Error("0x7df17c9bc5c29772556b178c5df73bdd8d3991943196ea8e0e331ecc1c43a908 was in grace period and shouldn't be deleted. -", err)
	}

	fileNotExpired, err := fsFileMocks.Get(util.StrHexToBytes32("0x9fccf1e7169d31dc8e19473891e19fb8eff3c580b82e72f00c04c8864541cae3"))
	if fileNotExpired == nil || err != nil {
		t.Error("0x9fccf1e7169d31dc8e19473891e19fb8eff3c580b82e72f00c04c8864541cae3 was not expired and shouldn't be deleted. -", err)
	}

	manipulatedFileNotExpired, err := fsFileMocks.Get(util.StrHexToBytes32("0xcd9afd2b4c1c663b96a19c1405d437ee760ccb6eda35eb7d81ccf2b9faa05ae9"))
	if manipulatedFileNotExpired == nil || err != nil {
		t.Error("0xcd9afd2b4c1c663b96a19c1405d437ee760ccb6eda35eb7d81ccf2b9faa05ae9 was not expired but manipulated to be, but shouldn't be deleted. -", err)
	}
}
