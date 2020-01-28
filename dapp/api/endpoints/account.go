package endpoints

import (
	"net/http"
	"os"

	"github.com/labstack/echo"

	"github.com/ProxeusApp/storage-app/dapp/core/account"
)

func AccountImportAndLogin(c echo.Context) error {
	form, err := c.MultipartForm()
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	var acc *account.Account
	passwordFormValue := form.Value["password"]
	files := form.File["files"]
	for _, f := range files {
		src, err := f.Open()
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		defer src.Close()
		var password string
		if len(passwordFormValue) > 0 {
			password = passwordFormValue[0]
		}
		var alreadyExists bool
		acc, alreadyExists, err = App.AccountImport(src, password)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		if alreadyExists {
			err = App.Login(acc.GetETHAddress(), password)
		} else {
			err = App.LoginWithImportedKeystore(acc.GetETHAddress(), password)
		}
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
	}
	return c.JSON(http.StatusOK, acc.GetETHAddress())
}

func AccountRemove(c echo.Context) error {
	params := struct {
		ETHAddress string `json:"address"`
		Pw         string `json:"pw"`
	}{}
	if err := c.Bind(&params); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	err := App.AccountRemove(params.ETHAddress, params.Pw)
	if err != nil {
		if os.IsPermission(err) {
			return c.JSON(http.StatusUnauthorized, err.Error())
		}
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.NoContent(http.StatusOK)
}

func AccountExport(c echo.Context) error {
	fp, err := App.AccountExport()
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	provisionFileHeaders(c.Response(), fp, false)
	return c.File(fp)
}

func AccountExportByAddress(c echo.Context) error {
	params := struct {
		ETHAddress string `json:"address"`
		Pw         string `json:"pw"`
	}{}
	if err := c.Bind(&params); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	fp, err := App.AccountExportByAddress(params.ETHAddress, params.Pw)
	if err != nil {
		if os.IsPermission(err) {
			return c.JSON(http.StatusUnauthorized, err.Error())
		}
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	provisionFileHeaders(c.Response(), fp, false)
	return c.File(fp)
}
