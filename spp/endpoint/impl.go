package endpoint

import (
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/ProxeusApp/storage-app/spp/client/models"

	"github.com/labstack/echo"

	"github.com/ProxeusApp/storage-app/dapp/core/ethereum"
	"github.com/ProxeusApp/storage-app/spp/fs"
)

var ProxeusFS *fs.ProxeusFS
var EthClient *ethereum.SppClient
var ServiceProviderInfo models.StorageProviderInfo

func GetChallenge(c echo.Context) error {
	r, err := ProxeusFS.CreateSignInChallenge()
	if err != nil {
		c.Logger().Error(err)
		return c.NoContent(http.StatusBadRequest)
	}
	return c.JSON(http.StatusOK, r)
}

func PostFile(c echo.Context) error {
	body := http.MaxBytesReader(c.Response().Writer, c.Request().Body, ServiceProviderInfo.MaxFileSizeByte) // TODO file validation
	defer body.Close()

	log.Println("spp: Requesting file upload for: ", c.Param("fileHash"))

	duration, err := strconv.Atoi(c.QueryParam("duration"))
	if err != nil {
		return err
	}

	_, err = ProxeusFS.Input(
		c.Param("fileHash"),
		c.Param("token"),
		c.Param("signature"),
		body,
		duration)

	if err != nil {
		c.Logger().Error(err)
		if err == ethereum.ErrFilePaymentNotFound {
			return c.NoContent(http.StatusPaymentRequired)
		}
		return c.NoContent(http.StatusBadRequest)
	}

	c.Logger().Info("spp: successfully uploaded file ", c.Param("fileHash"))
	return c.NoContent(http.StatusOK)
}

func GetFile(c echo.Context) error {
	force := false
	if _, ok := c.QueryParams()["force"]; ok {
		force = true
	}
	p, err := ProxeusFS.Output(
		c.Param("fileHash"),
		c.Param("token"),
		c.Param("signature"),
		force,
	)
	if err != nil {
		c.Logger().Error(err)
		if err == ethereum.ErrFileNotFound {
			// The file hasn't been registered
			return c.NoContent(http.StatusNotFound)
		} else if os.IsNotExist(err) {
			// The file has been registered on the blockchain but the file isn't arrived yet. Try again
			return c.NoContent(http.StatusAccepted)
		} else if err == fs.ErrNoPermission {
			return c.NoContent(http.StatusForbidden)
		} else if err == fs.ErrNotSatisfiable {
			return c.String(http.StatusRequestedRangeNotSatisfiable, err.Error())
		} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return c.NoContent(http.StatusServiceUnavailable)
		} else if err == fs.ErrFileRemoved {
			return c.NoContent(http.StatusGone)
		} else {
			return c.NoContent(http.StatusBadRequest)
		}
	}
	return c.File(p)
}

func Info(c echo.Context) error {
	return c.JSON(http.StatusOK, ServiceProviderInfo)
}

func Ping(c echo.Context) error {
	return c.String(http.StatusOK, "pong")
}
