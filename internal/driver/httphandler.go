package driver

import (
	"net/http"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
)

type HttpHandler struct {
	driver *Driver
}

func NewHttpHandler(driver *Driver) HttpHandler {
	return HttpHandler{driver}
}

func (h HttpHandler) RefreshExistingDevicePaths(writer http.ResponseWriter, request *http.Request) {
	go h.driver.RefreshExistingDevicePaths()
	correlationID := request.Header.Get(common.CorrelationHeader)
	writer.Header().Set(common.CorrelationHeader, correlationID)
	writer.Header().Set(common.ContentType, common.ContentTypeJSON)
	writer.WriteHeader(http.StatusAccepted)
}
