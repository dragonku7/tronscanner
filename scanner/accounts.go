package scanner

import (
	"github.com/go-xorm/xorm"
	"github.com/tronprotocol/go-client-api/api"
	"time"
)

type Accounts struct {
	cli                 api.WalletClient
	ticker              *time.Ticker
	AccountName         string
	Type                int32
	Address             string
	Balance             int64
	LatestOperationTime int64
}

func NewAccountsWorker(client api.WalletClient) *Accounts {
	return &Accounts{cli: client, ticker: time.NewTicker(time.Second * 1)}
}

func (acc *Accounts) DoWork(engine *xorm.Engine) error {
	return nil
}
