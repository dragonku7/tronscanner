package scanner

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/go-xorm/xorm"
	"github.com/taczc64/tronscanner/models"
	"github.com/tronprotocol/go-client-api/api"
	"github.com/tronprotocol/go-client-api/core"
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
	accountChan := make(chan *api.AccountList, 32)
	go acc.requestChannel(accountChan)
	//write to db
	for accountList := range accountChan {
		//account is exist ?
		for _, account := range accountList.Accounts {
			// fmt.Println("ready to insert account :", hex.EncodeToString(account.GetAddress()))
			sqldata := apiAccountToSqlAcc(account)

			if has, _ := acc.eng.Exist(&models.Account{Address: sqldata.Address}); has { //update
				_, err := acc.eng.Update(sqldata)
				if err != nil {
					fmt.Errorf(err.Error())
					continue
				}
			} else { //insert
				_, err := acc.eng.InsertOne(sqldata)
				if err != nil {
					fmt.Errorf(err.Error())
					continue
				}
			}
		}

	}
	return nil
}

func apiAccountToSqlAcc(acc *core.Account) *models.Account {
	sqlacc := new(models.Account)
	sqlacc.AccountName = hex.EncodeToString(acc.GetAccountName())
	sqlacc.Address = hex.EncodeToString(acc.GetAddress())
	sqlacc.Asset = maptostring(acc.GetAsset())
	sqlacc.Balance = acc.GetBalance()
	sqlacc.Type = int32(acc.GetType())
	sqlacc.LatestOperationTime = acc.GetLatestOprationTime()
	votes := acc.GetVotes()
	vstr := ""
	for _, vote := range votes {
		vstr = vstr + vote.String()
	}
	return sqlacc
}

func maptostring(m map[string]int64) string {
	bs := new(bytes.Buffer)
	for key, value := range m {
		fmt.Fprintf(bs, "%s=\"%d\";", key, value)
	}
	return bs.String()
}

func (acc *Accounts) requestChannel(accChan chan *api.AccountList) {
	for {
		select {
		case <-acc.ticker.C:
			accounts, err := acc.cli.ListAccounts(context.Background(), new(api.EmptyMessage))
			if err != nil {
				continue
			}
			accChan <- accounts
		}
	}
}
