package bandersnatch

import (
	"fmt"
	"math/big"
	"testing"
)

func TestPlayground(t *testing.T) {
	x := big.NewInt(1)
	x.Lsh(x, 256)
	x.Mod(x, BaseFieldSize)
	t.Logf("%x", x)
}

func TestPowersOfDOverA(t *testing.T) {
	for i := -4; i < 5; i++ {
		temp := powerOfSqrtDOverA(i, false)
		fmt.Println("Sqrt(D/A) to the power ", i, " is ", temp.String())
		temp = powerOfSqrtDOverA(i, true)
		fmt.Println("negative of that is ", temp.String())

	}
	var temp1, temp2 FieldElement
	temp1 = powerOfSqrtDOverA(-1, false)
	temp1.MulEq(&TwistedEdwardsD_fe)
	fmt.Println("sqrt(ad) is ", temp1.String())
	temp2.Neg(&temp1)
	fmt.Println("Negative of that is ", temp2.String())
	temp2.Square(&temp1)
	fmt.Println("a*d is ", temp2.String())
	temp2.Neg(&temp2)
	fmt.Println("-a*d is ", temp2.String())

	fmt.Println("Endo a_1 is", endo_a1_fe.String())
	fmt.Println("Endo a_2 is", endo_a2_fe.String())
	fmt.Println("Endo b is", endo_b_fe.String())

	temp1 = SqrtDDivA_fe
	temp1.SubEq(&FieldElementOne)
	temp1.Square(&temp1)
	fmt.Println("!!!", temp1.String(), "!!!")

	temp1.SetOne()
	temp1.multiply_by_five()
	temp1.Inv(&temp1)
	temp1.MulEq(&endo_a2_fe)
	temp1.Neg(&temp1)
	fmt.Println("-a2/5 = ", temp1.String())
	var sqrt2 FieldElement
	sqrt2 = SqrtDDivA_fe
	sqrt2.SubEq(&FieldElementOne)
	temp1.Square(&sqrt2)
	temp1.SubEq(&FieldElementOne)
	temp1.SubEq(&FieldElementOne)
	if !temp1.IsZero() {
		panic(0)
	}
	var i uint64
	for i = 0; i < 100; i++ {
		temp2.SetUInt64(i)
		temp1.Mul(&temp2, &sqrt2)
		fmt.Println(i, "*sqrt(2) = ", temp1.String())
		temp1.Neg(&temp1)
		fmt.Println("(-",i,")*sqrt(2) = ", temp1.String())
	}
}

func powerOfSqrtDOverA(power int, negate_sign bool) FieldElement {
	var acc FieldElement = FieldElementOne
	var m FieldElement = SqrtDDivA_fe
	if power < 0 {
		m.Inv(&m)
		power = -power
	}
	for i := 0; i < power; i++ {
		acc.MulEq(&m)
	}
	if negate_sign {
		acc.Neg(&acc)
	}
	return acc
}
