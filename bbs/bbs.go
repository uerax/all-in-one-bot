package bbs

type Bbs struct {
	Bitcointalk *Bitcointalk
	Nodeseek *Nodeseek
}

func NewBbs() *Bbs {
	return &Bbs{
		Bitcointalk: NewBitcointalk(),
		Nodeseek: NewNodeseek(),
	}
}

