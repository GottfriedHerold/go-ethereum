package bandersnatch

import (
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
	var temp1, temp2, temp3, result1, result2, result3 Point_xtw
	temp1 = make_random_twedwards_full(drng)
	result1.psi_xx(&temp1)
	if !result1.verify_Point_on_Curve() {
		t.Fatal("Psi(random point) is not on curve")
	}

	temp2 = make_random_twedwards_full(drng)
	temp3.add_xxx(&temp1, &temp2)
	result2.psi_xx(&temp2)
	result1.add_xxx(&result1, &result2)
	result3.psi_xx(&temp3)
	if !result1.is_equal_safe_xx(&result3) {
		t.Fatal("Psi is not homomorphic")
	}

	temp1.SetZero()
	result1.psi_xx(&temp1)
	if !result1.IsZero_safe() {
		t.Fatal("Psi(Neutral) != Neutral")
	}

	temp1 = orderTwoPoint_xtw
	result1.psi_xx(&temp1)
	if !result1.IsZero_safe() {
		t.Fatal("Psi(affine order-2 point) != Neutral")
	}

	temp2 = make_random_twedwards_full(drng)
	temp1.sub_xxx(&orderTwoPoint_xtw, &temp2)
	result1.psi_xx(&temp1)
	result2.psi_xx(&temp2)
	result3.add_xxx(&temp1, &temp2)
	if !result3.IsZero_safe() {
		t.Fatal("Psi is not homomorphic for sum = affine-order-2")
	}

	result1.psi_xx_safe(&exceptionalPoint_1)
	if !result1.is_equal_safe_xx(&orderTwoPoint_xtw) {
		t.Fatal("Psi(E1) != affine-order-2")
	}
	temp2 = make_random_twedwards_full(drng)
	temp1.sub_xxx(&exceptionalPoint_1, &temp2)
	if temp1.IsSingular() {
		t.Fatal("Unexpected singularity encountered")
	}
	result1.psi_xx_safe(&temp1)
	result2.psi_xx_safe(&temp2)
	temp3.add_xxx(&temp1, &temp2)
	if result1.IsSingular() || result2.IsSingular() || temp3.IsSingular() {
		t.Fatal("Unexpected singularity encountered")
	}
	if !temp3.is_equal_safe_xx(&exceptionalPoint_1) {
		t.Fatal("Associative Law fails when sum is E1")
	}
	result1.add_xxx(&result1, &result2) // requires add_xxx to be safe enough, which it is.
	if !result1.is_equal_safe_xx(&orderTwoPoint_xtw) {
		t.Fatal("Homomorphic properties of Psi unsatisfied when sum is E1")
	}
	result1.psi_xx_safe(&exceptionalPoint_2)
	if !result1.is_equal_safe_xx(&orderTwoPoint_xtw) {
		t.Fatal("Psi(E2) != affine-order-2 point")
	}

	temp1 = make_random_x(drng)
	result1.psi_xx(&temp1)
	result2.exp_naive_xx(&temp1, EndoEigenvalue_Int)
	if !result1.is_equal_safe_xx(&result2) {
		t.Fatal("Psi does not act as multiplication by EndoEigenvalue on random point in subgroup")
	}
}
