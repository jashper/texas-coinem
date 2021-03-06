package Server

/*
	This class stores the progression of blind and ante values, each
	level lasting a constant duration defined by GameParameters.
*/

// ############################################
//     Helper Struct
// ############################################

type BlindLevel struct {
	sb   float64 // the bb is twice this amount
	ante float64
}

// ############################################
//     Constructor Struct & Init
// ############################################

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
