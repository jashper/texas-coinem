package Server

type ServerContext struct {
	DB       *Database
	Entropy  *EntropyPool
	HandEval *HandEvaluator
}
