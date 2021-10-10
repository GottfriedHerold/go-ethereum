package bandersnatch

import "math/big"

func (out *Point_xtw) exp_naive_xx(p *Point_xtw, exponent *big.Int) {
	// simple square-and-multiply
	var absexponent *big.Int
	var to_add Point_xtw = *p
	if exponent.Sign() < 0 {
		absexponent = new(big.Int).Abs(exponent)
		to_add.neg_xx(p)
	} else {
		absexponent = new(big.Int).Set(exponent)
	}
	bitlen := absexponent.BitLen()
	var accumulator Point_xtw = NeutralElement_xtw
	for i := bitlen - 1; i >= 0; i-- {
		accumulator.double_xx(&accumulator)
		if absexponent.Bit(i) == 1 {
			accumulator.add_xxx(&accumulator, &to_add)
		}
	}
	*out = accumulator
}
