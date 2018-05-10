package scanner

import (
	"time"

	"github.com/go-xorm/xorm"
	"github.com/tronprotocol/go-client-api/api"
)

type Nodes struct {
	cli    api.WalletClient
	ticker *time.Ticker
}

func NewNodesWorker(client api.WalletClient) *Nodes {
	return &Nodes{cli: client, ticker: time.NewTicker(time.Second * 1)}
}

func (n *Nodes) DoWork(engine *xorm.Engine) error {
	return nil

}
