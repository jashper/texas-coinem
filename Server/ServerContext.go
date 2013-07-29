package Server

type ServerContext struct {
	DB       *Database
	Entropy  *EntropyPool
	HandEval *HandEvaluator
}

func (this *ServerContext) Init(db *Database, entropy *EntropyPool,
	handEval *HandEvaluator) {

	this.DB = db
	this.Entropy = entropy
	this.HandEval = handEval

}
