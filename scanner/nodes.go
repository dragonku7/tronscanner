package scanner

import (
	"github.com/go-xorm/xorm"
)

type Nodes struct {
}

func (n *Nodes) DoWork(engine *xorm.Engine) error {
	return nil

}
