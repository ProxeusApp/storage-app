package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"time"

	"github.com/ProxeusApp/storage-app/dapp/api"
	"github.com/ProxeusApp/storage-app/spp/config"
	"github.com/labstack/echo"
)

func main() {
	serverAddress := flag.String("serverAddress", ":8081", "host:port")

	flag.Parse()

	config.Setup()

	e := echo.New()
	e.HideBanner = true

	{
		usr, err := user.Current()
		if err != nil {
			log.Println(err)
		}
		if config.Config.IsTestMode() {
			config.Config.StorageDir = filepath.Join(usr.HomeDir, ".proxeus-data-api-test")
		} else {
			config.Config.StorageDir = filepath.Join(usr.HomeDir, ".proxeus-data")
		}
	}

	chann, app := api.MainApi(e)
	// Start server
	go func() {
		if err := e.Start(*serverAddress); err != nil {
			e.Logger.Info("shutting down the server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 10 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
	chann.Close()
	app.Close()
	log.Println("Last printout")
}
