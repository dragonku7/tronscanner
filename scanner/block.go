package scanner

import (
	"github.com/go-xorm/xorm"
	"github.com/tronprotocol/go-client-api/api"
	"time"
)

type Block struct {
	cli    api.WalletClient
	ticker *time.Ticker
}

func NewBlockWorker(client api.WalletClient) *Block {
	return &Block{cli: client, ticker: time.NewTicker(time.Second * 1)}
}

func (b *Block) DoWork(engine *xorm.Engine) error {
	return nil
}
