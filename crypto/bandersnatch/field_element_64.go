// unused for now
package bandersnatch

import "math/bits"

// 2*modulus
const (
	mdoubled_64_0 = (2 * BaseFieldSize_untyped >> (iota * 64)) & 0xFFFFFFFF
	mdoubled_64_1
	mdoubled_64_2
	mdoubled_64_3
)

type bsFieldElement_64 struct {
	// field elements stored in low-endian 64-bit uints in Montgomery form
	words [4]uint64
}

/*
	We define operations to use on field elements via explicit functions add, sub, etc. with names matching big.Int.
	The receiver argument is overwritten by the result of the computation.
	i.e. the usual syntax is z.add(x,y) for z = x+y. Matching big.Int
	As opposed to big.Int, we do not guarantee that the functions work if &z == &x or &z == &y, because we do not
	want to limit possible optimizations.
*/

func (z *bsFieldElement_64) maybe_reduce_once() {
	var borrow uint64
	if z.words[3] > m_64_3 {
		z.words[0], borrow = bits.Sub64(z.words[0], m_64_0, 0)
		z.words[1], borrow = bits.Sub64(z.words[1], m_64_1, borrow)
		z.words[2], borrow = bits.Sub64(z.words[2], m_64_2, borrow)
		z.words[3] -= m_64_3
		z.words[3] -= borrow
	}
}

func (z *bsFieldElement_64) add(x, y *bsFieldElement_64) {
	var carry uint64
	z.words[0], carry = bits.Add64(x.words[0], y.words[0], 0)
	z.words[1], carry = bits.Add64(x.words[1], y.words[1], carry)
	z.words[2], carry = bits.Add64(x.words[2], y.words[2], carry)
	z.words[3], carry = bits.Add64(x.words[3], y.words[3], carry)
	if carry != 0 {
		z.words[0], carry = bits.Sub64(z.words[0], mdoubled_64_0, 0)
		z.words[1], carry = bits.Sub64(z.words[1], mdoubled_64_1, carry)
		z.words[2], carry = bits.Sub64(z.words[2], mdoubled_64_2, carry)
		z.words[3], carry = bits.Sub64(z.words[3], mdoubled_64_3, carry)
		if carry == 0 {
			panic(0)
		}
	}
	// else?
	z.maybe_reduce_once()
}

// Multiply 4x64 bit number by a 1x64 bit number. The result is 5x64 bits, split as 1x64 (low) + 4x64 (high), everything low-endian.
func mul_four_one_64(x *[4]uint64, y uint64) (low uint64, high [4]uint64) {
	var carry, temp uint64
	high[0], low = bits.Mul64(x[0], y)
	high[1], temp = bits.Mul64(x[1], y)
	high[0], carry = bits.Add64(high[0], temp, 0)
	high[2], temp = bits.Mul64(x[2], y)
	high[1], carry = bits.Add64(high[1], temp, carry)
	high[3], temp = bits.Mul64(x[3], y)
	high[2], carry = bits.Add64(high[2], temp, carry)
	high[3] += carry
	return
}

// This computes (target + x * y) >> 64, stores the result in target and return the uint64 shifted out (everthing low-endian)
func add_mul_shift_64(target *[4]uint64, x *[4]uint64, y uint64) (low uint64) {
	panic(0)
	return
}

// This function computes t+= (q*BaseFieldSize)/2^64 + 1, assuming no overflow.
func montgomery_step_64(t *[4]uint64, q uint64) {

}

func (z *bsFieldElement_64) mul(x, y *bsFieldElement_64) {
	var temp [4]uint64

	/* We perform Montgomery multiplication, i.e. we need to find x*y / r^4 bmod BaseFieldSize with r==2^64
	To do so, note that x*y == x*(y[0] + ry[1]+r^2y[2]+r^3y[3]), so
	x*y / r^4 == 1/r^4 x*y[0] + 1/r^3 x*y[1] + 1/r^2 x*y[2] + 1/r x*y[3],
	which can be computed as ((((x*y[0]/r + x*y[1]) /r + x*y[1]) / r + x*y[2]) /r) + x*y[3]) /r
	i.e by interleaving adding x*y[i] and dividing by r (everything is mod BaseFieldSize).
	We store the intermediate results in temp

	Dividing by r modulo BaseFieldSize is done by adding a suitable multiple of BaseFieldSize
	(which we can always do mod BaseFieldSize) s.t. the result is divisible by r and just dividing by r.
	This has the effect of reducing the size of number, thereby performing a (partial) modular reduction (Montgomery's trick)
	*/

	// Montgomery
	var reducer, carry uint64

	// reducer = reducer / modulus bmod 2^64. Note the special form of the modulus.

	temp[0], reducer = bits.Mul64(x.words[0], y.words[0])
	const negativeInverseModulus uint64 = (0xFFFFFFFF_FFFFFFFF * 0x00000001_00000001) % (1 << 64)
	reducer *= negativeInverseModulus // reduce = - (reduce + reduce >> 32)

	temp[0], carry = bits.Mul64(x.words[1], y.words[0])
	temp[1], carry = bits.Mul64(x.words[2], y.words[0])
	temp[2], carry = bits.Mul64(x.words[3], y.words[0])
	// temp[3], _ = bits.Mul64(x.words)
	panic(0)
	_ = carry
}
