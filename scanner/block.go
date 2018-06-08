package scanner

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-xorm/xorm"
	"github.com/gogo/protobuf/proto"
	google_protobuf "github.com/golang/protobuf/ptypes/any"
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
		fmt.Println(maxHeightLocal, maxHeightRemote)
		if maxHeightLocal < 0 || maxHeightRemote < 0 {
			time.Sleep(time.Second * 5)
			continue
		}
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
				blocknumber		INT,
				refblocknumber	INT,
				owneraddr		VARCHAR(64),
				toaddr			VARCHAR(64),
				expiration		INT,
				timestamp		INT,
				refblockhash	VARCHAR(64),
				scrpits			TEXT,
				data			TEXT,
				contracts		TEXT,
				sigs			TEXT
			)`); err != nil {
			panic(err)
		}

		if _, err := b.e.Exec(`CREATE INDEX IF NOT EXISTS txheight ON tx (blocknumber);`); err != nil {
			panic(err)
		}
		if _, err := b.e.Exec(`CREATE INDEX IF NOT EXISTS txfrom ON tx (owneraddr);`); err != nil {
			panic(err)
		}
		if _, err := b.e.Exec(`CREATE INDEX IF NOT EXISTS txto ON tx (toaddr);`); err != nil {
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
		fmt.Println(err)
		return -1
	}

	if !has {
		fmt.Println("max number not found")
		return -1
	}

	return maxid.Id
}

func (b *Block) GetMaxHeightRemote() int64 {
	now, err := b.cli.GetNowBlock(context.Background(), new(api.EmptyMessage))
	if err != nil {
		fmt.Println(err)
		return -1
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
		fmt.Printf("saved block %d with %d txs\n", i, c)
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

func parseContracts(contracts []*protocol1.Transaction_Contract) (cc string, from, to *[]byte) {
	type contract struct {
		Type     string
		Name     string
		Provider string
		Content  string
	}
	vals := make([]*contract, len(contracts))
	for i, v := range contracts {
		var txc string
		txc, from, to = parseContractContent(v.GetType(), v.GetParameter())
		vals[i] = &contract{
			Type:     v.GetType().String(),
			Name:     string(v.GetContractName()),
			Provider: bytesToString(v.GetProvider()),
			Content:  txc,
		}
	}
	return toJSON(vals), from, to
}

func parseContractContent(t protocol1.Transaction_Contract_ContractType, p *google_protobuf.Any) (txc string, from, to *[]byte) {
	var tx proto.Message
	switch t {
	case protocol1.Transaction_Contract_AccountCreateContract:
		raw := &protocol1.AccountCreateContract{}
		tx = raw
		from = &raw.OwnerAddress
	case protocol1.Transaction_Contract_TransferContract:
		raw := &protocol1.TransferContract{}
		tx = raw
		from = &raw.OwnerAddress
		to = &raw.ToAddress
	case protocol1.Transaction_Contract_TransferAssetContract:
		raw := &protocol1.TransferAssetContract{}
		tx = raw
		from = &raw.OwnerAddress
		to = &raw.ToAddress
	case protocol1.Transaction_Contract_VoteAssetContract:
		raw := &protocol1.VoteAssetContract{}
		tx = raw
		from = &raw.OwnerAddress
	case protocol1.Transaction_Contract_VoteWitnessContract:
		raw := &protocol1.VoteWitnessContract{}
		tx = raw
		from = &raw.OwnerAddress
	case protocol1.Transaction_Contract_WitnessCreateContract:
		raw := &protocol1.WitnessCreateContract{}
		tx = raw
		from = &raw.OwnerAddress
	case protocol1.Transaction_Contract_AssetIssueContract:
		raw := &protocol1.AssetIssueContract{}
		tx = raw
		from = &raw.OwnerAddress
	case protocol1.Transaction_Contract_DeployContract:
		raw := &protocol1.DeployContract{}
		tx = raw
		from = &raw.OwnerAddress
	case protocol1.Transaction_Contract_WitnessUpdateContract:
		raw := &protocol1.WitnessUpdateContract{}
		tx = raw
		from = &raw.OwnerAddress
	case protocol1.Transaction_Contract_ParticipateAssetIssueContract:
		raw := &protocol1.ParticipateAssetIssueContract{}
		tx = raw
		from = &raw.OwnerAddress
		to = &raw.ToAddress
	case protocol1.Transaction_Contract_AccountUpdateContract:
		raw := &protocol1.AccountUpdateContract{}
		tx = raw
		from = &raw.OwnerAddress
	case protocol1.Transaction_Contract_FreezeBalanceContract:
		raw := &protocol1.FreezeBalanceContract{}
		tx = raw
		from = &raw.OwnerAddress
	case protocol1.Transaction_Contract_UnfreezeBalanceContract:
		raw := &protocol1.UnfreezeBalanceContract{}
		tx = raw
		from = &raw.OwnerAddress
	case protocol1.Transaction_Contract_WithdrawBalanceContract:
		raw := &protocol1.WithdrawBalanceContract{}
		tx = raw
		from = &raw.OwnerAddress
	case protocol1.Transaction_Contract_CustomContract:
	default:
		return "", nil, nil
	}

	if err := proto.Unmarshal(p.GetValue(), tx); err != nil {
		panic(err)
	}

	txc = toJSON(tx)

	return
}

func (b *Block) SaveTxs(bl *protocol1.Block) int {
	txs := bl.GetTransactions()
	height := bl.GetBlockHeader().GetRawData().GetNumber()
	for _, v := range txs {
		refblocknumber := v.GetRawData().GetRefBlockNum()
		expiration := v.GetRawData().GetExpiration()
		timestamp := v.GetRawData().GetTimestamp()
		refblockhash := v.GetRawData().GetRefBlockHash()
		scrpits := v.GetRawData().GetScripts()
		data := v.GetRawData().GetData()
		contracts := v.GetRawData().GetContract()
		sigs := v.GetSignature()
		cc, from, to := parseContracts(contracts)
		var fromaddr, toaddr interface{}
		if from != nil {
			fromaddr = bytesToString(*from)
		} else {
			fromaddr = nil
		}
		if to != nil {
			toaddr = bytesToString(*to)
		} else {
			toaddr = nil
		}

		res, err := b.e.Exec(`insert into tx
		(blocknumber,refblocknumber,owneraddr,toaddr,expiration,timestamp,refblockhash,scrpits,data,contracts,sigs)
		values
		(?,?,?,?,?,?,?,?,?,?,?)`,
			height, refblocknumber, fromaddr, toaddr, expiration, timestamp, bytesToString(refblockhash), scrpits, data, cc, parseSigs(sigs))
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
