package scanner

import (
	"github.com/go-xorm/xorm"
)

type Block struct {
}

func (b *Block) DoWork(engine *xorm.Engine) error {

	return nil
}
