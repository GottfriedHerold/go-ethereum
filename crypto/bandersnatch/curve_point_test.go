package bandersnatch

import (
	"math/big"
	"testing"
)

func TestGlobalParameter(t *testing.T) {
	if big.Jacobi(big.NewInt(TwistedEdwardsA), BaseFieldSize) == 1 {
		t.Fatal("Parameter a of curve is a square")
	}
	if big.Jacobi(TwistedEdwardsD_Int, BaseFieldSize) == 1 {
		t.Fatal("Parameter d of curve is a square")
	}
	var temp FieldElement
	temp.Square(&SqrtDDivA_fe)
	temp.multiply_by_five()
	temp.Neg(&temp)
	if !temp.IsEqual(&TwistedEdwardsD_fe) {
		t.Fatal("SqrtDDivA is not a square root of d/a")
	}
}

func TestInterfaces(t *testing.T) {
	var _ CurvePointRead = &Point_xtw{}
}
