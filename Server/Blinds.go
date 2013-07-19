package Server

type BlindLevel struct {
	sb   float64
	ante float64
}

type Blinds struct {
	blinds []BlindLevel
}

func (this *Blinds) Init(sbs, antes []float64) {
	levels := len(sbs)
	this.blinds = make([]BlindLevel, levels)
	for i := 0; i < levels; i++ {
		this.blinds[i].sb = sbs[i]
		this.blinds[i].ante = antes[i]
	}
}

func (this Blinds) GetSB(level int) float64 {
	return this.blinds[level].sb
}

func (this Blinds) GetAnte(level int) float64 {
	return this.blinds[level].ante
}
