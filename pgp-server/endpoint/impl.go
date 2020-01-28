package endpoint

import (
	"log"
	"net/http"
	"strings"

	"github.com/ProxeusApp/pgp"

	"github.com/labstack/echo"

	"github.com/ProxeusApp/storage-app/lib/airdrop"
	"github.com/ProxeusApp/storage-app/pgp-server/storage"
	"github.com/ProxeusApp/storage-app/spp/fs"
)

var ProxeusFS *fs.ProxeusFS

func AddPublicKey(c echo.Context) error {
	params := struct {
		Pubkey    string `json:"pubkey"`
		Token     string `json:"token"`
		Signature string `json:"signature"`
	}{}

	if err := c.Bind(&params); err != nil {
		return err
	}

	c.Logger().Info("adding pubkey with token ", params.Token)

	if params.Pubkey == "" {
		//InternalServerError - in order to follow openpgp key server specification
		return c.String(http.StatusInternalServerError, "Empty public key")
	}

	addr, err := ProxeusFS.Validate(params.Token, params.Signature)
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	ethereumAddress, err := GetIdentity(params.Pubkey)

	addr = strings.ToLower(addr)
	ethereumAddress = strings.ToLower(ethereumAddress)

	if err != nil || addr != ethereumAddress {
		return c.NoContent(http.StatusUnauthorized)
	}

	_, err = storage.GetPublicKey(ethereumAddress)
	if err != nil {
		if err == storage.ErrNotFound {
			err = nil
			go func() {
				defer func() {
					if r := recover(); r != nil {
						log.Println("airdrop recover with err ", r)
					}
				}()
				airdrop.GiveTokens(ethereumAddress)
			}()
		} else {
			c.Logger().Error(err)
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}

	err = storage.SetPublicKey(ethereumAddress, params.Pubkey)
	if err != nil {
		c.Logger().Error(err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.String(http.StatusOK, ethereumAddress)
}

func GetPublicKey(c echo.Context) error {
	ethAddress := c.QueryParam("search")

	if ethAddress == "" {
		return c.String(http.StatusBadRequest, "Empty ethereum address")
	}

	ethAddress = strings.ToLower(ethAddress)
	publicKey, err := storage.GetPublicKey(ethAddress)

	if err != nil {
		if err == storage.ErrNotFound {
			return c.String(http.StatusNotFound, err.Error())
		}
		c.Logger().Warn(err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.String(http.StatusOK, publicKey)
}

func GetChallenge(c echo.Context) error {
	r, err := ProxeusFS.CreateSignInChallenge()
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, r)
}

func GetIdentity(publicKey string) (ethAddress string, err error) {
	identity, err := pgp.ReadIdentity([][]byte{[]byte(publicKey)})
	if err != nil {
		return
	}
	ethAddress = identity[0]["name"]
	return
}
