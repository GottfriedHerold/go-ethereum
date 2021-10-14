package bandersnatch

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestExampleIsGenerator(t *testing.T) {
	if !NeutralElement_xtw.verify_Point_on_Curve() {
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
	var p1, p2, res1, res2 Point_xtw

	res1.add_xxx(&NeutralElement_xtw, &NeutralElement_xtw)
	if !res1.verify_Point_on_Curve() {
		t.Fatal("0+0 not on curve for add_xxx")
	}
	if !res1.is_equal_safe_xx(&NeutralElement_xtw) {
		t.Fatal("0 + 0 != 0 on curve for add_xxx")
	}

	for i := 0; i < iterations; i++ {

		p1 = make_random_twedwards_full(drng)
		p2.add_xxx(&p1, &NeutralElement_xtw)
		if !p2.verify_Point_on_Curve() {
			t.Fatal("x + 0 is not on curve for random x on curve in add_xxx")
		}
		if !p1.is_equal_safe_xx(&p2) {
			t.Fatal("x + 0 != x for random x in add_xxx")
		}
		p2.add_xxx(&NeutralElement_xtw, &p1)
		if !p2.verify_Point_on_Curve() {
			t.Fatal("0 + x is not on curve for random x on curve in add_xxx")
		}
		if !p1.is_equal_safe_xx(&p2) {
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
		if !res1.is_equal_safe_xx(&res2) {
			t.Fatal("Results of curve addition do not match for add_xxx and add_xxx_naive")
		}
		res2.add_xxx(&p2, &p1)
		if !res1.is_equal_safe_xx(&res2) {
			t.Fatal("x + y != y + x for random x, y with add_xxx")
		}
	}
}

func TestSingularAddition(t *testing.T) {
	var drng *rand.Rand = rand.New(rand.NewSource(666))

	var temp1 Point_xtw = make_random_twedwards_full(drng)
	var temp2, temp3, temp4, temp5 Point_xtw
	temp2.add_xxx(&temp1, &exceptionalPoint_1)
	temp3.add_xxx(&temp1, &temp2)
	temp4.add_xxx(&temp1, &temp1)
	temp5.add_xxx(&temp4, &exceptionalPoint_1)
	if temp1.IsSingular() || temp2.IsSingular() || temp4.IsSingular() || temp5.IsSingular() {
		t.Fatal("Singular point after Point addition")
	}
	if !temp3.IsSingular() {
		t.Error("Addition where singularity was expected did not result in singularity.")
	}
}

func TestPsi(t *testing.T) {
	var drng *rand.Rand = rand.New(rand.NewSource(6666))
	var temp, result1, result2 Point_xtw
	temp = make_random_twedwards_full(drng)
	result1.psi_regular_xx(&temp)
	result2.psi_exceptional_xx(&temp)
	if !result1.verify_Point_on_Curve() {
		t.Fatal("Result of Endomorphism(regular) is not on curve")
	}
	if !result2.verify_Point_on_Curve() {
		t.Fatal("Result of Endomorphism(exc) is not on curve")
	}
	if !result1.is_equal_safe_xx(&result2) {
		t.Fatal("Endomorphisms versions for regular and exceptional do not match on random point")
	}

	temp.SetZero()
	result1.psi_regular_xx(&temp)
	result2.psi_exceptional_xx(&temp)
	if !result2.IsSingular() {
		t.Error("Psi_exc(Neutral) is Non-singular")
	}

	if result1.IsSingular() {
		t.Fatal("Psi_reg(Neutral) is singular")
	}
	if !result1.verify_Point_on_Curve() {
		t.Fatal("Psi_reg(Neutral) is not on curve")
	}
	if !result1.IsZero_safe() {
		t.Fatal("Psi_reg(Neutral) != Neutral")
	}

	temp = orderTwoPoint
	result1.psi_regular_xx(&temp)
	result2.psi_exceptional_xx(&temp)
	if !result2.IsSingular() {
		t.Error("Psi_exc(affine order-2 point is non-singular")
	}
	if result1.IsSingular() {
		t.Fatal("Psi_reg(affine order-2 point) is singular")
	}
	if !result1.verify_Point_on_Curve() {
		t.Fatal("Psi_reg(affine order-2 point) is not on curve")
	}
	if !result1.IsZero_safe() {
		t.Fatal("Psi_reg(affine order-2 point) != Neutral")
	}

	// Note that the choice of Psi vs. its dual breaks the symmetry between the two points at infinity.
	temp = exceptionalPoint_1
	result1.psi_regular_xx(&temp)
	result2.psi_exceptional_xx(&temp)
	if !result1.IsSingular() {
		t.Error("Psi_reg(order-2 point1 at inifinity is non-singular")
	}
	if result2.IsSingular() {
		t.Fatal("Psi_exc(order-2 point1 at infinity is singular")
	}
	if !result2.verify_Point_on_Curve() {
		t.Fatal("Psi_exc(order-2 point1 at infinity is not on curve")
	}
	if result2.is_equal_safe_xx(&exceptionalPoint_1) {
		fmt.Println("Psi_exc(E1) == E1")
	} else if result1.is_equal_safe_xx(&exceptionalPoint_2) {
		fmt.Println("Psi_exc(E1) == E2")
	}
}
