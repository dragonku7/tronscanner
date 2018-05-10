package scanner

import (
	"github.com/go-xorm/xorm"
)

type Witness struct {
}

func (t *Witness) DoWork(engine *xorm.Engine) error {
	return nil
}
