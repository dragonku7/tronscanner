package scanner

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-xorm/xorm"
	"github.com/taczc64/tronscanner/types"
	"github.com/tronprotocol/go-client-api/api"
	protocol1 "github.com/tronprotocol/go-client-api/core"
)

type Block struct {
	cli    api.WalletClient
	ticker *time.Ticker
	e      *xorm.Engine
}

func NewBlockWorker(client api.WalletClient) *Block {
	return &Block{cli: client, ticker: time.NewTicker(time.Second * 5)}
}

func (b *Block) DoWork(engine *xorm.Engine) error {
	b.e = engine
	b.Init()

	for {
		maxHeightLocal := b.GetMaxHeightLocal()
		maxHeightRemote := b.GetMaxHeightRemote()
		if maxHeightRemote > maxHeightLocal {
			fmt.Printf("start pulling from %d to %d\n", maxHeightLocal, maxHeightRemote)
		}
		b.Pull(maxHeightLocal, maxHeightRemote)

		<-b.ticker.C
	}

	return nil
}

func (b *Block) Init() {
	if err := b.e.Ping(); err != nil {
		panic(err)
	}

	exist, err := b.e.IsTableExist(types.TableNameBlock)
	if err != nil {
		panic(err)
	}

	if !exist {
		fmt.Println("table not exist")
		if _, err := b.e.Exec(`create table block
			(
				number		INT	PRIMARY KEY,
				witnessid	INT,
				timestamp	INT,
				parenthash	VARCHAR(64),
				trieroot	VARCHAR(64),
				witnessaddr	VARCHAR(64),
				witnesssig	VARCHAR(128)
			)`); err != nil {
			panic(err)
		}

		if _, err := b.e.Exec(`create table tx
			(
				id				INTEGER	PRIMARY KEY AUTOINCREMENT,
				refblocknumber	INT,
				expiration		INT,
				timestamp		INT,
				refblockhash	VARCHAR(64),
				scrpits			TEXT,
				date			TEXT,
				sigs			TEXT
			)`); err != nil {
			panic(err)
		}
	}
}

func (b *Block) GetMaxHeightLocal() int64 {
	type maxidst struct {
		Id int64
	}

	//var id int64
	var maxid maxidst
	sql := "select max(number) as id from block"
	has, err := b.e.SQL(sql).Get(&maxid)
	if err != nil {
		panic(err)
	}

	if !has {
		panic("max number not found")
	}

	return maxid.Id
}

func (b *Block) GetMaxHeightRemote() int64 {
	now, err := b.cli.GetNowBlock(context.Background(), new(api.EmptyMessage))
	if err != nil {
		panic(err)
	}

	return now.GetBlockHeader().GetRawData().GetNumber()
}

func (b *Block) Pull(start, end int64) {
	for i := start + 1; i <= end; i++ {
		bh := &api.NumberMessage{Num: i}
		bl, err := b.cli.GetBlockByNum(context.Background(), bh)
		if err != nil {
			fmt.Println("GetBlockByNum failed:", err)
			continue
		}

		b.SaveBlock(bl)
		c := b.SaveTxs(bl)
		fmt.Printf("block %d is saved with %d txs\n", bl.GetBlockHeader().GetRawData().GetNumber(), c)
	}
}

func bytesToString(bs []byte) string {
	return hex.EncodeToString(bs)
}

func toJSON(o interface{}) string {
	bs, _ := json.Marshal(o)
	return string(bs)
}

func (b *Block) SaveBlock(bl *protocol1.Block) {
	number := bl.GetBlockHeader().GetRawData().GetNumber()
	witnessID := bl.GetBlockHeader().GetRawData().GetWitnessId()
	timestamp := bl.GetBlockHeader().GetRawData().GetTimestamp()
	parentHash := bl.GetBlockHeader().GetRawData().GetParentHash()
	trieRoot := bl.GetBlockHeader().GetRawData().GetTxTrieRoot()
	witnessAddr := bl.GetBlockHeader().GetRawData().GetWitnessAddress()
	witnessSig := bl.GetBlockHeader().GetWitnessSignature()

	res, err := b.e.Exec(`insert into block
		(number,witnessid,timestamp,parenthash,trieroot,witnessaddr,witnesssig)
		values
		(?,?,?,?,?,?,?)`,
		number, witnessID, timestamp, bytesToString(parentHash), bytesToString(trieRoot), bytesToString(witnessAddr), bytesToString(witnessSig))
	if err != nil {
		panic(err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		panic(err)
	}
	if rows != 1 {
		panic("insert error")
	}
}

func parseSigs(sigs [][]byte) string {
	sigStrs := make([]string, len(sigs))
	for i, v := range sigs {
		sigStrs[i] = bytesToString(v)
	}
	return toJSON(sigStrs)
}

func (b *Block) SaveTxs(bl *protocol1.Block) int {
	txs := bl.GetTransactions()
	for _, v := range txs {
		refblocknumber := v.GetRawData().GetRefBlockNum()
		expiration := v.GetRawData().GetExpiration()
		timestamp := v.GetRawData().GetTimestamp()
		refblockhash := v.GetRawData().GetRefBlockHash()
		scrpits := v.GetRawData().GetScripts()
		data := v.GetRawData().GetData()
		sigs := v.GetSignature()

		res, err := b.e.Exec(`insert into tx
		(refblocknumber,expiration,timestamp,refblockhash,scrpits,date,sigs)
		values
		(?,?,?,?,?,?,?)`,
			refblocknumber, expiration, timestamp, bytesToString(refblockhash), scrpits, data, parseSigs(sigs))
		if err != nil {
			panic(err)
		}
		rows, err := res.RowsAffected()
		if err != nil {
			panic(err)
		}
		if rows != 1 {
			panic("insert error")
		}
	}
	return len(txs)
}
