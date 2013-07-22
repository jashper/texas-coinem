package Server

type ServerContext struct {
	DB          *Database
	Entropy     *EntropyPool
	HandEval    *HandEvaluator
	Connections []*Connection
	CurrentGame *GameInstance
}

func (this *ServerContext) Init(db *Database, entropy *EntropyPool,
	handEval *HandEvaluator) {

	this.DB = db
	this.Entropy = entropy
	this.HandEval = handEval
	this.Connections = make([]*Connection, 0)

}
