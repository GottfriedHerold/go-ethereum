package bandersnatch

// Point_efgh describes points (usually on the p253-subgroup of) the bandersnatch curve in E:G, H:F - coordinates (called double-projective), i.e.
// we represent X/Z as E/G and Y/Z as H/F. From a computational view, this effectively means that we use a separate denominator for X and Y (instead of a joint one Z).
// We can recover X:Y:Z coordinates by computing Z = F*G, X = E*G, Y = H*F. Then T = E*G. This is meaningful even if one of F,G is zero. Note that at most one of E,F,G,H must be zero.
// Observe that in fact all our formulae *produce* points in such coordinates in such a form and then transform them into the desired form
// Using double-projective coordinates can be used to make this explicit and can save computation if a coordinate is unused:
// The doubling formula and the endomorphism can be expressed in double-projective coordinates by first converting to extended twisted edwards and then computing (rather than the other way round).
// Since these formulae do not use the input's t coordinate, this saves a multiplication.
// On the p253-subgroup, the only coordinate that may be zero is e.
type Point_efgh struct {
	e FieldElement
	f FieldElement
	g FieldElement
	h FieldElement
}

var (
	NeutralElement_efgh     = Point_efgh{e: FieldElementZero, f: FieldElementOne, g: FieldElementOne, h: FieldElementOne}
	orderTwoPoint_efgh      = Point_efgh{e: FieldElementZero, f: FieldElementOne, g: FieldElementOne, h: FieldElementMinusOne}
	exceptionalPoint_1_efgh = Point_efgh{e: FieldElementOne, f: SqrtDDivA_fe, g: FieldElementZero, h: FieldElementOne}
	exceptionalPoint_2_efgh = Point_efgh{e: FieldElementOne, f: SqrtDDivA_fe, g: FieldElementZero, h: FieldElementMinusOne}
)

func (P *Point_efgh) is_normalized() bool {
	return P.f.IsOne() && P.g.IsOne()
}

func (P *Point_efgh) normalize_affine() {
	if P.is_normalized() {
		return
	}
	var temp FieldElement
	temp.Mul(&P.f, &P.g)
	if temp.IsZero() {
		panic("Trying to normalize singular or infinite point")
	}
	temp.Inv(&temp)
	P.e.MulEq(&P.f)
	P.h.MulEq(&P.g)
	P.e.MulEq(&temp)
	P.h.MulEq(&temp)
	P.f.SetOne()
	P.g.SetOne()
}

func (P *Point_efgh) X_affine() FieldElement {
	P.normalize_affine()
	return P.e
}

func (P *Point_efgh) Y_affine() FieldElement {
	P.normalize_affine()
	return P.h
}

func (P *Point_efgh) T_affine() (T FieldElement) {
	P.normalize_affine()
	T.Mul(&P.e, &P.h)
	return
}

func (P *Point_efgh) X_projective() (X FieldElement) {
	X.Mul(&P.e, &P.f)
	return
}

func (P *Point_efgh) Y_projective() (Y FieldElement) {
	Y.Mul(&P.g, &P.h)
	return
}

func (P *Point_efgh) T_projective() (T FieldElement) {
	T.Mul(&P.e, &P.h)
	return
}

func (P *Point_efgh) Z_projective() (Z FieldElement) {
	Z.Mul(&P.f, &P.g)
	return
}

func (P *Point_efgh) AffineExtended() (ret Point_axtw) {
	P.normalize_affine()
	ret.x = P.e
	ret.y = P.h
	ret.t.Mul(&P.e, &P.h)
	return
}

func (P *Point_efgh) ExtendedTwistedEdwards() (ret Point_xtw) {
	ret.x.Mul(&P.e, &P.f)
	ret.y.Mul(&P.g, &P.h)
	ret.t.Mul(&P.e, &P.h)
	ret.z.Mul(&P.f, &P.g)
	return
}

func (P *Point_efgh) IsSingular() bool {
	return (P.e.IsZero() && P.h.IsZero()) || (P.f.IsZero() && P.g.IsZero())
}

func (P *Point_efgh) IsNeutralElement() bool {
	// The only points with e==0 are the neutral element and the affine order-2 point
	if P.e.IsZero() {
		if P.h.IsZero() || P.f.IsZero() { // Note f==0 <=> g==0 <=> P singular here
			// TODO: Error handling for singularity
			return false
		}
		return true
	}
	return false
}

func (P *Point_efgh) IsNeutralElement_exact() bool {
	return P.IsNeutralElement() && P.f.IsEqual(&P.h)
}

func (P *Point_efgh) IsAtInfinity() bool {
	// Only points with g==0 are at infinity
	if P.g.IsZero() {
		if P.f.IsZero() || P.e.IsZero() { // Note: e==0 <=> h==0 <=> Point is singular
			// TODO: Error handling: Point is singular!
			return false
		}
		return true
	}
	return false
}

func (P *Point_efgh) IsEqual(other CurvePointRead) bool {
	switch other_real := other.(type) {
	default:
		if P.IsSingular() || other.IsSingular() {
			// TODO: Handle error
			return false
		}
		var other_x = other_real.X_projective()
		var other_y = other_real.Y_projective()
		// other.x * P.y == other.y * P.y
		other_x.MulEq(&P.g)
		other_x.MulEq(&P.h)
		other_y.MulEq(&P.e)
		other_y.MulEq(&P.f)
		return other_x.IsEqual(&other_y)
	}
}

func (P *Point_efgh) IsEqualExact(other CurvePointRead) bool {
	temp := P.ExtendedTwistedEdwards()
	return temp.IsEqualExact(other)
}
