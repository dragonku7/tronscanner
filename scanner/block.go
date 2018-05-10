package scanner

import (
	"time"

	"github.com/go-xorm/xorm"
	"github.com/tronprotocol/go-client-api/api"
)

type Block struct {
	cli    api.WalletClient
	ticker *time.Ticker
	e      *xorm.Engine
}

func NewBlockWorker(client api.WalletClient) *Block {
	return &Block{cli: client, ticker: time.NewTicker(time.Second * 1)}
}

func (b *Block) DoWork(engine *xorm.Engine) error {
	b.e = engine
	b.Init()

	return nil
}

func (b *Block) Init() {
	if err := b.e.Ping(); err != nil {
		panic(err)
	}

	// exist, err := b.e.IsTableExist(types.TableNameBlock)
	// if err != nil {
	// 	panic(err)
	// }

	// if !exist {
	// 	b.e.Exec(`create table block(
	// 		id		INT	PRIMARY KEY,

	// 		`)
	// }
}
