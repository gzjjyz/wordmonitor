package wordmonitor

import (
	"github.com/go-resty/resty/v2"
)

var _sharedRestyClient *resty.Client

func init() {
	_sharedRestyClient = resty.New()
}

func GetRestyClient() *resty.Client {
	return _sharedRestyClient
}
