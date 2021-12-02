package bandersnatch

/*
	Note: Suffixes like _ttt or _tta refer to the type of input point (with order output, input1 [,input2] )
	t denote extended projective,
	a denotes extended affine (i.e. Z==1)
	s denotes double-projective
*/

func (out *Point_xtw) neg_tt(input1 *Point_xtw) {
	out.x.Neg(&input1.x)
	out.y = input1.y
	out.t.Neg(&input1.t)
	out.z = input1.z
}

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
