package scanner

import (
	"context"
	"fmt"
	"time"

	"github.com/go-xorm/xorm"
	"github.com/taczc64/tronscanner/models"
	"github.com/tronprotocol/go-client-api/api"
)

type Accounts struct {
	cli    api.WalletClient
	ticker *time.Ticker
	eng    *xorm.Engine
}

func NewAccountsWorker(client api.WalletClient) *Accounts {
	return &Accounts{cli: client, ticker: time.NewTicker(time.Second * 3)}
}

func (acc *Accounts) DoWork(engine *xorm.Engine) error {
	acc.eng = engine
	err := acc.eng.Sync2(new(models.Account))
	if err != nil {
		fmt.Println("error :", err)
		return err
	}
	//get Accounts
	accountChan := make(chan *api.AccountList, 1024)
	go func() {
		for {
			select {
			case <-acc.ticker.C:
				accounts, err := acc.cli.ListAccounts(context.Background(), new(api.EmptyMessage))
				if err != nil {
					continue
				}
				accountChan <- accounts
			}
		}
	}()

	//write to db
	// for account := range accountChan {
	// }
	return nil
}
