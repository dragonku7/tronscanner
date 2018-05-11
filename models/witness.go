package models

type Witness struct {
	ID             int64  `xorm:"'Id' int pk autoincr"`
	Address        string `xorm:"'address' varchar(64) notnull unique"`
	VoteCount      int64  `xorm:"'voteCount' int notnull"`
	PubKey         string `xorm:"'pubKey' varchar(128)"`
	URL            string `xorm:"'url' text"`
	TotalProduced  int64  `xorm:"'totalProduced' int"`
	TotalMissed    int64  `xorm:"'totalMissed' int"`
	LatestBlockNum int64  `xorm:"'latestBlockNum' int"`
	LatestSlotNum  int64  `xorm:"'latestSlotNum' int"`
	IsJobs         bool   `xorm:"'isJobs' boolean"`
}

func (w *Witness) TableName() string {
	return "tron_witness"
}

type Nodes struct {
	ID   int64  `xorm:"'Id' int pk autoincr"`
	Host string `xorm:"'host' text notnull"`
	Port int32  `xorm:"'port' int notnull"`
}

func (n *Nodes) TableName() string {
	return "tron_nodes"
}
