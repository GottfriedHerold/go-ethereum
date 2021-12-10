package bandersnatch

// Point_axtw describes points on the p253-subgroup of the Bandersnatch curve in affine extended twisted Edwards coordinates.
// Extended means that we additionally store T with T = X*Y.
// a Point_axtw with coos x:y:t corresponds to a Point_xtw with coos x:y:t:1 (i.e. with z==1). Note that on the p253 subgroup, all points have z!=0.
type Point_axtw struct {
	x FieldElement
	y FieldElement
	t FieldElement
}

// NeutralElement_axtw denotes the Neutral Element of the Bandersnatch curve in affine extended twisted Edwards coordinates.
var NeutralElement_axtw Point_axtw = Point_axtw{x: FieldElementZero, y: FieldElementOne, t: FieldElementZero}

// X_affine returns the X coordinate of the given point in affine twisted Edwards coordinates, i.e. X/Z
func (p *Point_axtw) X_affine() FieldElement {
	return p.x
}

// Y_affine returns the Y coordinate of the given point in affine twisted Edwards coordinates, i.e. Y/Z
func (p *Point_axtw) Y_affine() FieldElement {
	return p.y
}

// X_projective returns the X coordinate of the given point P in projective twisted Edwards coordinates.
// Note that in general, calling functions on P other than X_projective(), Y_projective(), Z() might change the representations of P at will,
// so callers must not interleave calling other functions.
func (p *Point_axtw) X_projective() FieldElement {
	return p.x
}

// Y_projective returns the Y coordinate of the given point P in projective twisted Edwards coordinates.
// Note that calling functions on P other than X_projective(), Y_projective(), Z() might change the representations of P at will,
// so callers must not interleave calling other functions.
func (p *Point_axtw) Y_projective() FieldElement {
	return p.y
}

// Z_projective returns the Z coordinate of the given point P in projective twisted Edwards coordinates.
// Note that calling functions on P other than X_projective(), Y_projective(), Z() might change the representations of P at will,
// so callers must not interleave calling other functions.
func (p *Point_axtw) Z_projective() FieldElement {
	return FieldElementOne
}

// func (P *Point_axtw) MakeAffine() {
//	// Do nothing.
//}

// func (p *Point_axtw) IsAffine() bool {
// 	return true
// }

// AffineExtended returns a copy of the point in affine extended coordinates (i.e. a copy)
func (p *Point_axtw) AffineExtended() Point_axtw {
	// technically, we could return *p. There is no way for the caller to modify it without copying it on the caller side.
	return Point_axtw{x: p.x, y: p.y, t: p.t}
}

// ExtendedTwistedEdwards returns a copy of the point in extended twisted Edwards coordinates.
func (p *Point_axtw) ExtendedTwistedEdwards() Point_xtw {
	return Point_xtw{x: p.x, y: p.y, t: p.t, z: FieldElementOne}
}

// IsNeutralElement checks if the point P is the neutral element of the curve (modulo the identification of P with P+A).
// Use IsNeutralElement_exact if you do not want this identification.
func (p *Point_axtw) IsNeutralElement() bool {

	// NOTE: This is only correct since we work modulo the affine order-2 point (x=0, y=-c, t=0, z=c).
	if p.x.IsZero() {
		if p.y.IsZero() {
			return handle_errors("When checking whether an axtw point is the neutral element, an NaP was encountered", true, p)
		}
		return true
	}
	return false
}

// IsNeutralElement_exact tests for zero-ness. It does *NOT* identify P with P+A. We only assume that x,y,t,z satisfy the curve equations.
func (p *Point_axtw) IsNeutralElement_exact() bool {
	if !p.x.IsZero() {
		return false
	}
	if p.y.IsZero() {
		return handle_errors("When checking whether an axtw point is exactly the neutral element, a NaP was encountered", true, p)
	}
	if !p.t.IsZero() {
		panic("axtw Point with x==0, y!=0, t!=0 encountered. This must never happen")
	}
	return p.y.IsOne() // p.y must be either 1 or -1
}

// SetNeutral sets the Point p to the neutral element of the curve.
func (p *Point_axtw) SetNeutral() {
	*p = NeutralElement_axtw
}

// IsSingular checks whether the point is singular (x==y==0, indeed most likely x==y==t==0). Singular points must never appear if the library is used correctly. They can appear by
// a) performing operations on points that are not in the correct subgroup
// b) zero-initialized points are singular (Go lacks constructors to fix that).
func (p *Point_axtw) IsSingular() bool {
	return p.x.IsZero() && p.y.IsZero()
}

// IsAtInfinity tests whether the point is an infinite (neccessarily order-2) point. Since these points cannot be represented in affine coordinates in the first place, this always returns false.
func (p *Point_axtw) IsAtInfinity() bool {
	if p.IsSingular() {
		return handle_errors("When chekcking whether an axtw point is infinite, a NaP was encountered", false, p)
	}
	return false
}

// IsEqual compares two curve points for equality, working modulo the P = P + A identification. The two points do not have the be in the same coordinate format.
// TODO: Export variants for specific non-interface types to get more type safety?
func (p *Point_axtw) IsEqual(other CurvePointRead) bool {
	switch other_real := other.(type) {
	case *Point_xtw:
		return p.is_equal_at(other_real)
	case *Point_axtw:
		return p.is_equal_aa(other_real)
	default:
		if p.IsSingular() || other.IsSingular() {
			return handle_errors("When comparing an axtw point with another point, a NaP was encountered", true, p, other_real)
		}
		var temp1, temp2 FieldElement
		var temp_fe FieldElement = other_real.Y_projective()
		temp1.Mul(&p.x, &temp_fe)
		temp_fe = other_real.X_projective()
		temp2.Mul(&p.y, &temp_fe)
		return temp1.IsEqual(&temp2)
	}
}

// IsEqual_exact compares two curve points for equality WITHOUT working modulo the P = P+A identification. The two points do not have to be in the same coordinate format.
func (p *Point_axtw) IsEqual_exact(other CurvePointRead) bool {
	if p.IsSingular() || other.IsSingular() {
		return handle_errors("When comparing an axtw point exactly with another point, a NaP was encountered", true, p, other)
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

// SetFrom initializes the point from the given input point (which may have a different coordinate format)
func (p *Point_axtw) SetFrom(input CurvePointRead) {
	*p = input.AffineExtended()
}

// Clone creates a copy of the given point as a CurvePointRead. (Be aware that this interface stores pointers)
func (p *Point_axtw) Clone() CurvePointRead {
	p_copy := *p
	return &p_copy
}

// String prints the point in X:Y:T - format
func (p *Point_axtw) String() string {
	// Not the most efficient way, but good enough.
	return p.x.String() + ":" + p.y.String() + ":" + p.t.String()
}

// Add performs curve point addition according to the group law.
// Use p.Add(&x, &y) for p := x + y.
// TODO: Export variants for specific types
func (p *Point_axtw) Add(x, y CurvePointRead) {
	var temp Point_xtw
	temp.Add(x, y)
	p.SetFrom(&temp)
}

// Sub performs curve point addition according to the group law.
// Use p.Sub(&x, &y) for p := x - y.
// TODO: Export variants for specific types
func (p *Point_axtw) Sub(x, y CurvePointRead) {
	var temp Point_xtw
	temp.Sub(x, y)
	p.SetFrom(&temp)
}

// Neg computes the negative of the point wrt the elliptic curve group law.
// Use p.Neg(&input) for p := -input.
func (p *Point_axtw) Neg(input CurvePointRead) {
	var temp Point_xtw
	temp.Neg(input)
	p.SetFrom(&temp)
}

// Endo computes the efficient order-2 endomorphism on the given point.
func (p *Point_axtw) Endo(input CurvePointRead) {
	var temp Point_xtw
	temp.Endo(input)
	p.SetFrom(&temp)
}

// Endo_safe computes the efficient order-2 endomorphism on the given input point (of any coordinate format).
// This function works even if the input may be a point at infinity; note that the output is never at infinity anyway.
// Be aware that, depending on interpretation, the statement that the endomorpism acts by multiplication by the constant sqrt(2) mod p253 is only true on the good subgroup.
func (p *Point_axtw) Endo_safe(input CurvePointRead) {
	var temp Point_xtw
	temp.Endo_safe(input)
	p.SetFrom(&temp)
}

// EndoEq applies the endomorphism on the given point. p.EndoEq() is shorthand for p.EndoEq(&p)
func (p *Point_axtw) EndoEq() {
	var temp Point_xtw
	temp.computeEndomorphism_ta(p)
	p.SetFrom(&temp)
}

// NeqEq replaces the given point by its negative (wrt the elliptic curve group addition law)
func (p *Point_axtw) NegEq() {
	p.y.NegEq()
	p.t.NegEq()
}

// AddEq adds (via the elliptic curve group addition law) the given curve point x (in any coordinate format) to the received p, overwriting p.
func (p *Point_axtw) AddEq(x CurvePointRead) {
	p.Add(p, x)
}

// SubEq subtracts (via the elliptic curve group addition law) the given curve point x (in any coordinate format) from the received p, overwriting p.
func (p *Point_axtw) SubEq(x CurvePointRead) {
	p.Sub(p, x)
}
