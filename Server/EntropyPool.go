package Server

import (
	"math/rand"
	"time"
)

type EntropyPool struct {
	Decks chan [52]int
}

func (this *EntropyPool) Init(bufCount int) {
	this.Decks = make(chan [52]int, bufCount)
	rand.Seed(time.Now().UTC().UnixNano())
}

func (this EntropyPool) Run(reseed int) {
	for count := 1; ; count++ {
		var testDeck [52]int

		cards := make([]int, 52)
		for c := 1; c <= 52; c++ {
			cards[c-1] = c
		}
		for i := 0; i < 51; i++ {
			toPop := rand.Intn(52 - i)
			testDeck[i] = cards[toPop]
			cards = cards[:toPop+
				copy(cards[toPop:], cards[toPop+1:])]
		}
		testDeck[51] = cards[0]

		this.Decks <- testDeck

		if count%reseed == 0 {
			rand.Seed(time.Now().UTC().UnixNano())
		}
	}
}
