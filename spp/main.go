package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ProxeusApp/storage-app/dapp/core/ethereum"
	"github.com/ProxeusApp/storage-app/spp/service"

	"github.com/labstack/echo"

	"github.com/ProxeusApp/storage-app/lib/default_server"
	"github.com/ProxeusApp/storage-app/spp/config"
	"github.com/ProxeusApp/storage-app/spp/endpoint"
	"github.com/ProxeusApp/storage-app/spp/fs"
)

const settingsFilename = "settings.json"

var cfg *config.Configuration
var sppWorkerRunning chan bool

func main() {
	config.Setup()
	var err error
	cfg = &config.Config

	wd, err := os.Getwd()
	if err != nil {
		log.Panic(err)
	}
	filePath := filepath.Join(wd, settingsFilename)
	providerInfoService, err := service.NewProviderInfoService(filePath)
	if err != nil {
		log.Panic(err)
	}

	if config.Config.IsTestMode() {
		log.Println("#######################################################")
		log.Println("# STARTING SPP IN TEST MODE - NOT FOR PRODUCTION #")
		log.Println("#######################################################")
	}

	c, _ := json.MarshalIndent(providerInfoService.Get(), "", "  ")
	log.Println("providerInfoService defined configuration:")
	log.Println(string(c))

	log.Printf("Starting new spp with config: EthClientURL: %s, StorageProviderAddress: %s, EthWebSocketURL: %s, StorageDir: %s, ContractAddress: %s",
		cfg.EthClientURL, cfg.StorageProviderAddress, cfg.EthWebSocketURL, cfg.StorageDir, cfg.ContractAddress)
	ethClient, err := ethereum.NewSppClient(cfg.EthClientURL, cfg.EthWebSocketURL, cfg.StorageDir, cfg.ContractAddress)
	if err != nil {
		log.Panic(err)
	}

	if err = ethClient.InitListeners(cfg.StorageDir, cfg.StorageProviderAddress, eventsHandler); err != nil {
		log.Panic(err)
	}

	fileMetaHandler, err := fs.NewFileMetaHandler(cfg.StorageDir)
	if err != nil {
		log.Panic(err)
	}

	// Pass dependencies to `endpoint`
	endpoint.EthClient = ethClient
	endpoint.ProxeusFS, err = fs.NewProxeusFS(cfg, ethClient, fileMetaHandler, providerInfoService)
	endpoint.ServiceProviderInfo = providerInfoService.Get()
	if err != nil {
		log.Panic(err)
	}

	e := newEcho()

	proxeusFSDeleteEvent, err := fs.NewProxeusFSDeleteEvent(&config.Config)
	if err != nil {
		log.Panic(err)
	}

	proxeusFSDeleteEvent.StartSubscribeDeleteHandler()

	defer stopWorker()
	startWorker()

	default_server.StartServer(e, config.Config.ServiceAddress, config.Config.AutoTLS)
	proxeusFSDeleteEvent.Close()
}

func eventsHandler(lg *types.Log, _ bool) error {
	payRecv := endpoint.EthClient.LogAsPaymentReceived(lg)
	if payRecv != nil && eventInterestingForMe(payRecv.StorageProvider) {
		log.Printf("Event[PaymentReceived] %v %v incoming... tx %s fileHash %s, xesAmount: %v\n", lg.BlockNumber, lg.TxIndex, lg.TxHash.Hex(), common.Hash(payRecv.Hash).Hex(), payRecv.XesAmount)
		endpoint.EthClient.HandlePaymentReceivedEvent(payRecv.Hash, payRecv.XesAmount)
	}
	return nil
}

func eventInterestingForMe(provAddress common.Address) bool {
	return cfg.StorageProviderAddress == provAddress.String()
}

func newEcho() *echo.Echo {
	e := default_server.Setup("/var/log/spp.log")

	e.GET("/challenge", endpoint.GetChallenge)
	e.POST("/:fileHash/:token/:signature", endpoint.PostFile)
	e.GET("/:fileHash/:token/:signature", endpoint.GetFile)
	e.GET("/info", endpoint.Info)
	e.GET("/ping", endpoint.Ping)
	e.GET("/health", endpoint.Health)

	return e
}

func startWorker() {
	log.Println("Start SPP worker...")

	go func() {
		sppWorkerRunning = make(chan bool, 1)
		ticker := time.NewTicker(60 * time.Second)
		var lastExecutedDate string

		defer ticker.Stop()

		for {
			select {
			case t := <-ticker.C:
				tickDateStr := t.Format("20060102")
				if lastExecutedDate != tickDateStr && t.Hour() == 2 && t.Minute() == 0 {
					lastExecutedDate = tickDateStr
					go endpoint.ProxeusFS.CheckForExpiredFiles()
				}
			case <-sppWorkerRunning:
				return
			}
		}
	}()
}

func stopWorker() {
	log.Println("Stopping SPP worker...")
	close(sppWorkerRunning)
}
