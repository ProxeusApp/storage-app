package test

import (
	"fmt"
	"net/http"
)

func testCreateUser(s *session) {
	responseBody := s.e.PUT("/api/account").WithJSON(map[string]string{"name": "test-account", "pw": "ian72Am"}).Expect().Status(http.StatusOK).Body()
	responseBody.Contains("0x")
	ethAddress = trimResponseString(responseBody.Raw())

	fmt.Println("testCreateUser created user with address: ", ethAddress)
}

func testLogout(s *session) {

	waitForHttpStatusCode(s.base+"/api/account/balance", 204, 2)

	fmt.Println("testLogout will logout user with address: ", ethAddress)
	s.e.POST("/api/logout").Expect().Status(http.StatusOK)
	fmt.Println("testLogout successful logout user with address: ", ethAddress)
}

func testLogin(s *session) {
	waitForHttpStatusCode(s.base+"/api/account/balance", 401, 2)
	fmt.Println("testLogin will login with ethAddress: ", ethAddress)
	s.e.POST("/api/login").WithJSON(map[string]string{"ethAddr": ethAddress, "pw": "ian72Am"}).Expect().Status(http.StatusOK)
	fmt.Println("testLogin successful login with ethAddress: ", ethAddress)
}
