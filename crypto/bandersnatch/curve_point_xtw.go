package bandersnatch

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// Point_xtw describes points on the p253-subgroup of the Bandersnatch curve in extended twisted Edwards coordinates.
// Extended means that we additionally store T with T = X*Y/Z. Note that Z is never 0.)
// cf. https://iacr.org/archive/asiacrypt2008/53500329/53500329.pdf
type Point_xtw struct {
	x FieldElement
	y FieldElement
	z FieldElement
	t FieldElement
}

// example point on the subgroup specified in the bandersnatch paper
var example_generator_x *big.Int = new(big.Int).SetBytes(common.FromHex("0x29c132cc2c0b34c5743711777bbe42f32b79c022ad998465e1e71866a252ae18"))
var example_generator_y *big.Int = new(big.Int).SetBytes(common.FromHex("0x2a6c669eda123e0f157d8b50badcd586358cad81eee464605e3167b6cc974166"))
var example_generator_t *big.Int = new(big.Int).Mul(example_generator_x, example_generator_y)
var example_generator_xtw Point_xtw = func() (ret Point_xtw) {
	ret.x.SetInt(example_generator_x)
	ret.y.SetInt(example_generator_y)
	ret.t.SetInt(example_generator_t)
	ret.z.SetOne()
	return
}()

/*
	Basic functions for Point_xtw
*/

// NeutralElement_<foo> denotes the Neutral Element of the Bandersnatch curve.
var (
	NeutralElement_xtw Point_xtw = Point_xtw{x: FieldElementZero, y: FieldElementOne, t: FieldElementZero, z: FieldElementOne}
)

// X_affine returns the X coordinate of the given point in affine twisted Edwards coordinates.
func (P *Point_xtw) X_affine() FieldElement {
	P.make_affine_x()
	return P.x
}

// Y_affine returns the Y coordinate of the given point in affine twisted Edwards coordinates.
func (P *Point_xtw) Y_affine() FieldElement {
	P.make_affine_x()
	return P.y
}

// X_projective returns the X coordinate of the given point P in projective twisted Edwards coordinates.
// Note that calling functions on P other than X_projective(), Y_projective(), Z() might change the representations of P at will,
// so callers must not interleave calling other functions.
func (P *Point_xtw) X_projective() FieldElement {
	return P.x
}

// Y_projective returns the Y coordinate of the given point P in projective twisted Edwards coordinates.
// Note that calling functions on P other than X_projective(), Y_projective(), Z() might change the representations of P at will,
// so callers must not interleave calling other functions.
func (P *Point_xtw) Y_projective() FieldElement {
	return P.y
}

// Z_projective returns the Z coordinate of the given point P in projective twisted Edwards coordinates.
// Note that calling functions on P other than X_projective(), Y_projective(), Z() might change the representations of P at will,
// so callers must not interleave calling other functions.
func (P *Point_xtw) Z_projective() FieldElement {
	return P.z
}

func (P *Point_xtw) MakeAffine() {
	if !P.IsAffine() {
		P.make_affine_x()
	}
}

func (p *Point_xtw) IsAffine() bool {
	return p.z.IsOne()
}

func (p *Point_xtw) AffineExtended() Point_axtw {
	p.MakeAffine()
	return Point_axtw{x: p.x, y: p.y, t: p.t}
}

// IsZero checks if the point P is the neutral element of the curve. This function assumes that P is on the curve *AND* in the subgroup.
func (P *Point_xtw) IsZero() bool {
	// NOTE: This asserts that P is in the correct subgroup. Otherwise the point point (x=0, y=-c, t=0, z=c) would also return true
	return P.x.IsZero()
}

// Sets the Point P to the neutral element of the curve.
func (P *Point_xtw) SetZero() {
	*P = NeutralElement_xtw
}

// internal function that tests for zero-ness even if we do not know that the point is in the subgroup. We only assume that x,y,t,z satisfy the curve equation. Returns false for z == 0
func (P *Point_xtw) isZero_safe() bool {
	// We check this separately, because we might log this event.
	if P.z.IsZero() {
		return false
	}
	return P.x.IsZero() && P.t.IsZero() && P.y.IsEqual(&P.z)
}

/*
	Note: Suffixes like _xxx or _xxa refer to the type of input point (with order output, input1 [,input2] )
	x denote extended projective,
	a denotes extended affine
*/

// https://www.hyperelliptic.org/EFD/g1p/auto-twisted-extended.html#addition-add-2008-hwcd, due to Hisil–Wong–Carter–Dawson 2008, http://eprint.iacr.org/2008/522, Section 3.1.
func (out *Point_xtw) add_xxx(input1, input2 *Point_xtw) {
	var A, B, C, D, E, F, G, H FieldElement // We follow the notation of the link above

	A.Mul(&input1.x, &input2.x) // A = X1 * X2
	B.Mul(&input1.y, &input2.y) // B = Y1 * Y2
	C.Mul(&input1.t, &input2.t)
	C.MulEq(&TwistedEdwardsD_fe) // C = d * T1 * T2
	D.Mul(&input1.z, &input2.z)  // D = Z1 * Z2
	E.Add(&input1.x, &input1.y)
	F.Add(&input2.x, &input2.y) // F serves as temporary
	E.MulEq(&F)
	E.SubEq(&A)
	E.SubEq(&B)   // E = (X1 + Y1) * (X2 + Y2) - A - B == X1*Y2 + Y1*X2
	F.Sub(&D, &C) // F = D - C
	G.Add(&D, &C) // G = D + C

	A.multiply_by_five()
	H.Add(&B, &A) // H = B + 5X1 * X2 = Y1*Y2 - a*X1*X2  (a=-5 is a parameter of the curve)

	out.x.Mul(&E, &F) // X3 = E * F
	out.y.Mul(&G, &H) // Y3 = G * H
	out.t.Mul(&E, &H) // T3 = E * H
	out.z.Mul(&F, &G) // Z3 = F * G
	if !out.verify_Point_on_Curve() {
		E.Inv(&out.t)
		out.x.MulEq(&E)
		out.y.MulEq(&E)
		out.t.MulEq(&E)
		out.z.MulEq(&E)
		// fmt.Print("Point is x==", out.x.String(), " y== ", out.y.String(), "t== ", out.t.String(), " z== ", out.z.String(), "\n")
	}
}

func (out *Point_xtw) double_xx(input1 *Point_xtw) {
	// TODO: Use https://www.hyperelliptic.org/EFD/g1p/auto-twisted-extended.html#doubling-dbl-2008-hwcd. Note that this formula gives the same result as add_xxx (modulo ax^2 + y^2 = z^2 + dt^2 and a global sign), hence
	// is strongly unified as well.
	out.add_xxx(input1, input1)
}

func (out *Point_xtw) neg_xx(input1 *Point_xtw) {
	out.x = input1.x
	out.y.Neg(&input1.y)
	out.t.Neg(&input1.t)
	out.z = input1.z
}

func (out *Point_xtw) sub_xxx(input1, input2 *Point_xtw) {
	out.neg_xx(input2)
	out.add_xxx(out, input1)
}

func (p1 *Point_xtw) is_equal_xx(p2 *Point_xtw) bool {
	var temp1, temp2 FieldElement
	temp1.Mul(&p1.x, &p2.z)
	temp2.Mul(&p1.z, &p2.x)

	if !temp1.IsEqual(&temp2) {
		return false
	}
	// TODO: This is just debugging code, actually:
	temp1.Mul(&p1.y, &p2.z)
	temp2.Mul(&p1.z, &p2.y)
	if !temp1.IsEqual(&temp2) {
		var temp3 FieldElement
		temp3.Neg(&temp1)
		if temp3.IsEqual(&temp2) {
			panic("Point not in subgroup")
		} else {
			panic("Point not on curve")
		}
	}
	return true
}

func (p *Point_xtw) make_affine_x() {
	var temp FieldElement
	if p.z.IsZero() {
		panic("Division by zero")
	}
	temp.Inv(&p.z)
	p.x.MulEq(&temp)
	p.y.MulEq(&temp)
	p.t.MulEq(&temp)
	p.z.SetOne()
}

func (p1 *Point_xtw) is_equal_safe_xx(p2 *Point_xtw) bool {
	var temp1, temp2 FieldElement
	temp1.Mul(&p1.x, &p2.z)
	temp2.Mul(&p1.z, &p2.x)
	return temp1.IsEqual(&temp2)
}

// Note: This does NOT verify that the point is in the correct subgroup.
func (p *Point_xtw) verify_Point_on_Curve() bool {
	if p.z.IsZero() {
		// fmt.Println("Point with z==0 encountered")
		return false
	}
	var u, v FieldElement
	u.Mul(&p.x, &p.y)
	v.Mul(&p.t, &p.z)
	if !u.IsEqual(&v) {
		fmt.Println("Point with inconsistent coordinates encountered")
		return false
	}
	// We now check the curve equation. Note that with z!=0, t/z == x/z * y/z, this can be simplified to ax^2 + y^2 = z^2 + d*t^2
	u.Mul(&p.t, &p.t)
	u.MulEq(&TwistedEdwardsD_fe) // u = d*t^2
	v.Mul(&p.z, &p.z)
	u.AddEq(&v) // u= dt^2 + z^2
	v.Mul(&p.y, &p.y)
	u.SubEq(&v) // u = z^2 + dt^2 - y^2
	v.Mul(&p.x, &p.x)
	v.multiply_by_five()
	u.AddEq(&v) // u = z^2 + dt^2 - y^2 + 5x^2 ==  z^2 + dt^2 - y^2 - ax^2
	if !u.IsZero() {
		fmt.Printf("Point not on curve encountered: x=0x%x y=0x%x z=0x%x t=0x%x", p.x.ToInt(), p.y.ToInt(), p.z.ToInt(), p.t.ToInt())
		return false
	}
	return true
}

func (p *Point_xtw) clearCofactor() {
	p.double_xx(p)
	p.double_xx(p)
}
