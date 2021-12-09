package bandersnatch

func (p *Point_xtw) makeAffine_x() {
	var temp FieldElement
	if p.z.IsZero() {
		if p.IsSingular() {
			handle_errors("Try to converting invalid point xtw to coos with z==1", false, p)
			*p = Point_xtw{z: FieldElementOne} // invalid point
			return
		}
		panic("Trying to make point at infinity affine")
	}
	temp.Inv(&p.z)
	p.x.MulEq(&temp)
	p.y.MulEq(&temp)
	p.t.MulEq(&temp)
	p.z.SetOne()
}
