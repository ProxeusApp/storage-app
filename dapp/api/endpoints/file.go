package endpoints

import (
	"encoding/json"
	"errors"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"

	"github.com/labstack/echo"

	"github.com/ProxeusApp/storage-app/dapp/core/account"
	"github.com/ProxeusApp/storage-app/dapp/core/file"
	"github.com/ProxeusApp/storage-app/spp/client/models"
)

type (
	// An incoming File upload request should be parsed to FileUploadRequest
	FileUploadRequest struct {
		Register              file.Register
		DefinedSignerList     []account.AddressBookEntry
		UndefinedSignersCount int64
		ProviderInfo          models.StorageProviderInfo
		XesAmount             *big.Int
	}
)

func FileList(c echo.Context) error {
	myFiles := false
	if _, ok := c.QueryParams()["myFiles"]; ok {
		myFiles = true
	}
	sharedWithMe := false
	if _, ok := c.QueryParams()["sharedWithMe"]; ok {
		sharedWithMe = true
	}
	signedByMe := false
	if _, ok := c.QueryParams()["signedByMe"]; ok {
		signedByMe = true
	}
	expiredFiles := false
	if _, ok := c.QueryParams()["expiredFiles"]; ok {
		expiredFiles = true
	}
	filter := c.QueryParam("filter")

	err := App.ListFiles(filter, myFiles, sharedWithMe, signedByMe, expiredFiles)

	if err != nil {
		log.Println(err.Error())
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.NoContent(http.StatusOK)
}

func FileDownload(c echo.Context) error {
	fileHash := c.Param("fileHash")
	fp, err := App.GetFile(fileHash)
	defer App.RemoveFileFromDiskKeepMeta(fileHash)

	if err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}
	provisionFileHeaders(c.Response(), fp, false)
	return c.File(fp)
}

func FileDownloadThumb(c echo.Context) error {
	fileHash := c.Param("fileHash")

	defer App.RemoveFileFromDiskKeepMeta(fileHash)

	fp, err := App.GetThumbnail(fileHash, false)
	if err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}

	resp := c.Response()
	provisionFileHeaders(resp, fp, true)

	contentType, err := detectMimeType(fp)
	if err != nil {
		log.Println("[endpoints][file][FileDownloadThumb] error while detecting mime type:", err.Error())
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	resp.Header().Set("Content-Type", contentType)

	return c.File(fp)
}

func FileSignEstimateGas(c echo.Context) error {
	fileHash := c.Param("fileHash")

	gasEstimate, err := App.SignFileEstimateGas(fileHash)
	if err != nil {
		log.Print("couldn't estimate gas for sign file, err: ", err)
		return c.JSON(http.StatusBadRequest, errors.New("couldn't estimate gas for sign file"))
	}

	return c.JSON(http.StatusOK, gasEstimate)
}

func FileSign(c echo.Context) error {
	fileHash := c.Param("fileHash")
	txHash, err := App.SignFile(fileHash)
	if err != nil {
		log.Println(err.Error())
		return c.JSON(http.StatusNotFound, err.Error())
	}
	return c.JSON(http.StatusOK, txHash)
}

func FileRemoveFromDiskKeepMeta(c echo.Context) error {
	fileHash := c.Param("fileHash")
	App.RemoveFileFromDiskKeepMeta(fileHash)
	return c.JSON(http.StatusOK, fileHash)
}

func FileRemoveEstimateGas(c echo.Context) error {
	fileHash := c.Param("fileHash")

	gasEstimate, err := App.RemoveFileEstimateGas(fileHash)
	if err != nil {
		log.Print("couldn't estimate gas for remove file ", err)
		return c.JSON(http.StatusBadRequest, errors.New("couldn't estimate gas for remove file"))
	}

	return c.JSON(http.StatusOK, gasEstimate)
}

func FileRemove(c echo.Context) error {
	fileHash := c.Param("fileHash")
	txHash, err := App.RemoveFile(fileHash)
	if err != nil {
		log.Println(err.Error())
		return c.JSON(http.StatusNotFound, err.Error())
	}
	return c.JSON(http.StatusOK, txHash)
}

func FileRemoveLocal(c echo.Context) error {
	fileHash := c.Param("fileHash")

	err := App.RemoveFileLocal(fileHash)
	if err != nil {
		log.Println(err.Error())
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func FileShareEstimateGas(c echo.Context) error {
	fileHash := c.Param("fileHash")

	var ETHAddrs []string
	if err := c.Bind(&ETHAddrs); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	gasEstimate, err := App.ShareFileEstimateGas(fileHash, ETHAddrs)
	if err != nil {
		log.Print("couldn't estimate gas for share file ", err)
		return c.JSON(http.StatusBadRequest, errors.New("couldn't estimate gas for share file"))
	}

	return c.JSON(http.StatusOK, gasEstimate)
}

func FileShare(c echo.Context) error {
	fileHash := c.Param("fileHash")
	var ETHAddrs []string
	if err := c.Bind(&ETHAddrs); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	txHash, err := App.ShareFile(fileHash, ETHAddrs)
	if err != nil {
		log.Println(err.Error())
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, txHash)
}

func FileSigningRequestEstimateGas(c echo.Context) error {
	fileHash := c.Param("fileHash")

	var ETHAddrs []string
	if err := c.Bind(&ETHAddrs); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	gasEstimate, err := App.SendSigningRequestFileEstimateGas(fileHash, ETHAddrs)
	if err != nil {
		log.Print("couldn't estimate gas for send signing request ", err)
		return c.JSON(http.StatusBadRequest, errors.New("couldn't estimate gas for send signing request"))
	}

	return c.JSON(http.StatusOK, gasEstimate)
}

func FileSigningRequest(c echo.Context) error {
	fileHash := c.Param("fileHash")
	var ETHAddrs []string
	if err := c.Bind(&ETHAddrs); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	txHash, err := App.SendSigningRequestFile(fileHash, ETHAddrs)
	if err != nil {
		log.Println(err.Error())
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, txHash)
}

func FileRevokeHashEstimateGas(c echo.Context) error {
	fileHash := c.Param("fileHash")

	var ETHAddrs []string
	if err := c.Bind(&ETHAddrs); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	gasEstimate, err := App.RevokeFileEstimateGas(fileHash, ETHAddrs)
	if err != nil {
		log.Print("couldn't estimate gas for unshare file ", err)
		return c.JSON(http.StatusBadRequest, errors.New("couldn't estimate gas for unshare file"))
	}

	return c.JSON(http.StatusOK, gasEstimate)
}

func FileRevokeHash(c echo.Context) error {
	fileHash := c.Param("fileHash")
	var ETHAddrs []string
	if err := c.Bind(&ETHAddrs); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	txHash, err := App.RevokeFile(fileHash, ETHAddrs)
	if err != nil {
		log.Println(err.Error())
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, txHash)
}

func NewFileEstimateGas(c echo.Context) error {
	fileUploadRequest, err := ParseFileUpload(c)
	if err != nil {
		log.Print("new file estimate gas parse error ", err)
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	gasEstimate, err := App.RegisterFileEstimateGas(fileUploadRequest.Register, fileUploadRequest.DefinedSignerList,
		fileUploadRequest.UndefinedSignersCount, fileUploadRequest.ProviderInfo)
	if err != nil {
		log.Print("couldn't estimate gas for new file ", err)
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, gasEstimate)
}

func NewFile(c echo.Context) error {
	fileUploadRequest, err := ParseFileUpload(c)
	if err != nil {
		log.Print("new file error ", err)
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	if fileUploadRequest.ProviderInfo == (models.StorageProviderInfo{}) {
		return c.JSON(http.StatusBadRequest, ErrNoSppSelected.Error())
	}

	fileHash, err := App.ArchiveFileAndRegister(fileUploadRequest.Register, fileUploadRequest.DefinedSignerList,
		fileUploadRequest.UndefinedSignersCount, fileUploadRequest.ProviderInfo, nil)
	if err != nil {
		if os.IsPermission(err) {
			return c.JSON(http.StatusUnauthorized, err.Error())
		}
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, fileHash)
}

// Returns an estimation of cost
func FileQuote(c echo.Context) error {
	fileUploadRequest, err := ParseFileUpload(c)
	if err != nil {
		log.Print("file quote parse error ", err)
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	encryptedArchiveInfo, err := App.ArchiveFile(fileUploadRequest.Register, fileUploadRequest.DefinedSignerList, fileUploadRequest.UndefinedSignersCount, fileUploadRequest.ProviderInfo)
	if err != nil {
		if s, ok := err.(*os.PathError); ok {
			return c.JSON(http.StatusInternalServerError, errors.New("File "+s.Path+" not found"))
		}
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	log.Printf("dapp file quote: fileHash: %v | fileSizeByte: %v | duration in days: %v",
		encryptedArchiveInfo.FileHash, encryptedArchiveInfo.Size, fileUploadRequest.Register.DurationDays)

	quote, err := App.Quotes(fileUploadRequest.Register.DurationDays, encryptedArchiveInfo.Size)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	quote.FileHash = encryptedArchiveInfo.FileHash

	return c.JSON(http.StatusOK, quote)
}

var ErrNoSppSelected = errors.New("no service provider selected")
var ErrDurationOutOfRange = errors.New("duration out of range")
var ErrDurationNegative = errors.New("duration negative")
var ErrDurationNotSet = errors.New("duration not set")
var ErrNoFileUploaded = errors.New("no file uploaded")

// Parses the request and returns a FileUploadRequest with basic checks.
// If required attributes are missing will return an error. It's up to the caller to check on non-required properties
func ParseFileUpload(c echo.Context) (FileUploadRequest, error) {
	var fileUploadRequest FileUploadRequest
	form, err := c.MultipartForm()
	if err != nil {
		return fileUploadRequest, err
	}
	var definedSignerList []account.AddressBookEntry
	var undefinedSignersCount int64

	var providerInfo models.StorageProviderInfo
	definedSigners := form.Value["definedSigners"]
	if definedSigners != nil && len(definedSigners) > 0 {
		err = json.Unmarshal([]byte(definedSigners[0]), &definedSignerList)
		if err != nil {
			return fileUploadRequest, err
		}
	}
	fileUploadRequest.DefinedSignerList = definedSignerList

	undefinedSigners := form.Value["undefinedSigners"]
	if undefinedSigners != nil && len(undefinedSigners) > 0 {
		err = json.Unmarshal([]byte(undefinedSigners[0]), &undefinedSignersCount)
		if err != nil {
			return fileUploadRequest, err
		}
	}
	fileUploadRequest.UndefinedSignersCount = undefinedSignersCount

	providerAddressForm := form.Value["providerAddress"]
	if len(providerAddressForm) > 0 {
		providerAddress := providerAddressForm[0]
		providerInfo, err = App.GetStorageProvider(providerAddress)
		if err != nil {
			return fileUploadRequest, err
		}
		fileUploadRequest.ProviderInfo = providerInfo
	}

	log.Printf("[endpoints.file][ParseFileUpload] spp name: %s | address: %s | file defined signers count: %d", providerInfo.Name, providerInfo.Address, len(definedSignerList))

	reg := file.Register{}
	files := form.File["file"]

	if len(files) == 0 {
		return fileUploadRequest, ErrNoFileUploaded
	}

	for _, f := range files {
		src, err := f.Open()
		if err != nil {
			return fileUploadRequest, err
		}
		reg.FileName = f.Filename
		reg.FileReader = src
		reg.FileSize = f.Size
		defer src.Close()
		break
	}
	thumbnail := form.File["thumbnail"]
	for _, t := range thumbnail {
		src, err := t.Open()
		if err != nil {
			return fileUploadRequest, err
		}
		reg.ThumbName = t.Filename
		reg.ThumbReader = src
		defer src.Close()
		break
	}
	durationDays := form.Value["duration"]
	if len(durationDays) == 0 {
		return fileUploadRequest, ErrDurationNotSet
	}

	reg.DurationDays, err = strconv.Atoi(durationDays[0])
	if err != nil {
		log.Println("ParseFileUpload error: ", err.Error())
		return fileUploadRequest, ErrDurationOutOfRange
	}

	if reg.DurationDays <= 0 {
		return fileUploadRequest, ErrDurationNegative
	}

	fileUploadRequest.Register = reg
	return fileUploadRequest, nil
}
