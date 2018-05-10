package models

type Account struct {
	ID                  int64  `xorm:"'Id' int pk autoincr"`
	AccountName         string `xorm:"'account_name' text notnull"`
	Type                int32  `xorm:"'type' int notnull"`
	Address             string `xorm:"'address' varchar(64) notnull unique"`
	Balance             int64  `xorm:"'balance' int notnull"`
	Votes               string `xorm:"'votes' text notnull"`
	Asset               string `xorm:"'asset' text notnull"`
	LatestOperationTime int64  `xorm:"'latest_operation_time' int notnull"`
}

func (a *Account) TableName() string {
	return "tron_accounts"
}
