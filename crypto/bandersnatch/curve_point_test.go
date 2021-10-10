package bandersnatch

import (
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
}

func TestInterfaces(t *testing.T) {
	var _ CurvePointRead = &Point_xtw{}
}
