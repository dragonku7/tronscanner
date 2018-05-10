package scanner

import (
	"github.com/go-xorm/xorm"
)

type Transaction struct {
}

func (t *Transaction) DoWork(engine *xorm.Engine) error {

	return nil
}
