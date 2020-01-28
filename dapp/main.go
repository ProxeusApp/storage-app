package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/ProxeusApp/storage-app/dapp/core/updater"
	"github.com/asticode/go-astilectron"
	"github.com/asticode/go-astilog"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"

	"github.com/ProxeusApp/storage-app/artifacts/dapp/resources"
	"github.com/ProxeusApp/storage-app/artifacts/dapp/spp"
	"github.com/ProxeusApp/storage-app/dapp/api"
	"github.com/ProxeusApp/storage-app/dapp/embedded"
	"github.com/ProxeusApp/storage-app/spp/config"
)

//go:generate go-bindata -pkg spp -o ../artifacts/dapp/spp/bindata.go -prefix ../artifacts/dapp/dist ../artifacts/dapp/dist/...
//go:generate go-bindata -pkg resources -o ../artifacts/dapp/resources/bindata.go  -prefix resources ./resources/...

var sessionCookieStore = sessions.NewCookieStore([]byte("secret_Dummy_1234"), []byte("12345678901234567890123456789012"))
var readFile = ioutil.ReadFile

const localAddr = "127.0.0.1:56535"
const appName = "Proxeus"

var closeOnce sync.Once

func main() {
	config.Setup()

	usr, err := user.Current()
	if err != nil {
		log.Println(err)
	}
	dataDir := filepath.Join(usr.HomeDir, ".proxeus-data")
	copyResources(dataDir)

	var astiloglogger astilog.Logger
	astilogger := logrus.New()
	astilogger.Formatter = &logrus.TextFormatter{DisableColors: true}
	astilogger.Level = logrus.DebugLevel
	astiloglogger = astilogger
	astilog.SetLogger(astiloglogger)

	astilog.Debug("starting")

	defer func() {
		err := recover()
		if err != nil {
			astilog.Error(err)
		}
	}()

	a := electronInstance(dataDir)
	if !isLocalAddrAvailable() {
		showWindowWithError(a,
			fmt.Sprintf("Error while starting %s", appName),
			fmt.Sprintf("Couldn\\'t start the application because \"%s\" is already in use.\\nEither the application is already running, or another application is using the same port.", localAddr))
		a.Wait()
		return
	}

	if config.Config.StorageDir == "" || config.Config.StorageDir == "./" {
		config.Config.StorageDir = a.Paths().DataDirectory()
	}
	multilogwriter := setupMultiLogWriter(filepath.Join(config.Config.StorageDir, "proxeus.log"))
	astilogger.Out = multilogwriter
	e := echo.New()
	chann, app := api.MainApi(e)

	e.Logger.SetOutput(multilogwriter)
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{Output: multilogwriter}))

	closeFunc := func() {
		chann.Close()
		app.Close()
		a.Close()
		e.Close()
	}

	defer func() {
		closeOnce.Do(closeFunc)
	}()

	go func() {
		defer func() {
			err := recover()
			if err != nil {
				astilog.Error(err)
			}
			closeOnce.Do(closeFunc)
		}()

		startElectron(a)
		a.Wait()
	}()

	ch := make(chan os.Signal)
	go func() {
		_, ok := <-ch
		if ok {
			closeOnce.Do(closeFunc)
		}
	}()
	signal.Notify(ch, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	go updater.Cleanup()
	updater.OnClose(closeFunc)

	hookEndpoints(e)

	//api.MainApi(e)
	err = e.Start(localAddr)
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}

func copyResources(dataDir string) {
	resDir := filepath.Join(dataDir, "res")
	if err := os.MkdirAll(resDir, 0755); err != nil {
		panic(err)
	}
	for _, fname := range []string{"icon.png", "icon.icns"} {
		b, err := resources.Asset(fname)
		if err != nil {
			panic(err)
		}
		err = ioutil.WriteFile(filepath.Join(resDir, fname), b, 0644)
		if err != nil {
			// TODO: in OSX dmg these files are read-only, we need to find solution
			astilog.Warn(err)
		}
	}
}

func electronInstance(dataDir string) *astilectron.Astilectron {
	a, err := astilectron.New(astilectron.Options{
		AppName:            appName,
		AppIconDefaultPath: "res/icon.png",
		AppIconDarwinPath:  "res/icon.icns",
		DataDirectoryPath:  dataDir,
		BaseDirectoryPath:  dataDir,
	})
	if err != nil {
		panic(err)
	}
	vendorDir := a.Paths().VendorDirectory()
	astilog.Debug("vendor directory ", vendorDir)

	if err := os.MkdirAll(vendorDir, 0755); err != nil {
		panic(err)
	}

	resources := []struct {
		Filename string
		Asset    string
	}{
		{
			Filename: fmt.Sprintf("astilectron-v%s.zip", astilectron.VersionAstilectron),
			Asset:    "vendor_astilectron_bundler/astilectron.zip",
		},
		{
			Filename: fmt.Sprintf("electron-%s-%s-v%s.zip", runtime.GOOS, runtime.GOARCH, astilectron.VersionElectron),
			Asset:    "vendor_astilectron_bundler/electron.zip",
		},
	}

	for _, r := range resources {
		if _, err := os.Stat(r.Filename); err == nil {
			continue
		}
		err := ioutil.WriteFile(filepath.Join(vendorDir, r.Filename), MustAsset(r.Asset), 0644)
		if err != nil {
			astilog.Warn(err)
		}
	}
	return a
}

func startElectron(a *astilectron.Astilectron) {
	err := a.Start()
	if err != nil {
		panic(err)
	}
	w, err := a.NewWindow(fmt.Sprintf("http://%s/", localAddr), &astilectron.WindowOptions{
		Title:    astilectron.PtrStr(appName),
		Center:   astilectron.PtrBool(true),
		Height:   astilectron.PtrInt(800),
		Width:    astilectron.PtrInt(1200), //1200 because with 1000 windows login tour step 1 view is broken
		MinWidth: astilectron.PtrInt(900),
		//inject electron and current os into window (see: https://github.com/asticode/astilectron/blob/v0.27.0/main.js#L483)
		Custom: &astilectron.WindowCustomOptions{
			Script: fmt.Sprintf("const electron = require('electron'); const osPlatform = '%s'", runtime.GOOS),
		},
	})
	if err != nil {
		panic(err)
	}
	err = w.Create()
	if err != nil {
		panic(err)
	}
	if config.Config.DevMode {
		w.OpenDevTools()
	}
}

func showWindowWithError(a *astilectron.Astilectron, title, message string) {
	err := a.Start()
	if err != nil {
		panic(err)
	}

	w, err := a.NewWindow("http://127.0.0.1", &astilectron.WindowOptions{
		Title:  astilectron.PtrStr(appName),
		Center: astilectron.PtrBool(true),
		Width:  astilectron.PtrInt(640),
		Height: astilectron.PtrInt(480),
		Custom: &astilectron.WindowCustomOptions{
			Script: fmt.Sprintf("astilectron.showErrorBox('%s', '%s');", title, message),
		},
	})
	if err != nil {
		panic(err)
	}

	err = w.Create()
	if err != nil {
		panic(err)
	}
}

func hookEndpoints(e *echo.Echo) {
	e.HideBanner = true
	embed := &embedded.Embedded{Asset: spp.Asset}
	readFile = func(path string) ([]byte, error) {
		return embed.Asset2(path)
	}

	e.GET("/static/*", func(c echo.Context) error {
		url := c.Request().URL.String()
		i := strings.LastIndex(url, "?")
		if i != -1 {
			url = url[:i]
		}
		ct := ""
		b, err := embed.FindAssetWithCT(url, &ct)
		log.Println(url, ct)
		if err == nil {
			return c.Blob(http.StatusOK, ct, b)
		}
		return echo.ErrNotFound
	})

	e.GET("/", indexHandler)
}

func indexHandler(c echo.Context) error {
	appBts, err := readFile("view/index.html")
	if err != nil {
		return c.NoContent(http.StatusNotFound)
	}
	return c.HTMLBlob(http.StatusOK, appBts)
}

func setupMultiLogWriter(logFileLocation string) io.Writer {
	var multiLogWriter io.Writer

	multiLogWriter = os.Stdout

	lumberjackLogger := &lumberjack.Logger{
		Filename: logFileLocation,
		MaxSize:  50, // MB
		MaxAge:   7,  // days
	}

	_, err := lumberjackLogger.Write([]byte("log init\n"))
	if err != nil {
		log.Printf("File logging disabled due to: <%s>\n", err)
	} else {
		multiLogWriter = io.MultiWriter(lumberjackLogger, os.Stdout)

		log.Printf("Logging to: %s\n", logFileLocation)
		log.SetOutput(multiLogWriter)
	}
	return multiLogWriter
}

func isLocalAddrAvailable() bool {
	conn, err := net.DialTimeout("tcp", localAddr, 500*time.Millisecond)
	if err != nil {
		log.Printf("[main][isLocalAddrAvailable] %s is available", localAddr)
		return true
	}

	_ = conn.Close()
	log.Printf("[main][isLocalAddrAvailable] %s already in use", localAddr)
	return false
}
