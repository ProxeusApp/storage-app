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

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"git.proxeus.com/core/central/dapp/api"
	"git.proxeus.com/core/central/spp/config"
)

func main() {
	serverAddress := flag.String("serverAddress", ":8081", "host:port")

	flag.Parse()

	config.Setup()

	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:8080"},
		//AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderXRequestedWith},
	}))

	{
		usr, err := user.Current()
		if err != nil {
			log.Println(err)
		}
		config.Config.StorageDir = filepath.Join(usr.HomeDir, ".proxeus-data")
	}

	//api.MainApi(e)
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
