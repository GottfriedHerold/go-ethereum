package bandersnatch

// Point_axtw describes points on the p253-subgroup of the Bandersnatch curve in affine extended twisted Edwards coordinates.
// Extended means that we additionally store T with T = X*Y.
type Point_axtw struct {
	x FieldElement
	y FieldElement
	t FieldElement
}

// NeutralElement_axtw denotes the Neutral Element of the Bandersnatch curve in affine extended twisted Edwards coordinates.
var NeutralElement_axtw Point_axtw = Point_axtw{x: FieldElementZero, y: FieldElementOne, t: FieldElementZero}

// X_affine returns the X coordinate of the given point in affine twisted Edwards coordinates.
func (P *Point_axtw) X_affine() FieldElement {
	return P.x
}

// Y_affine returns the Y coordinate of the given point in affine twisted Edwards coordinates.
func (P *Point_axtw) Y_affine() FieldElement {
	return P.y
}

// X_projective returns the X coordinate of the given point P in projective twisted Edwards coordinates.
// Note that in general, calling functions on P other than X_projective(), Y_projective(), Z() might change the representations of P at will,
// so callers must not interleave calling other functions.
func (P *Point_axtw) X_projective() FieldElement {
	return P.x
}

// Y_projective returns the Y coordinate of the given point P in projective twisted Edwards coordinates.
// Note that calling functions on P other than X_projective(), Y_projective(), Z() might change the representations of P at will,
// so callers must not interleave calling other functions.
func (P *Point_axtw) Y_projective() FieldElement {
	return P.y
}

// Z_projective returns the Z coordinate of the given point P in projective twisted Edwards coordinates.
// Note that calling functions on P other than X_projective(), Y_projective(), Z() might change the representations of P at will,
// so callers must not interleave calling other functions.
func (P *Point_axtw) Z_projective() FieldElement {
	return FieldElementOne
}

func (P *Point_axtw) MakeAffine() {
	// Do nothing.
}

func (p *Point_axtw) IsAffine() bool {
	return true
}

func (p *Point_axtw) AffineExtended() Point_axtw {
	return Point_axtw{x: p.x, y: p.y, t: p.t}
}

func (p *Point_axtw) ExtendedTwistedEdwards() Point_xtw {
	return Point_xtw{x: p.x, y: p.y, t: p.t, z: FieldElementOne}
}

// IsNeutralElement checks if the point P is the neutral element of the curve (modulo the identification of P with P+A).
// Use IsNeutralElement_exact if you do not want this identification.
func (P *Point_axtw) IsNeutralElement() bool {

	// NOTE: This asserts that P is in the correct subgroup or that we work modulo the affine order-2 point (x=0, y=-c, t=0, z=c).
	if P.x.IsZero() {
		if P.y.IsZero() {
			// TODO: Handle error: Singular point
			return false
		}
		return true
	}
	return false
}

// IsNeutralElement_exact tests for zero-ness. It does *NOT* identify P with P+A. We only assume that x,y,t,z satisfy the curve equations.
func (P *Point_axtw) IsNeutralElement_exact() bool {
	return P.x.IsZero() && P.y.IsOne() && P.t.IsZero()
}

// SetNeutral sets the Point P to the neutral element of the curve.
func (P *Point_axtw) SetNeutral() {
	*P = NeutralElement_axtw
}

// IsSingular checks whether the point is singular (x==y==0, indeed most likely x==y==t==0). Singular points must never appear if the library is used correctly. They can appear by
// a) performing operations on points that are not in the correct subgroup
// b) zero-initialized points are singular (Go lacks constructors to fix that).
func (p *Point_axtw) IsSingular() bool {
	return p.x.IsZero() && p.y.IsZero()
}

func (p *Point_axtw) IsAtInfinity() bool {
	return false
}

func (p *Point_axtw) IsEqual(other CurvePointRead) bool {
	switch other_real := other.(type) {
	case *Point_xtw:
		return p.is_equal_at(other_real)
	case *Point_axtw:
		return p.is_equal_aa(other_real)
	default:
		if p.IsSingular() || other.IsSingular() {
			// TODO: Handle error
			return false
		}
		var temp1, temp2 FieldElement
		var temp_fe FieldElement = other_real.Y_projective()
		temp1.Mul(&p.x, &temp_fe)
		temp_fe = other_real.X_projective()
		temp2.Mul(&p.y, &temp_fe)
		return temp1.IsEqual(&temp2)
	}
}

func (p *Point_axtw) IsEqualExact(other CurvePointRead) bool {
	if p.IsSingular() || other.IsSingular() {
		// TODO: Error handling
		return false
	}
	switch other_real := other.(type) {
	case *Point_xtw:
		return p.is_equal_exact_at(other_real)
	case *Point_axtw:
		return p.is_equal_exact_aa(other_real)
	default:
		other_temp := other.ExtendedTwistedEdwards()
		return p.is_equal_exact_at(&other_temp)
	}
}

func (p *Point_axtw) SetFrom(input CurvePointRead) {
	*p = input.AffineExtended()
}

func (p *Point_axtw) Add(x, y CurvePointRead) {
	var temp Point_xtw
	temp.Add(x, y)
	p.SetFrom(&temp)
}

func (p *Point_axtw) Sub(x, y CurvePointRead) {
	var temp Point_xtw
	temp.Sub(x, y)
	p.SetFrom(&temp)
}

func (p *Point_axtw) Neg(input CurvePointRead) {
	var temp Point_xtw
	temp.Neg(input)
	p.SetFrom(&temp)
}

func (p *Point_axtw) Endo(input CurvePointRead) {
	var temp Point_xtw
	temp.Endo(input)
	p.SetFrom(&temp)
}

func (p *Point_axtw) Endo_safe(input CurvePointRead) {
	var temp Point_xtw
	temp.Endo_safe(input)
	p.SetFrom(&temp)
}

func (p *Point_axtw) EndoEq() {
	var temp Point_xtw
	temp.computeEndomorphism_ta(p)
	p.SetFrom(&temp)
}

func (p *Point_axtw) NegEq() {
	p.y.NegEq()
	p.t.NegEq()
}

func (p *Point_axtw) AddEq(x CurvePointRead) {
	p.Add(p, x)
}

func (p *Point_axtw) SubEq(x CurvePointRead) {
	p.Sub(p, x)
}
