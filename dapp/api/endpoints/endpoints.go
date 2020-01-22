package endpoints

import (
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"git.proxeus.com/core/central/dapp/core"
)

var App *core.App

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

func detectMimeType(filePath string) (string, error) {
	contentType := mime.TypeByExtension(filepath.Ext(filePath))
	if contentType == "" {
		file, err := os.Open(filePath)
		if err != nil {
			return "", err
		}
		defer file.Close()

		buffer := make([]byte, 512)
		byteCount, err := file.Read(buffer)
		contentType = http.DetectContentType(buffer[:byteCount])
		// Workaround because DetectContentType doesn't detect image/svg+xml
		if strings.Contains(contentType, "text/xml") && strings.Contains(string(buffer[:byteCount]), "<svg") {
			contentType = strings.Replace(contentType, "text/xml", "image/svg+xml", 1)
		}
	}

	return contentType, nil
}
