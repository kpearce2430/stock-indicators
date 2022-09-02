package http_client

import (
	"github.com/imroc/req/v3"
	"time"
)

func GetDefaultClient(timeout int64, devMode bool) *req.Client {

	client := req.C().SetTimeout(time.Duration(timeout) * time.Second)

	if devMode {
		client.DevMode()
	}

	return client
}
