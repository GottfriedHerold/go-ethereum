package bandersnatch

import (
	"math/big"
	"math/rand"
)

/*
	This file contains naive implementations of various elliptic curve operations that are not actually used in production, but only serve
	to compare the actual implementation against for the purpose of testing correctness and debugging.
*/

// naive implementation using the affine definition. This is just used to test the other formulas against.
func (out *Point_xtw) add_xxx_naive(input1, input2 *Point_xtw) {
	var x1, y1, z1inv, x2, y2, z2inv bsFieldElement_64

	z1inv.Inv(&input1.z)
	z2inv.Inv(&input2.z)
	x1.Mul(&input1.x, &z1inv)
	y1.Mul(&input1.y, &z1inv)
	x2.Mul(&input2.x, &z2inv)
	y2.Mul(&input2.y, &z2inv)

	var denom_common bsFieldElement_64
	denom_common.Mul(&x1, &x2)
	denom_common.MulEq(&y1)
	denom_common.MulEq(&y2)
	denom_common.MulEq(&TwistedEdwardsD_fe) // denom_common == dx1x2y1y2

	var denom_x, denom_y bsFieldElement_64
	denom_x.Add(&bsFieldElement_64_one, &denom_common) // denom_x = 1+dx1x2y1y2
	denom_y.Sub(&bsFieldElement_64_one, &denom_common) // denom_y = 1-dx1x2y1y2

	var numerator_x, numerator_y, temp bsFieldElement_64
	numerator_x.Mul(&x1, &y2)
	temp.Mul(&y1, &x2)
	numerator_x.AddEq(&temp) // x1y2+y1x2

	numerator_y.Mul(&x1, &x2)
	numerator_y.multiply_by_five()
	temp.Mul(&y1, &y2)
	numerator_y.AddEq(&temp) // x1x2 + 5y1y2 = x1x2 - ax1x2

	out.t.Mul(&numerator_x, &numerator_y)
	out.z.Mul(&denom_x, &denom_y)
	out.x.Mul(&numerator_x, &denom_y)
	out.y.Mul(&numerator_y, &denom_x)
}

// Creates a random point on the curve, which does not neccessarily need to be in the correct subgroup.
func make_random_twedwards_full(rnd *rand.Rand) Point_xtw {

	var x, x2, y, t, z, num, denom bsFieldElement_64
	var d bsFieldElement_64
	d.SetInt(TwistedEdwardsD_Int)

	for {
		x.setRandomUnsafe(rnd)
		// x.SetUInt64(1)

		// compute y = sqrt( (1-ax^2)/(1-dx^2) )

		x2.Mul(&x, &x)                            // x2 = x^2
		denom.Mul(&x2, &d)                        // denom = dx^2
		denom.Sub(&bsFieldElement_64_one, &denom) // denom = 1 - d*x^2
		x2.multiply_by_five()                     // x2 = 5x^2
		num.Add(&bsFieldElement_64_one, &x2)      // num = 1 + 5x^2 = 1 - ax^2
		numInt := num.ToInt()
		denomInt := denom.ToInt()
		// Note: denom,num != 0, because d and a are non-squares
		if big.Jacobi(numInt, BaseFieldSize)*big.Jacobi(denomInt, BaseFieldSize) == -1 {
			continue
		}
		yInt := big.NewInt(0)
		yInt.ModInverse(denomInt, BaseFieldSize) // y = 1/denom
		yInt.Mul(yInt, numInt)                   // y = num/denom
		yInt.Mod(yInt, BaseFieldSize)            // modular reduction
		yInt.ModSqrt(yInt, BaseFieldSize)        // y = sqrt(num/denom)
		y.SetInt(yInt)
		break
	}
	t.Mul(&x, &y)

	z.setRandomUnsafe(rnd)
	if z.IsZero() {
		z.SetOne()
	}
	x.MulEq(&z)
	y.MulEq(&z)
	t.MulEq(&z)
	return Point_xtw{x: x, y: y, z: z, t: t}
}

// Creates a random point on the correct subgroup
func make_random_x(rnd *rand.Rand) Point_xtw {
	r := make_random_twedwards_full(rnd)
	r.clearCofactor2()
	return r
}
