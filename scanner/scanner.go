package scanner

import (
	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tronprotocol/go-client-api/api"
	"google.golang.org/grpc"
)

const (
	address = "47.91.216.69:50051"
)

type Scanner struct {
	Engine      *xorm.Engine
	tronClient  api.WalletClient
	conn        *grpc.ClientConn
	nodeAddress string
}

func NewScanner() (*Scanner, error) {
	var err error

	var scan Scanner
	scan.Engine, err = xorm.NewEngine("sqlite3", "./data/scanner.db")
	if err != nil {
		return nil, err
	}
	if err := scan.Engine.Ping(); err != nil {
		panic(err)
	}
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

	//routine for getting new block

	//routine for getting new transaction
	return nil
}

func (s *Scanner) loadLastInfo() {

}

//NewblockWriter request for Newblock info & write to db
func (s *Scanner) NewblockWriter() {

}

//NewTransactionWriter request for transaction info & write to db
func (s *Scanner) NewTransactionWriter() {

}

func (s *Scanner) AccountsUpdater() {

}

func (s *Scanner) WitnessUpdater() {

}

func (s *Scanner) NodesUpdater() {

}

func (s *Scanner) Stop() {
	s.conn.Close()
}
