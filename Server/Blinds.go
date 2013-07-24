package Server

type BlindLevel struct {
	sb   float64
	ante float64
}

type Blinds struct {
	levels []BlindLevel
}

func (this *Blinds) Init(sbs, antes []float64) {
	levels := len(sbs)
	this.levels = make([]BlindLevel, levels)
	for i := 0; i < levels; i++ {
		this.levels[i].sb = sbs[i]
		this.levels[i].ante = antes[i]
	}
}
