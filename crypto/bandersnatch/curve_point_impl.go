package bandersnatch

import (
	"math/big"
)

/*
	Note: Suffixes like _ttt or _tta refer to the type of input point (with order output, input1 [,input2] )
	t denote extended projective,
	a denotes extended affine (i.e. Z==1)
	s denotes double-projective
*/

func (out *Point_xtw) neg_tt(input1 *Point_xtw) {
	out.x = input1.x
	out.y.Neg(&input1.y)
	out.t.Neg(&input1.t)
	out.z = input1.z
}

// This checks whether the X/Z coordinate may be in the subgroup.
func (p *Point_xtw) legendre_check_point() bool {
	var temp FieldElement
	/// p.MakeAffine()  -- removed in favour of homogenous formula
	temp.Square(&p.x)
	temp.multiply_by_five()
	var zz FieldElement
	zz.Square(&p.z)
	temp.AddEq(&zz) // temp = z^2 + 5x^2 = z^2-ax^2
	tempInt := temp.ToInt()
	result := big.Jacobi(tempInt, BaseFieldSize)
	if result == 0 {
		panic("z^2-ax^2 is 0") // Cannot happen, because a is a non-square.
	}
	return result > 0
}

func (p *Point_xtw) makeAffine_x() {
	var temp FieldElement
	if p.z.IsZero() {
		panic("Trying to make point at infinity or singular point affine")
	}
	temp.Inv(&p.z)
	p.x.MulEq(&temp)
	p.y.MulEq(&temp)
	p.t.MulEq(&temp)
	p.z.SetOne()
}
