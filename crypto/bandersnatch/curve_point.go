package bandersnatch

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// Curve parameters

const (
	GroupOrder        = 0x1cfb69d4ca675f520cce760202687600ff8f87007419047174fd06b52876e7e1
	GroupOrder_string = "0x1cfb69d4ca675f520cce760202687600ff8f87007419047174fd06b52876e7e1"
)

var GroupOrder_Int *big.Int = new(big.Int).SetBytes(common.FromHex(GroupOrder_string))

const Cofactor = 4

const (
	GLSEigenvalue_string = "0x13b4f3dc4a39a493edf849562b38c72bcfc49db970a5056ed13d21408783df05"
	GLSEigenvalue        = 0x13b4f3dc4a39a493edf849562b38c72bcfc49db970a5056ed13d21408783df05
)

var GLSEigenvalue_Int *big.Int = new(big.Int).SetBytes(common.FromHex(GLSEigenvalue_string))

// parameters a, d in twisted Edwards form ax^2 + y^2 = 1 + dx^2y^2

// Note: both a and d are non-squares

const TwistedEdwardsA = -5
const (
	TwistedEdwardsD        = 0x6389c12633c267cbc66e3bf86be3b6d8cb66677177e54f92b369f2f5188d58e7
	TwistedEdwardsD_string = "0x6389c12633c267cbc66e3bf86be3b6d8cb66677177e54f92b369f2f5188d58e7"
)

var TwistedEdwardsD_Int *big.Int = new(big.Int).SetBytes(common.FromHex(TwistedEdwardsD_string))
var TwistedEdwardsD_fe = func() (ret bsFieldElement_64) { ret.SetInt(TwistedEdwardsD_Int); return }()

/*
	Caveat: Bandersnatch is typically represented as a twisted Edwards curve, which means there are singularities
	at infinity. These singularities are not in the large-prime order subgroup. (the cofactor is 4)
	and only really correspond to curve points after desingularization anyway.
	To avoid these issues, we shall assert (and check on external input by default) that all points
	are in the correct subgroup unless explicitly specified otherwise.
	Relying on the user to "manually" performing, say, cofactor clearing is error-prone, as
	the addition formulas might not be correct outside of the subgroup.
*/

// A CurvePoint represents a rational point on the bandersnatch curve in the correct subgroup.
type CurvePoint interface {
	IsZero() bool
	SetZero()
}
