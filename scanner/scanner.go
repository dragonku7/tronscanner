package scanner

import (
	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tronprotocol/go-client-api/api"
	"google.golang.org/grpc"
	"sync"
)

var Rlock *sync.RWMutex

const (
	address = "47.91.216.69:50051"
)

type Scanner struct {
	Engines     map[string]*xorm.Engine
	tronClient  api.WalletClient
	conn        *grpc.ClientConn
	nodeAddress string
}

func NewScanner() (*Scanner, error) {
	var err error
	Rlock = new(sync.RWMutex)

	var scan Scanner
	scan.Engines = make(map[string]*xorm.Engine)
	scan.Engines["block"], err = xorm.NewEngine("sqlite3", "./data/block.db")
	scan.Engines["accounts"], err = xorm.NewEngine("sqlite3", "./data/accounts.db")
	scan.Engines["witness"], err = xorm.NewEngine("sqlite3", "./data/witness.db")
	scan.Engines["nodes"] = scan.Engines["witness"]
	if err != nil {
		panic(err)
	}
	// if err := scan.Engines.Ping(); err != nil {
	// panic(err)
	// }
	scan.nodeAddress = address
	return &scan, err
}

func (s *Scanner) Start() error {
	var err error
	s.conn, err = grpc.Dial(s.nodeAddress, grpc.WithInsecure())
	if err != nil {
		return err
	}
	s.tronClient = api.NewWalletClient(s.conn)
	//load db data
	s.loadLastInfo()

	//routines to get data
	for dbname, engine := range s.Engines {
		var worker BaseWork
		switch dbname {
		case "block":
			worker = NewBlockWorker(s.tronClient)
		case "accounts":
			worker = NewAccountsWorker(s.tronClient)
		case "witness":
			worker = NewWitnessWorker(s.tronClient)
		case "nodes":
			worker = NewNodesWorker(s.tronClient)
		}
		go worker.DoWork(engine)
	}
	return nil
}

func (s *Scanner) loadLastInfo() {

}

type BaseWork interface {
	DoWork(*xorm.Engine) error
}

func (s *Scanner) Stop() {
	s.conn.Close()
}
