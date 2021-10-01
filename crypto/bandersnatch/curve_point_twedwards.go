package bandersnatch

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// coordinates in extended projetive twisted edwards form
// (extended means that we additionally store T with T = X*Y/Z)
// cf. https://iacr.org/archive/asiacrypt2008/53500329/53500329.pdf
type xtw_edwards_point struct {
	x bsFieldElement_64
	y bsFieldElement_64
	z bsFieldElement_64
	t bsFieldElement_64
}

// example point on the subgroup specified in the bandersnatch paper
var example_generator_x *big.Int = new(big.Int).SetBytes(common.FromHex("0x29c132cc2c0b34c5743711777bbe42f32b79c022ad998465e1e71866a252ae18"))
var example_generator_y *big.Int = new(big.Int).SetBytes(common.FromHex("0x2a6c669eda123e0f157d8b50badcd586358cad81eee464605e3167b6cc974166"))
var example_generator_t *big.Int = new(big.Int).Mul(example_generator_x, example_generator_y)
var example_generator_xtw xtw_edwards_point = func() (ret xtw_edwards_point) {
	ret.x.SetInt(example_generator_x)
	ret.y.SetInt(example_generator_y)
	ret.t.SetInt(example_generator_t)
	ret.z.SetOne()
	return
}()

var xtw_edwards_point_neutral xtw_edwards_point = xtw_edwards_point{x: bsFieldElement_64_zero, y: bsFieldElement_64_one, t: bsFieldElement_64_zero, z: bsFieldElement_64_one}

func (P *xtw_edwards_point) SetZero() {
	*P = xtw_edwards_point_neutral
}

func (P *xtw_edwards_point) IsZero() bool {
	return *P == xtw_edwards_point_neutral
}

/*
	Note: Suffixes like _xxx or _xxa refer to the type of input point (with order output, input1 [,input2] )
	x denote extended projective,
	a denotes extended affine
*/

// https://www.hyperelliptic.org/EFD/g1p/auto-twisted-extended.html#addition-add-2008-hwcd, due to Hisil–Wong–Carter–Dawson 2008, http://eprint.iacr.org/2008/522, Section 3.1.
func (out *xtw_edwards_point) add_xxx(input1, input2 *xtw_edwards_point) {
	var A, B, C, D, E, F, G, H bsFieldElement_64 // We follow the notation of the link above

	A.Mul(&input1.x, &input2.x) // A = X1 * X2
	B.Mul(&input1.y, &input2.y) // B = Y1 * Y2
	C.Mul(&input1.t, &input2.t)
	C.MulEq(&TwistedEdwardsD_fe) // C = d * T1 * T2
	D.Mul(&input1.z, &input2.z)  // D = Z1 * Z2
	E.Add(&input1.x, &input1.y)
	F.Add(&input2.x, &input2.y) // F serves as temporary
	E.MulEq(&F)
	E.SubEq(&A)
	E.SubEq(&B)   // E = (X1 + X2) * (X2 + Y2) - A - B
	F.Sub(&D, &C) // F = D - C
	G.Add(&D, &C) // G = D + C

	A.multiply_by_five()
	H.Add(&B, &A) // H = B + 5X1 * X2 = Y1*Y2 - a*X1*X2  (a=-5 is a parameter of the curve)

	out.x.Mul(&E, &F) // X3 = E * F
	out.y.Mul(&G, &H) // Y3 = G * H
	out.t.Mul(&E, &H) // T3 = E * H
	out.z.Mul(&F, &G) // Z3 = F * G
}

func (p1 *xtw_edwards_point) is_equal_xx(p2 *xtw_edwards_point) bool {
	var temp1, temp2 bsFieldElement_64
	temp1.Mul(&p1.x, &p2.z)
	temp2.Mul(&p1.z, &p2.x)

	if !temp1.IsEqual(&temp2) {
		return false
	}
	temp1.Mul(&p1.y, &p2.z)
	temp2.Mul(&p1.z, &p2.y)
	return temp1.IsEqual(&temp2)
}

// Note: This does NOT verify that the point is in the correct subgroup.
func (p *xtw_edwards_point) verify_Point_on_Curve() bool {
	if p.z.IsZero() {
		fmt.Println("Point with z==0 encountered")
		return false
	}
	var u, v bsFieldElement_64
	u.Mul(&p.x, &p.y)
	v.Mul(&p.t, &p.z)
	if !u.IsEqual(&v) {
		fmt.Println("Point with inconsisten coordinates encountered")
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

// affine version of the above, i.e. corresponding z == 1
type axtw_edwards_point struct {
	x bsFieldElement_64
	y bsFieldElement_64
	t bsFieldElement_64
}
