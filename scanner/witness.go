package scanner

import (
	"github.com/go-xorm/xorm"
	"github.com/tronprotocol/go-client-api/api"
	"time"
)

type Witness struct {
	cli    api.WalletClient
	ticker *time.Ticker
}

func NewWitnessWorker(client api.WalletClient) *Witness {
	return &Witness{cli: client, ticker: time.NewTicker(time.Second * 1)}
}

func (t *Witness) DoWork(engine *xorm.Engine) error {
	return nil
}
