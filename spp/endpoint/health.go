package endpoint

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"
)

/**
Health is used to check dependencies' statuses
*/

const (
	up   string = "UP"
	down string = "DOWN"
)

type serviceStatus struct {
	ServiceName string
	Status      string
	Info        string
}

// Checks dependencies' health and returns ServiceStatuses
func Health(c echo.Context) error {
	serviceStatuses := []serviceStatus{
		ethServiceStatus(),
	}
	return c.JSON(http.StatusOK, serviceStatuses)
}

// Tests Ethereum connection from this service and returns a serviceStatus
func ethServiceStatus() serviceStatus {
	ethStatus := serviceStatus{
		ServiceName: "ethereum",
		Status:      down,
	}
	header, err := EthClient.HeaderByNumber(nil)
	if err != nil || header == nil {
		ethStatus.Info = "Not connected"
	} else {
		ethStatus.Status = up
		ethStatus.Info = fmt.Sprintf("Connected. Last block: %d", header.Number)
	}
	return ethStatus
}
