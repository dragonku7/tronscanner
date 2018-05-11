package scanner

import (
	"context"
	"encoding/hex"
	"github.com/tronprotocol/go-client-api/core"
	"time"

	"fmt"
	"github.com/go-xorm/xorm"
	"github.com/taczc64/tronscanner/models"
	"github.com/tronprotocol/go-client-api/api"
)

type Witness struct {
	cli    api.WalletClient
	ticker *time.Ticker
	eng    *xorm.Engine
}

func NewWitnessWorker(client api.WalletClient) *Witness {
	return &Witness{cli: client, ticker: time.NewTicker(time.Second * 20)}
}

func (t *Witness) DoWork(engine *xorm.Engine) error {
	t.eng = engine
	err := t.eng.Sync2(new(models.Witness))
	if err != nil {
		fmt.Errorf(err.Error())
		return err
	}
	witnessChan := make(chan *api.WitnessList, 32)
	go t.requestChannel(witnessChan)
	//write to db
	for witnessList := range witnessChan {
		//delete db old data
		_, err := t.eng.Exec("delete from tron_witness;")
		if err != nil {
			fmt.Errorf(err.Error())
			continue
		}
		for _, witness := range witnessList.Witnesses {
			witObj := apiWitnessToSqlAcc(witness)
			//insert
			_, err := t.eng.InsertOne(witObj)
			if err != nil {
				fmt.Errorf(err.Error())
				continue
			}
		}
	}
	return nil
}

func apiWitnessToSqlAcc(wit *core.Witness) *models.Witness {
	witobj := new(models.Witness)
	witobj.Address = hex.EncodeToString(wit.GetAddress())
	witobj.VoteCount = wit.GetVoteCount()
	witobj.PubKey = hex.EncodeToString(wit.GetPubKey())
	witobj.URL = wit.GetUrl()
	witobj.TotalProduced = wit.GetTotalProduced()
	witobj.TotalMissed = wit.GetTotalMissed()
	witobj.LatestBlockNum = wit.GetLatestBlockNum()
	witobj.LatestSlotNum = wit.GetLatestSlotNum()
	witobj.IsJobs = wit.GetIsJobs()
	return witobj
}

func (w *Witness) requestChannel(witnessChan chan *api.WitnessList) {
	list, err := w.cli.ListWitnesses(context.Background(), new(api.EmptyMessage))
	if err != nil {
		fmt.Errorf(err.Error())
		return
	}
	witnessChan <- list
	for {
		select {
		case <-w.ticker.C:
			witnesslist, err := w.cli.ListWitnesses(context.Background(), new(api.EmptyMessage))
			if err != nil {
				continue
			}
			witnessChan <- witnesslist
		}
	}
}
