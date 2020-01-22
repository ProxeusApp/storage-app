package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"git.proxeus.com/core/central/lib/default_server"
	"git.proxeus.com/core/central/pgp-server/endpoint"
	"git.proxeus.com/core/central/pgp-server/storage"
	"git.proxeus.com/core/central/spp/config"
	"git.proxeus.com/core/central/spp/fs"
)

var storageDir string
var serverAddress string

func main() {
	e := newEcho()
	defer storage.CloseDB()
	// Start server
	e.Logger.Debug(e.Start(serverAddress))
}

func newEcho() *echo.Echo {
	flag.StringVar(&storageDir, "storageDir", "/tmp", "databaseDir directory")
	flag.StringVar(&serverAddress, "serverAddress", ":8080", "host:port")
	flag.Parse()

	e := default_server.Setup("/var/log/pgp.log")

	var err error
	storage.DatabaseDir, err = filepath.Abs(storageDir)
	if err != nil {
		e.Logger.Panic(err)
	}

	_, err = os.Stat(storage.DatabaseDir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(storage.DatabaseDir, 0750)
		if err != nil {
			e.Logger.Panic(err)
		}
	}
	storage.DatabaseDir = filepath.Join(storage.DatabaseDir, "database.db")
	e.Logger.Print("DB path:", storage.DatabaseDir)
	err = storage.OpenDB()
	if err != nil {
		e.Logger.Panic(err)
	}

	endpoint.ProxeusFS, err = fs.NewProxeusFS(&config.Configuration{}, nil, nil, nil) // ETH connection not used
	if err != nil {
		e.Logger.Panic(err)
	}

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
	}))

	// Routes
	e.GET("/pks/challenge", endpoint.GetChallenge)
	e.POST("/pks/add", endpoint.AddPublicKey)
	e.GET("/pks/lookup", endpoint.GetPublicKey)

	return e
}
