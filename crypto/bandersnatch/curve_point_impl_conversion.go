package bandersnatch

func (p *Point_xtw) makeAffine_x() {
	var temp FieldElement
	if p.z.IsZero() {
		panic("Trying to make point at infinity or singular point affine")
	}
	temp.Inv(&p.z)
	p.x.MulEq(&temp)
	p.y.MulEq(&temp)
	p.t.MulEq(&temp)
	p.z.SetOne()
}
