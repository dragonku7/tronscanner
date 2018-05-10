package scanner

import (
	"github.com/go-xorm/xorm"
)

type Accounts struct {
}

func (acc *Accounts) DoWork(engine *xorm.Engine) error {
	return nil
}
