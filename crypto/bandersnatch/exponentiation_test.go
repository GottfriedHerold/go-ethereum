package bandersnatch

import (
	"math/big"
	"math/rand"
	"testing"
)

func TestSimpleExponentiation(t *testing.T) {
	const iterations = 10
	var temp1, temp2, temp3, temp4 Point_xtw
	temp1.exp_naive_xx(&example_generator_xtw, GroupOrder_Int)
	if !temp1.isZero_safe() {
		t.Fatal("Either naive exponentiation is wrong or example point not in subgroup")
	}
	var drng *rand.Rand = rand.New(rand.NewSource(1024))
	var exp1 = big.NewInt(0)
	var exp2 = big.NewInt(1)
	var exp3 = big.NewInt(-1)

	temp1 = make_random_twedwards_full(drng)
	temp2.exp_naive_xx(&temp1, exp2) // exponent is 1
	if !temp2.is_equal_safe_xx(&temp1) {
		t.Error("1 * P != P for naive exponentiation")
	}
	temp2.exp_naive_xx(&temp1, exp1) // exponent is 0
	if !temp2.is_equal_safe_xx(&NeutralElement_xtw) {
		t.Error("0 * P != Neutral element for naive exponentiation")
	}
	temp2.exp_naive_xx(&temp1, exp3)
	temp1.neg_xx(&temp1)
	if !temp1.is_equal_safe_xx(&temp2) {
		t.Error("-1 * P != -P for naive exponentiation")
	}

	var p1, p2, p3 Point_xtw
	for i := 0; i < iterations; i++ {
		p1 = make_random_twedwards_full(drng)
		p2 = make_random_twedwards_full(drng)
		p3.add_xxx(&p1, &p2)
		exp1.Rand(drng, CurveOrder_Int)
		exp2.Rand(drng, CurveOrder_Int)
		exp3.Add(exp1, exp2)
		temp1.exp_naive_xx(&p1, exp1)
		temp2.exp_naive_xx(&p2, exp1)
		temp3.exp_naive_xx(&p3, exp1)
		temp4.add_xxx(&temp1, &temp2)
		if !temp3.is_equal_safe_xx(&temp4) {
			t.Error("a * (P+Q) != a*P + a*Q for naive exponentiation")
		}
		temp2.exp_naive_xx(&p1, exp2)
		temp3.exp_naive_xx(&p1, exp3)
		temp4.add_xxx(&temp1, &temp2)
		if !temp3.is_equal_safe_xx(&temp4) {
			t.Error("(a+b) * P != a*P + b*P for naive exponentiation")
		}
	}
}

func TestQuotientGroup(t *testing.T) {
	const iterations = 100
	var drng *rand.Rand = rand.New(rand.NewSource(1024))
	var temp Point_xtw
	for i := 0; i < iterations; i++ {
		temp = make_random_twedwards_full(drng)
		temp.exp_naive_xx(&temp, GroupOrder_Int)
		if !temp.z.IsZero() {
			temp.MakeAffine()
		}
		outX := temp.X_projective()
		outY := temp.Y_projective()
		outT := temp.t
		outZ := temp.Z_projective()
		t.Log("Result is ", outX.String(), outY.String(), outT.String(), outZ.String())
	}
}
