package bandersnatch

import (
	"math/big"
	"math/rand"
	"testing"
)

// naive implementation using the affine definition. This is just used to test the other formulas against.
func (out *xtw_edwards_point) add_xxx_naive(input1, input2 *xtw_edwards_point) {
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

// Creates a random point on the curve, which need not be in the correct subgroup.
func make_random_twedwards_full(rnd *rand.Rand) xtw_edwards_point {

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
	return xtw_edwards_point{x: x, y: y, z: z, t: t}
}

func TestExampleIsGenerator(t *testing.T) {
	if !xtw_edwards_point_neutral.verify_Point_on_Curve() {
		t.Fatal("Neutral element not on curve")
	}
	if !example_generator_xtw.verify_Point_on_Curve() {
		t.Fatal("Example point is not on curve")
	}
}

func TestRandomSampling(t *testing.T) {
	const iterations = 1000
	var drng *rand.Rand = rand.New(rand.NewSource(666))
	for i := 0; i < iterations; i++ {
		p := make_random_twedwards_full(drng)
		if !p.verify_Point_on_Curve() {
			t.Fatal("Randomly generated curve point is actually not on curve", i)
		}
	}
}

func Test_add_xxx(t *testing.T) {
	const iterations = 1000
	var drng *rand.Rand = rand.New(rand.NewSource(666))
	var p1, p2, res1, res2 xtw_edwards_point

	res1.add_xxx(&xtw_edwards_point_neutral, &xtw_edwards_point_neutral)
	if !res1.verify_Point_on_Curve() {
		t.Fatal("0+0 not on curve for add_xxx")
	}
	if !res1.is_equal_xx(&xtw_edwards_point_neutral) {
		t.Fatal("0 + 0 != 0 on curve for add_xxx")
	}

	for i := 0; i < iterations; i++ {

		p1 = make_random_twedwards_full(drng)
		p2.add_xxx(&p1, &xtw_edwards_point_neutral)
		if !p2.verify_Point_on_Curve() {
			t.Fatal("x + 0 is not on curve for random x on curve in add_xxx")
		}
		if !p1.is_equal_xx(&p2) {
			t.Fatal("x + 0 != x for random x in add_xxx")
		}
		p2.add_xxx(&xtw_edwards_point_neutral, &p1)
		if !p2.verify_Point_on_Curve() {
			t.Fatal("0 + x is not on curve for random x on curve in add_xxx")
		}
		if !p1.is_equal_xx(&p2) {
			t.Fatal("0 + x != x for random x in add_xxx")
		}

		p2 = make_random_twedwards_full(drng)
		_ = p2.verify_Point_on_Curve()
		_ = p1.verify_Point_on_Curve()
		res1.add_xxx(&p1, &p2)
		res2.add_xxx_naive(&p1, &p2)
		if !res1.verify_Point_on_Curve() {
			t.Fatal("Result of curve addition not on curve for add_xxx")
		}
		if !res1.is_equal_xx(&res2) {
			t.Fatal("Results of curve addition do not match for add_xxx and add_xxx_naive")
		}
		res2.add_xxx(&p2, &p1)
		if !res1.is_equal_xx(&res2) {
			t.Fatal("x + y != y + x for random x, y with add_xxx")
		}
	}
}
