package proto

type Repl struct {
	Rbid       float64  `bson:"rbid"`
	SetVersion float64  `bson:"setVersion"`
	ElectionId string   `bson:"electionId"`
	Primary    string   `bson:"primary"`
	Me         string   `bson:"me"`
	Secondary  bool     `bson:"secondary"`
	SetName    string   `bson:"setName"`
	Hosts      []string `bson:"hosts"`
	Ismaster   bool     `bson:"ismaster"`
}
