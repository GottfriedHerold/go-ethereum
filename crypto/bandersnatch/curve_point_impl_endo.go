package bandersnatch

/*
	Note: Suffixes like _ttt or _tta refer to the type of input point (with order output, input1 [,input2] )
	t denote extended projective,
	a denotes extended affine (i.e. Z==1)
	s denotes double-projective
*/

// computeEndomorphism_tt computes the GLV Endomorphism (degree-2 isogeny with kernel {Neutral, Affine oder-2}) on a given input point. It is valid unless input is at infinity.
// Note that our identification of P with P+A is taken care of automatically.
func (output *Point_xtw) computeEndomorphism_tt(input *Point_xtw) {
	// The formula used below is valid unless for the input xy==zt is zero, which happens iff the input has order 2 or 1.
	if input.x.IsZero() {
		// Since we assume to be on the p253 subgroup/identify P = P+A, we know that the input is actually the neutral element, so the output is the neutral element.
		// Note that for the other point with x==0 (i.e. the affine order-2 point A), outputting the neutral element is actually correct even without the identification.
		// To avoid problems, we verify that the input is not singular.
		if input.IsSingular() {
			// TODO: Panic / Log? We set the output to the input to maintain the propery that Operation(singularity) == singularity, i.e. singularity has NaN-like behaviour.
			*output = Point_xtw{}
			return
		}
		*output = NeutralElement_xtw
		return
	}
	var bzz, yy, A, B, C, D FieldElement
	bzz.Square(&input.z)
	yy.Square(&input.y)
	A.Sub(&bzz, &yy)
	A.MulEq(&endo_c_fe) // A = c*(z^2 - y^2)

	bzz.MulEq(&endo_b_fe)
	B.Sub(&yy, &bzz) // B = y^2 - bz^2

	C.Add(&yy, &bzz)
	C.MulEq(&endo_b_fe) // C = b(y^2 + bz^2)

	D.Mul(&input.t, &input.z) // D = t*z == x*y

	output.x.Mul(&A, &B)
	output.y.Mul(&C, &D)
	output.t.Mul(&A, &C)
	output.z.Mul(&B, &D)
}

// same as above, but with z == 1 for the input
func (output *Point_xtw) computeEndomorphism_ta(input *Point_axtw) {
	// The formula used below is valid unless for the input xy==zt is zero, which happens iff the input has order 2 or 1.
	if input.x.IsZero() {
		// Since we assume to be on the p253 subgroup, we know that the input is actuall the neutral element, so the output is the neutral element.
		// Note that for the other point with x==0 (i.e. the affine order-2 point), outputting the neutral element is actually correct.
		// To avoid problems, we verify that the input is not singular.
		if input.IsSingular() {
			// TODO: Panic / Log? We set the output to the input to maintain the propery that Operation(singularity) == singularity, i.e. singularity has NaN-like behaviour.
			*output = Point_xtw{}
			return
		}
		*output = NeutralElement_xtw
		return
	}
	var yy, A, B, C FieldElement
	// bzz.Square(&input.z)
	yy.Square(&input.y)
	A.Sub(&FieldElementOne, &yy)
	A.MulEq(&endo_c_fe) // A = c*(z^2 - y^2) == c*(1-y^2)

	// bzz.MulEq(&endo_b_fe)
	B.Sub(&yy, &endo_b_fe) // B = y^2 - bz^2 == y^2 - b

	C.Add(&yy, &endo_b_fe)
	C.MulEq(&endo_b_fe) // C = b(y^2 + bz^2) == b (y^2 + b)

	// D == t
	// D.Mul(&input.t, &input.z) // D = t*z == x*y

	output.x.Mul(&A, &B)
	output.y.Mul(&C, &input.t)
	output.t.Mul(&A, &C)
	output.z.Mul(&B, &input.t)
}
