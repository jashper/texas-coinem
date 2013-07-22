package Server

import (
	"math/rand"
	"time"
)

type EntropyPool struct {
	Decks chan [52]int
}

func (this *EntropyPool) Init(bufCount, reseed int) {
	this.Decks = make(chan [52]int, bufCount)
	rand.Seed(time.Now().UTC().UnixNano())

	go this.run(reseed)
}

func (this EntropyPool) run(reseed int) {
	for count := 1; ; count++ {
		var testDeck [52]int

		cards := make([]int, 52)
		for c := 1; c <= 52; c++ {
			cards[c-1] = c
		}
		for i := 0; i < 51; i++ {
			toPop := rand.Intn(52 - i)
			testDeck[i] = cards[toPop]
			cards = append(cards[:toPop], cards[toPop+1:]...)
		}
		testDeck[51] = cards[0]

		this.Decks <- testDeck

		if count%reseed == 0 {
			rand.Seed(time.Now().UTC().UnixNano())
		}
	}
}

func cardToString(card int) (cardStr string) {
	rank := (card - 1) / 4

	if rank == 0 {
		cardStr = "2"
	} else if rank == 1 {
		cardStr = "3"
	} else if rank == 2 {
		cardStr = "4"
	} else if rank == 3 {
		cardStr = "5"
	} else if rank == 4 {
		cardStr = "6"
	} else if rank == 5 {
		cardStr = "7"
	} else if rank == 6 {
		cardStr = "8"
	} else if rank == 7 {
		cardStr = "9"
	} else if rank == 8 {
		cardStr = "T"
	} else if rank == 9 {
		cardStr = "J"
	} else if rank == 10 {
		cardStr = "Q"
	} else if rank == 11 {
		cardStr = "K"
	} else if rank == 12 {
		cardStr = "A"
	}

	suit := card % 4
	if suit == 0 {
		cardStr += "s"
	} else if suit == 1 {
		cardStr += "c"
	} else if suit == 2 {
		cardStr += "d"
	} else if suit == 3 {
		cardStr += "h"
	}

	return
}
