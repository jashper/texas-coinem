package Server

import (
	"bufio"
	"os"
)

const HAND_RANK_SIZE = 32476671

type HandEvaluator struct {
	hr [HAND_RANK_SIZE]uint
}

func (this *HandEvaluator) Init(path string) (err error) {
	var file *os.File

	if file, err = os.Open(path); err != nil {
		return err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	var b [4]byte
	for idx := 0; idx < HAND_RANK_SIZE; idx++ {
		b[0], err = reader.ReadByte()
		b[1], err = reader.ReadByte()
		b[2], err = reader.ReadByte()
		b[3], err = reader.ReadByte()
		this.hr[idx] = uint(b[0]) + (uint(b[1]) << 8) +
			(uint(b[2]) << 16) + (uint(b[3]) << 24)
	}

	return nil
}

func (this HandEvaluator) HandInfo(cards [7]uint) (value, category uint) {
	value = this.hr[ 53 + cards[0] ]
	value = this.hr[ value + cards[1] ]
	value = this.hr[ value + cards[2] ]
	value = this.hr[ value + cards[3] ]
	value = this.hr[ value + cards[4] ]
	value = this.hr[ value + cards[5] ]
	value = this.hr[ value + cards[6] ]
	category = value >> 12
	return
}