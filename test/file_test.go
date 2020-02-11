package test

import (
	"fmt"
	"net/http"
	"os"
)

var fileHash string
var fileContent string

func testFileUpload(s *session) {
	fileUnderTest, err := os.Create("./data/test-file")
	if err != nil {
		s.t.Fatal(err)
	}
	fileContent = "file-" + ethAddress
	_, err = fileUnderTest.Write([]byte(fileContent))
	if err != nil {
		s.t.Fatal(err)
	}
	err = fileUnderTest.Close()
	if err != nil {
		s.t.Fatal(err)
	}

	fh, err := os.Open("./data/test-file")
	if err != nil {
		s.t.Fatal(err)
	}

	fhThumb, err := os.Open("./data/test-thumbnail.png")
	if err != nil {
		s.t.Fatal(err)
	}

	responseBody := s.e.POST("/api/file/new").WithMultipart().WithFormField("providerAddress", "0x5C9eDfaaC887552D6b521E38dAA3BFf1f645fD36").
		WithFormField("duration", 1).WithFile("file", "test-file.png", fh).WithFile("thumbnail", "test-thumbnail.png", fhThumb).
		Expect().Status(http.StatusOK).Body()

	fileHash = trimResponseString(responseBody.Raw())

	err = fh.Close()
	if err != nil {
		s.t.Fatal(err)
	}
}

func testGetDownload(s *session) {
	url := fmt.Sprintf("/api/file/download/%s", fileHash)
	waitForHttpStatusCode(fmt.Sprintf("%s%s", s.base, url), 200, 1)
	response := s.e.GET(url).Expect().Status(200).Body().Raw()

	if fileContent != response {
		s.t.Fatalf("Expected downloaded filename to be '%s' but got '%s'", fileContent, response)
	}

	fmt.Println("testLogout will logout user with address: ", ethAddress)
	s.e.POST("/api/logout").Expect().Status(http.StatusOK)
	fmt.Println("testLogout successful logout user with address: ", ethAddress)

	s.e.GET(url).Expect().Status(404)
}
