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

// IsZero checks if the point P is the neutral element of the curve. This function assumes that P is on the curve *AND* in the subgroup.
func (P *Point_axtw) IsZero() bool {
	// NOTE: This asserts that P is in the correct subgroup. Otherwise the point point (x=0, y=-c, t=0, z=c) would also return true
	return P.x.IsZero()
}

// Sets the Point P to the neutral element of the curve.
func (P *Point_axtw) SetZero() {
	*P = NeutralElement_axtw
}

// internal function that tests for zero-ness even if we do not know that the point is in the subgroup. We only assume that x,y satisfy the curve equation.
func (P *Point_axtw) isZero_safe() bool {
	return P.x.IsZero() && P.t.IsZero() && P.y.IsOne()
}
