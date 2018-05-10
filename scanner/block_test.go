package scanner

import (
	"testing"

	"github.com/go-xorm/xorm"
)

func TestFoo(t *testing.T) {

	e, err := xorm.NewEngine("sqlite3", "/tmp/block.db")
	if err != nil {
		panic(err)
	}

	if err := e.Ping(); err != nil {
		panic(err)
	}
}
