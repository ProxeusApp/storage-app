package api

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/ProxeusApp/storage-app/dapp/api/endpoints"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/ProxeusApp/storage-app/dapp/core"
	"github.com/ProxeusApp/storage-app/dapp/core/account"
	"github.com/ProxeusApp/storage-app/dapp/core/file"
	"github.com/ProxeusApp/storage-app/dapp/core/notification"
	"github.com/ProxeusApp/storage-app/spp/config"
	channelhub "github.com/ProxeusApp/storage-app/web"
)

func MainApi(e *echo.Echo) (*channelhub.ChannelHub, *core.App) {
	cfg := &config.Config
	var err error

	if cfg.PprofDebug {
		go func() {
			runtime.SetBlockProfileRate(1)
			log.Println("enabling profiling server on localhost:6060")
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	//TODO Change this when production ready
	//Temporary solution to allow front-end requests
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:8080"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderXRequestedWith},
	}))

	//e.Use(middleware.Logger())
	//e.Use(middleware.Recover())

	chanHub := &channelhub.ChannelHub{
		ChannelFind: func(cMsg *channelhub.ChannelHubMsg) (chnl *channelhub.Channel, create bool) {
			chnl = nil
			create = true
			return
		},
		//ChannelDataFind: func(cMsg *channelhub.ChannelHubMsg) {
		//	//TODO fill cMsg.Data
		//},
		//ChannelCreated: func(chanl *channelhub.Channel, client *channelhub.Client) {
		//	fmt.Println("channel created: ", chanl.ID, client)
		//},
		//ChannelRemoved: func(chanl *channelhub.Channel, client *channelhub.Client) {
		//	fmt.Println("channel removed: ", chanl.ID, client)
		//},
	}

	app, err := core.NewApp(&config.Config, chanHub, time.Minute*30)
	if err != nil {
		panic(err)
	}
	endpoints.App = app // Inject app

	err = chanHub.Run(
		&channelhub.Channel{
			ID:    "global",
			Owner: "dapp",
			//BeforeBroadcast: func(chanl *channelhub.Channel, msg *channelhub.ChannelHubMsg) bool {
			//	fmt.Println("before broadcast: ", msg.ChannelID, msg.Data)
			//	return true
			//},
		},
		&channelhub.Channel{
			ID: "sys",
			//BeforeBroadcast: func(chanl *channelhub.Channel, msg *channelhub.ChannelHubMsg) bool {
			//	fmt.Println("before broadcast: ", msg.ChannelID, msg.Data)
			//	return true
			//},
			System: true,
		},
	)

	jsonApi := e.Group("/api")
	jsonApi.POST("/login", func(c echo.Context) error {
		params := struct {
			ETHAddr string `json:"ethAddr"`
			PW      string `json:"pw"`
		}{}
		if err := c.Bind(&params); err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		err := app.Login(params.ETHAddr, params.PW)
		if err != nil {
			log.Println("[MainApi] error while login:", err)
			return c.JSON(http.StatusUnauthorized, err)
		}
		return c.JSON(http.StatusOK, params.ETHAddr)
	})
	jsonApi.POST("/account/import/eth", func(c echo.Context) error {
		params := struct {
			AccountName string `json:"accountName"`
			ETHPriv     string `json:"ethPriv"`
			PW          string `json:"pw"`
		}{}
		if err := c.Bind(&params); err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		err := app.LoginWithETHPriv(params.ETHPriv, params.AccountName, params.PW)
		if err != nil {
			if os.IsPermission(err) {
				return c.JSON(http.StatusUnauthorized, err.Error())
			}
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		return c.JSON(http.StatusOK, app.GetActiveAccountETHAddress())
	})
	jsonApi.GET("/account/balance", func(c echo.Context) error {
		app.UpdateAccountInfo()
		return c.NoContent(http.StatusNoContent)
	})

	jsonApi.POST("/account/import/eth-and-pgp", func(c echo.Context) error {
		params := struct {
			AccountName string `json:"accountName"`
			ETHPriv     string `json:"ethPriv"`
			PGPPw       string `json:"pgpPw"`
			PGPPriv     string `json:"pgpPriv"`
			PW          string `json:"pw"`
		}{}
		if err := c.Bind(&params); err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		err := app.LoginWithETHPrivAndPGPPriv(params.ETHPriv, params.AccountName, params.PW, params.PGPPriv, params.PGPPw)
		if err != nil {
			if os.IsPermission(err) {
				return c.JSON(http.StatusUnauthorized, err.Error())
			}
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		return c.JSON(http.StatusOK, app.GetActiveAccountETHAddress())
	})

	jsonApi.PUT("/account", func(c echo.Context) error {
		accountInfoRepresentation := struct {
			Name string
			PW   string
		}{}
		if err := c.Bind(&accountInfoRepresentation); err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		err := app.LoginWithNew(accountInfoRepresentation.Name, accountInfoRepresentation.PW)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, err)
		}
		return c.JSON(http.StatusOK, app.GetActiveAccountETHAddress())
	})

	// Updates an account
	jsonApi.POST("/account/:address", func(c echo.Context) error {
		accountInfo := core.AccountInfo{}
		address := c.Param("address")
		if err := c.Bind(&accountInfo); err != nil {
			return err
		}
		accountInfo.Address = address
		err = app.UpdateAccount(&accountInfo)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}
		app.UpdateAccountInfo()
		return c.NoContent(http.StatusNoContent)
	})

	jsonApi.POST("/XESAmountPerFile", func(c echo.Context) error {
		params := struct {
			Providers []string `json:"providers"`
		}{}
		if err := c.Bind(&params); err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		am, err := app.XESAmountPerFile(params.Providers)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		return c.JSON(http.StatusOK, am.String())
	})

	jsonApi.POST("/approveXESToContract/estimateGas", func(c echo.Context) error {
		params := struct {
			XESValue string `json:"xesValue"`
		}{}
		if err := c.Bind(&params); err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}

		gasEstimate, err := app.ApproveXESToContractEstimateGas(params.XESValue)
		if err != nil {
			log.Print("couldn't estimate gas for approve XES ", err)
			return c.JSON(http.StatusBadRequest, errors.New("couldn't estimate gas for approve XES"))
		}

		return c.JSON(http.StatusOK, gasEstimate)
	})

	jsonApi.POST("/approveXESToContract", func(c echo.Context) error {
		params := struct {
			XESValue string `json:"xesValue"`
		}{}
		if err := c.Bind(&params); err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		fmt.Println("approve", params.XESValue)
		err := app.ApproveXESToContract(params.XESValue)
		if err != nil {
			if os.IsPermission(err) {
				return c.JSON(http.StatusUnauthorized, err)
			}
			return c.JSON(http.StatusBadRequest, err)
		}
		return c.JSON(http.StatusOK, app.GetActiveAccountETHAddress())
	})

	jsonApi.POST("/sendXES/estimateGas", func(c echo.Context) error {
		params := struct {
			XESAmount    string `json:"xesAmount"`
			EthAddressTo string `json:"ethAddressTo"`
		}{}
		if err := c.Bind(&params); err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}

		gasEstimate, err := app.SendXESEstimateGas(params.EthAddressTo, params.XESAmount)
		if err != nil {
			log.Print("couldn't estimate gas for send XES ", err)
			return c.JSON(http.StatusBadRequest, errors.New("couldn't estimate gas for send XES"))
		}

		return c.JSON(http.StatusOK, gasEstimate)
	})

	jsonApi.POST("/sendXES", func(c echo.Context) error {
		params := struct {
			XESAmount    string `json:"xesAmount"`
			EthAddressTo string `json:"ethAddressTo"`
		}{}
		if err := c.Bind(&params); err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		err := app.SendXES(params.EthAddressTo, params.XESAmount)
		if err != nil {
			if os.IsPermission(err) {
				return c.JSON(http.StatusUnauthorized, err)
			}
			return c.JSON(http.StatusBadRequest, err)
		}
		return c.NoContent(http.StatusOK)
	})

	jsonApi.POST("/sendETH/estimateGas", func(c echo.Context) error {
		params := struct {
			ETHAmount    string `json:"ethAmount"`
			EthAddressTo string `json:"ethAddressTo"`
		}{}
		if err := c.Bind(&params); err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}

		gasEstimate, err := app.SendETHEstimateGas(params.EthAddressTo, params.ETHAmount)
		if err != nil {
			log.Print("couldn't estimate gas for send ETH ", err)
			return c.JSON(http.StatusBadRequest, errors.New("couldn't estimate gas for send ETH"))
		}

		return c.JSON(http.StatusOK, gasEstimate)
	})

	jsonApi.POST("/sendETH", func(c echo.Context) error {
		params := struct {
			ETHAmount    string `json:"ethAmount"`
			EthAddressTo string `json:"ethAddressTo"`
		}{}
		if err := c.Bind(&params); err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		err := app.SendETH(params.EthAddressTo, params.ETHAmount)
		if err != nil {
			if os.IsPermission(err) {
				return c.JSON(http.StatusUnauthorized, err)
			}
			return c.JSON(http.StatusBadRequest, err)
		}
		return c.NoContent(http.StatusOK)
	})

	jsonApi.GET("/process/drop/:fileHash", func(c echo.Context) error {
		fileHash := c.Param("fileHash")
		defer app.RemoveFileFromDiskKeepMeta(fileHash)
		localpath, _ := app.GetFile(fileHash)
		if err != nil {
			return c.JSON(http.StatusNotFound, err.Error())
		}
		location := config.Config.MainHostedURL
		dropID, err := app.DropFile(location, localpath)
		if err != nil {
			return c.JSON(http.StatusNotFound, err.Error())
		}
		return c.JSON(http.StatusOK, location+"/api/document/"+dropID+"/import")
	})

	jsonApi.POST("/process/share", func(c echo.Context) error {
		form, err := c.MultipartForm()
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		fileUploadRequest, err := endpoints.ParseFileUpload(c)

		reg := file.Register{}
		link := form.Value["link"][0]
		filelocation := filepath.Join(cfg.StorageDir, "sharefiles", link)

		_, _, recipients, err := core.ParseProxeusProtocol(link)
		if err != nil {
			fmt.Println("Protocol error")
			return c.JSON(http.StatusBadRequest, err)
		}
		f, err := os.Open(filelocation)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		reg.FileName = link
		reg.FileKind = 2
		reg.FileReader = f
		reg.DurationDays = fileUploadRequest.Register.DurationDays

		err = app.ArchiveFileAndRegister(reg, nil, 0, fileUploadRequest.ProviderInfo, recipients)
		if err != nil {
			if os.IsPermission(err) {
				return c.JSON(http.StatusUnauthorized, err.Error())
			}
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		return c.NoContent(http.StatusOK)
	})

	// account group
	jsonApi.POST("/account/import", endpoints.AccountImportAndLogin)
	jsonApi.POST("/account/remove", endpoints.AccountRemove)
	jsonApi.POST("/account/export", endpoints.AccountExportByAddress)
	jsonApi.GET("/account/export", endpoints.AccountExport)

	// /file group
	jsonApi.GET("/file/download/:fileHash", endpoints.FileDownload)
	jsonApi.GET("/file/thumb/:fileHash", endpoints.FileDownloadThumb)
	jsonApi.GET("/file/list", endpoints.FileList)
	jsonApi.GET("/file/sign/estimateGas/:fileHash", endpoints.FileSignEstimateGas)
	jsonApi.GET("/file/sign/:fileHash", endpoints.FileSign)
	jsonApi.GET("/file/remove/estimateGas/:fileHash", endpoints.FileRemoveEstimateGas)
	jsonApi.POST("/file/remove/:fileHash", endpoints.FileRemove)
	jsonApi.POST("/file/removeLocal/:fileHash", endpoints.FileRemoveLocal)
	jsonApi.POST("/file/removeDiskKeepMeta/:fileHash", endpoints.FileRemoveFromDiskKeepMeta)
	jsonApi.POST("/file/share/estimateGas/:fileHash", endpoints.FileShareEstimateGas)
	jsonApi.POST("/file/share/:fileHash", endpoints.FileShare)
	jsonApi.POST("/file/sendSigningRequest/estimateGas/:fileHash", endpoints.FileSigningRequestEstimateGas)
	jsonApi.POST("/file/sendSigningRequest/:fileHash", endpoints.FileSigningRequest)
	jsonApi.POST("/file/revoke/estimateGas/:fileHash", endpoints.FileRevokeHashEstimateGas)
	jsonApi.POST("/file/revoke/:fileHash", endpoints.FileRevokeHash)
	jsonApi.POST("/file/new/estimateGas", endpoints.NewFileEstimateGas)
	jsonApi.POST("/file/new", endpoints.NewFile)
	jsonApi.POST("/file/quote", endpoints.FileQuote)

	jsonApi.GET("/isUnlocked", func(c echo.Context) error {
		return c.JSON(http.StatusOK, app.HasActiveAndUnlockedAccount())
	})
	jsonApi.GET("/contacts", func(c echo.Context) error {
		list, err := app.Contacts()
		if err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		return c.JSON(http.StatusOK, list)
	})
	jsonApi.GET("/contacts/find/:ethAddr", func(c echo.Context) error {
		contact := app.ContactFind(c.Param("ethAddr"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		return c.JSON(http.StatusOK, contact)
	})

	// Creates a new contact. Returns error if already existing
	jsonApi.PUT("/contact", func(c echo.Context) error {
		abr := &AddressBookRepresentation{}
		if err := c.Bind(&abr); err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		contact, err := app.ContactCreate(abr.Name, abr.ETHAddress, abr.PGPPublicKey)
		if err != nil {
			switch err.(type) {
			case *account.ErrAddressExists:
				return c.NoContent(http.StatusConflict)
			default:
				return c.JSON(http.StatusInternalServerError, err.Error())
			}
		}
		return c.JSON(http.StatusCreated, contact)
	})

	// Updates a contact if existing
	jsonApi.POST("/contact", func(c echo.Context) error {
		abr := &AddressBookRepresentation{}
		if err := c.Bind(&abr); err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		ade, err := app.ContactUpdate(abr.Name, abr.ETHAddress, abr.PGPPublicKey)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, ade)
	})

	jsonApi.DELETE("/contact/:ethAddr", func(c echo.Context) error {
		err := app.ContactRemove(c.Param("ethAddr"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		return c.NoContent(http.StatusOK)
	})
	jsonApi.GET("/providers", func(c echo.Context) error {
		list, err := app.GetStorageProviders()
		if err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		return c.JSON(http.StatusOK, list)
	})
	jsonApi.POST("/ping", func(c echo.Context) error {
		log.Println("[api][base] /api/ping called.")
		err := app.SignalUserActivity()
		if err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		return c.String(http.StatusOK, "pong")
	})
	jsonApi.POST("/logout", func(c echo.Context) error {
		return c.JSON(http.StatusOK, app.Logout())
	})
	jsonApi.GET("/accounts", func(c echo.Context) error {
		return c.JSON(http.StatusOK, app.ListAccounts())
	})
	jsonApi.GET("/versions", func(c echo.Context) error {
		v, err := app.Versions()
		if err != nil {
			log.Print("update error ", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.JSON(http.StatusOK, v)
	})
	jsonApi.POST("/update/download", func(c echo.Context) error {
		err := app.DownloadUpdate()
		if err != nil {
			log.Print("update download error ", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	})
	jsonApi.POST("/update/apply", func(c echo.Context) error {
		err := app.ApplyUpdate()
		if err != nil {
			log.Print("update apply error ", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	})
	jsonApi.GET("/config/blockchainnet", func(c echo.Context) error {
		return c.JSON(http.StatusOK, &config.Config.BlockchainNet)
	})
	jsonApi.DELETE("/notification/remove/:id", func(c echo.Context) error {
		id := c.Param("id")
		err := app.RemoveNotification(id)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		return c.JSON(http.StatusOK, id)
	})
	jsonApi.PUT("/notification/markAllAsRead", func(c echo.Context) error {
		err := app.MarkAllNotificationsAsRead()
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		return c.JSON(http.StatusOK, nil)
	})
	jsonApi.PUT("/notification/update", func(c echo.Context) error {
		n := notification.Notification{}
		if err := c.Bind(&n); err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		err := app.UpdateNotification(n)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		return c.JSON(http.StatusOK, n.ID)
	})

	//Channelhub connection
	e.GET("/pipe", func(c echo.Context) error {
		w := c.Response().Writer
		r := c.Request()
		if app.HasActiveAccount() { //allow connection only if there is an active account
			fmt.Println("pipe: active account found. new channelhub will be initialized")
			err = chanHub.NewClient(w, r, "sessionID", "dapp", "dapp")
			if err == nil {
				return nil
			} else {
				fmt.Println(err)
			}
		} else {
			fmt.Println("pipe: no active account. no channelhub will be initialized")
			return c.NoContent(http.StatusUnauthorized)
		}
		// Check that the rw can be hijacked.
		hj, ok := w.(http.Hijacker)

		// The rw can't be hijacked, return early.
		if !ok {
			//http.Error(w, "can't hijack rw", http.StatusInternalServerError)
			return nil
		}
		// Hijack the rw.
		conn, _, err := hj.Hijack()
		if err != nil {
			// handle error
			return nil
		}
		// Close the hijacked raw tcp connection.
		if err = conn.Close(); err != nil {
			// handle error
			return nil
		}
		return nil

	})

	// TODO update i18n package to be able to use JSON files as storage and use that here.
	// For the dapp we can't use a mysqlDB
	// i18n.AttachWebAPI(jsonApi, jsonApi)
	return chanHub, app
}

func provisionFileHeaders(resp http.ResponseWriter, filePath string, inline bool) {
	inlineOrAttachment := "attachment"
	if inline {
		inlineOrAttachment = "inline"
	}
	fileName := filepath.Base(filePath)
	contentDisposition := fmt.Sprintf(`%s; filename="%s"`, inlineOrAttachment, url.QueryEscape(fileName))
	resp.Header().Set("Content-Disposition", contentDisposition)
	resp.Header().Set("Cache-Control", "no-store")
}
