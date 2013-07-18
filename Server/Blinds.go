package Server

type BlindLevel struct {
	sb   int
	ante int
}

type Blinds struct {
	blinds []BlindLevel
}

func (this *Blinds) Init(sbs, antes []int) {
	levels := len(sbs)
	this.blinds = make([]BlindLevel, levels)
	for i := 0; i < levels; i++ {
		this.blinds[i].sb = sbs[i]
		this.blinds[i].ante = antes[i]
	}
}

func (this Blinds) GetSB(level int) int {
	return this.blinds[level].sb
}

func (this Blinds) GetAnte(level int) int {
	return this.blinds[level].ante
}
