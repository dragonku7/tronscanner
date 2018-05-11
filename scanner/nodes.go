package scanner

import (
	"time"

	"context"
	"fmt"
	"github.com/go-xorm/xorm"
	"github.com/taczc64/tronscanner/models"
	"github.com/tronprotocol/go-client-api/api"
)

type Nodes struct {
	cli    api.WalletClient
	ticker *time.Ticker
	eng    *xorm.Engine
}

func NewNodesWorker(client api.WalletClient) *Nodes {
	return &Nodes{cli: client, ticker: time.NewTicker(time.Second * 5)}
}

func (n *Nodes) DoWork(engine *xorm.Engine) error {
	n.eng = engine
	err := n.eng.Sync2(new(models.Nodes))
	if err != nil {
		fmt.Errorf(err.Error())
		return err
	}
	nodesChan := make(chan *api.NodeList, 32)
	go n.requestChannel(nodesChan)
	//write to db
	for nodesList := range nodesChan {
		Rlock.Lock()
		//delete db old data
		_, err := n.eng.Exec("delete from tron_nodes;")
		if err != nil {
			fmt.Errorf(err.Error())
			continue
		}
		for _, node := range nodesList.Nodes {
			//insert
			_, err := n.eng.InsertOne(&models.Nodes{
				Host: string(node.GetAddress().GetHost()),
				Port: node.GetAddress().GetPort(),
			})
			if err != nil {
				fmt.Errorf(err.Error())
				continue
			}
		}
		Rlock.Unlock()
	}
	return nil
}

func (n *Nodes) requestChannel(nodesChan chan *api.NodeList) {
	list, err := n.cli.ListNodes(context.Background(), new(api.EmptyMessage))
	if err != nil {
		fmt.Errorf(err.Error())
		return
	}
	nodesChan <- list
	for {
		select {
		case <-n.ticker.C:
			nodeslist, err := n.cli.ListNodes(context.Background(), new(api.EmptyMessage))
			if err != nil {
				continue
			}
			nodesChan <- nodeslist
		}
	}
}
