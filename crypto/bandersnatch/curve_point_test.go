package bandersnatch

import (
	"fmt"
	"math/big"
	"testing"
)

func TestGlobalParameter(t *testing.T) {

	// Things would still work out (I guess), but some claims become wrong.
	if big.Jacobi(big.NewInt(TwistedEdwardsA), BaseFieldSize) == 1 {
		t.Fatal("Parameter a of curve is a square")
	}
	if big.Jacobi(TwistedEdwardsD_Int, BaseFieldSize) == 1 {
		t.Fatal("Parameter d of curve is a square")
	}
	if !sqrtDDivA_Good {
		t.Fatal("Could not create SqrtDDivA_Int")
	}
	var temp FieldElement
	temp.Square(&SqrtDDivA_fe)
	temp.multiply_by_five()
	temp.Neg(&temp)
	if !temp.IsEqual(&TwistedEdwardsD_fe) {
		t.Fatal("SqrtDDivA is not a square root of d/a")
	}
	temp.Inv(&SqrtDDivA_fe)
	var temp1, temp2 FieldElement
	temp1 = endo_a1_fe
	temp1.multiply_by_five()
	temp2.Mul(&endo_a2_fe, &TwistedEdwardsD_fe)
	fmt.Println(temp1.IsEqual(&temp2))

	var DOverA FieldElement
	DOverA.Square(&SqrtDDivA_fe)
	temp1.SetOne()
	temp2.Mul(&endo_b_fe, &DOverA)
	temp1.SubEq(&temp2) // temp1 == 1 - b D/A

	var temp3, temp4 FieldElement
	temp3 = endo_b_fe
	temp4.SetOne()
	temp4.SubEq(&endo_b_fe)
	temp4.SubEq(&endo_b_fe)
	temp4.MulEq(&DOverA)
	temp3.AddEq(&temp4) // temp3 == b + (1-2b) D/A

	var temp5 FieldElement
	temp5.Inv(&temp3)
	temp5.MulEq(&temp1)
	fmt.Println("Qutient is", temp5.String())
	temp5.Inv(&temp5)
	fmt.Println("Inverse is", temp5.String())
	temp = SqrtDDivA_fe
}

func TestInterfaces(t *testing.T) {
	var _ CurvePointRead = &Point_xtw{}
}
