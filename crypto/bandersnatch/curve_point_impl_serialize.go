package bandersnatch

func (p *Point_axtw) specialSerialzeXCoo_a() (ret FieldElement) {
	ret = p.x
	if p.y.Sign() < 0 {
		ret.NegEq()
	}
	return
}

func (p *Point_axtw) specialSerialzeYCoo_a() (ret FieldElement) {
	ret = p.y
	switch p.x.Sign() {
	// case 1: do nothing
	case -1:
		ret.NegEq()
	case 0:
		ret.SetZero()
	}
	return
}
