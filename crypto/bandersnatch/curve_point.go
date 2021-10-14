package bandersnatch

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// Curve parameters

// Order of the p253-subgroup of the Bandersnatch curve. This is a 253-bit prime.
const (
	GroupOrder        = 0x1cfb69d4ca675f520cce760202687600ff8f87007419047174fd06b52876e7e1
	GroupOrder_string = "0x1cfb69d4ca675f520cce760202687600ff8f87007419047174fd06b52876e7e1"
)

const Cofactor = 4
const CurverOrder = Cofactor * GroupOrder

var GroupOrder_Int *big.Int = new(big.Int).SetBytes(common.FromHex(GroupOrder_string))
var Cofactor_Int *big.Int = big.NewInt(Cofactor)
var CurveOrder_Int *big.Int = new(big.Int).Mul(GroupOrder_Int, Cofactor_Int)

const (
	GLSEigenvalue        = 0x13b4f3dc4a39a493edf849562b38c72bcfc49db970a5056ed13d21408783df05
	GLSEigenvalue_string = "0x13b4f3dc4a39a493edf849562b38c72bcfc49db970a5056ed13d21408783df05"
)

var GLSEigenvalue_Int *big.Int = new(big.Int).SetBytes(common.FromHex(GLSEigenvalue_string))

// parameters a, d in twisted Edwards form ax^2 + y^2 = 1 + dx^2y^2

// Note: both a and d are non-squares

const TwistedEdwardsA = -5
const (
	TwistedEdwardsD        = 0x6389c12633c267cbc66e3bf86be3b6d8cb66677177e54f92b369f2f5188d58e7
	TwistedEdwardsD_string = "0x6389c12633c267cbc66e3bf86be3b6d8cb66677177e54f92b369f2f5188d58e7"
)

var (
	TwistedEdwardsD_Int *big.Int     = new(big.Int).SetBytes(common.FromHex(TwistedEdwardsD_string))
	TwistedEdwardsD_fe  FieldElement = func() (ret FieldElement) { ret.SetInt(TwistedEdwardsD_Int); return }()
)

// SqrtDDivA is a square root of d/a
const (
	SqrtDDivA        = 37446463827641770816307242315180085052603635617490163568005256780843403514038 // Note: Not in hex
	SqrtDDivA_string = "37446463827641770816307242315180085052603635617490163568005256780843403514038"
)

var (
	SqrtDDivA_Int, sqrtDDivA_Good              = new(big.Int).SetString(SqrtDDivA_string, 0)
	SqrtDDivA_fe                  FieldElement = func() (ret FieldElement) { ret.SetInt(SqrtDDivA_Int); return }()
)

/*
	Caveat: Bandersnatch is typically represented as a twisted Edwards curve, which means there are singularities
	at infinity. These singularities are not in the large-prime order subgroup. (the cofactor is 4)
	and only really correspond to curve points after desingularization anyway.
	To avoid these issues, we shall assert (and check on external input) that all points in the correct subgroup.
	Unless explicitly specified otherwise, we do not guarantee correctness on our algorithms for points outside the subgroup.
*/

// A CurvePoint represents a rational point on the bandersnatch curve in the correct subgroup.

type CurvePointRead interface {
	IsZero() bool
	X_affine() FieldElement
	X_projective() FieldElement
	Y_affine() FieldElement
	Y_projective() FieldElement
	Z_projective() FieldElement
	IsAffine() bool
	MakeAffine()
}

type CurvePointWrite interface {
	SetZero()
	Add(CurvePointRead, CurvePointRead)
	Sub(CurvePointRead, CurvePointRead)
	Neg(CurvePointRead)
	Psi(CurvePointRead)
	// ClearCofactor()
}

type CurvePoint interface {
	CurvePointRead
	CurvePointWrite
}

const (
	endo_a1        = 0x23c58c92306dbb95960f739827ac195334fcd8fa17df036c692f7ddaa306c7d4
	endo_a1_string = "0x23c58c92306dbb95960f739827ac195334fcd8fa17df036c692f7ddaa306c7d4"
	endo_a2        = 0x23c58c92306dbb96b0b30d3513b222f50d02d8ff03e5036c69317ddaa306c7d4
	endo_a2_string = "0x23c58c92306dbb96b0b30d3513b222f50d02d8ff03e5036c69317ddaa306c7d4"
	endo_b         = 0x52c9f28b828426a561f00d3a63511a882ea712770d9af4d6ee0f014d172510b4
	endo_b_string  = "0x52c9f28b828426a561f00d3a63511a882ea712770d9af4d6ee0f014d172510b4"
	// endo_c1 == - endo_b
	//c1 = 0x2123b4c7a71956a2d149cacda650bd7d2516918bf263672811f0feb1e8daef4d
)

var (
	endo_a1_fe FieldElement = initFieldElementFromString(endo_a1_string)
	endo_a2_fe FieldElement = initFieldElementFromString(endo_a2_string)
	endo_b_fe  FieldElement = initFieldElementFromString(endo_b_string)
)
