// unused for now
package bandersnatch

import (
	"encoding/binary"
	"math/big"
	"math/bits"
	"math/rand"
)

// 2*modulus
const (
	mdoubled_64_0 = (2 * BaseFieldSize_untyped >> (iota * 64)) & 0xFFFFFFFF_FFFFFFFF
	mdoubled_64_1
	mdoubled_64_2
	mdoubled_64_3
)

// 2^256 - 2*modulus. This is also the Montgomery representation of 1.
// Note: Manually doing 2's complement. since writing 1<<256-2*BaseFieldSize_untyped is not portable according to the language spec.
const (
	neg_mdoubled_64_0 = 0x00000001_FFFFFFFE
	neg_mdoubled_64_1 = 0xFFFFFFFF_FFFFFFFF ^ mdoubled_64_1
	neg_mdoubled_64_2 = 0xFFFFFFFF_FFFFFFFF ^ mdoubled_64_2
	neg_mdoubled_64_3 = 0xFFFFFFFF_FFFFFFFF ^ mdoubled_64_3
)

type bsFieldElement_64 struct {
	// field elements stored in low-endian 64-bit uints in Montgomery form, i.e. words encodes a number x s.t.
	// words - x * (1<<256) == 0 (mod BaseFieldSize).
	// Note that the representation of x is actually NOT unique, as we only guarantee 0 <= words < 1<<256 - BaseFieldSize.
	// A given x might have either 1 or 2 possible representations.
	words [4]uint64
}

var bsFieldElement_64_zero bsFieldElement_64

// alternative representation of zero.
var bsFieldElement_64_zero_alt bsFieldElement_64 = bsFieldElement_64{words: [4]uint64{m_64_0, m_64_1, m_64_2, m_64_3}}

var bsFieldElement_64_one bsFieldElement_64 = bsFieldElement_64{words: [4]uint64{neg_mdoubled_64_0, neg_mdoubled_64_1, neg_mdoubled_64_2, neg_mdoubled_64_3}}

// Change the representation of z to restore the invariant that z.words + BaseFieldSize must not overflow.
func (z *bsFieldElement_64) maybe_reduce_once() {
	var borrow uint64
	if z.words[3] > m_64_3 {
		z.words[0], borrow = bits.Sub64(z.words[0], m_64_0, 0)
		z.words[1], borrow = bits.Sub64(z.words[1], m_64_1, borrow)
		z.words[2], borrow = bits.Sub64(z.words[2], m_64_2, borrow)
		z.words[3], _ = bits.Sub64(z.words[3], m_64_3, borrow)
	}
}

// Change the internal representation to a unique number in 0 <= . < BaseFieldSize
func (z *bsFieldElement_64) Normalize() {
	// Workaround for Go's lack of constexpr. Hoping for smart-ish compiler.
	var base_field_temp [4]uint64 = [4]uint64{m_64_0, m_64_1, m_64_2, m_64_3}
	for i := 3; i >= 0; i-- {
		if z.words[i] < base_field_temp[i] {
			return
		} else if z.words[i] > base_field_temp[i] {
			break
		}
	}
	var borrow uint64
	z.words[0], borrow = bits.Sub64(z.words[0], m_64_0, 0)
	z.words[1], borrow = bits.Sub64(z.words[1], m_64_1, borrow)
	z.words[2], borrow = bits.Sub64(z.words[2], m_64_2, borrow)
	z.words[3], borrow = bits.Sub64(z.words[3], m_64_3, borrow)
	if borrow != 0 {
		panic("Underflow in normalization. This should never happen.")
	}
}

// Add x + y and store the result in z
func (z *bsFieldElement_64) Add(x, y *bsFieldElement_64) {
	var carry uint64
	z.words[0], carry = bits.Add64(x.words[0], y.words[0], 0)
	z.words[1], carry = bits.Add64(x.words[1], y.words[1], carry)
	z.words[2], carry = bits.Add64(x.words[2], y.words[2], carry)
	z.words[3], carry = bits.Add64(x.words[3], y.words[3], carry)
	// At this point carry == 1 basically only happens if you do it on purpose.
	// NOTE: If carry ==1, then z.maybe_reduce_once() actually commutes with the -=mdoubled here: it won't do anything either before or after it.
	if carry != 0 {
		z.words[0], carry = bits.Sub64(z.words[0], mdoubled_64_0, 0)
		z.words[1], carry = bits.Sub64(z.words[1], mdoubled_64_1, carry)
		z.words[2], carry = bits.Sub64(z.words[2], mdoubled_64_2, carry)
		z.words[3], _ = bits.Sub64(z.words[3], mdoubled_64_3, carry)
	}
	// else?
	z.maybe_reduce_once()

}

// Subtract x - y and store the result in z
func (z *bsFieldElement_64) Sub(x, y *bsFieldElement_64) {
	var borrow uint64 // only takes values 0,1
	z.words[0], borrow = bits.Sub64(x.words[0], y.words[0], 0)
	z.words[1], borrow = bits.Sub64(x.words[1], y.words[1], borrow)
	z.words[2], borrow = bits.Sub64(x.words[2], y.words[2], borrow)
	z.words[3], borrow = bits.Sub64(x.words[3], y.words[3], borrow)
	if borrow != 0 {
		// mentally rename borrow -> carry
		if z.words[3] > 0xFFFFFFFF_FFFFFFFF-m_64_3 {
			z.words[0], borrow = bits.Add64(z.words[0], m_64_0, 0)
			z.words[1], borrow = bits.Add64(z.words[1], m_64_1, borrow)
			z.words[2], borrow = bits.Add64(z.words[2], m_64_2, borrow)
			z.words[3], _ = bits.Add64(z.words[3], m_64_3, borrow) // _ is one
		} else {
			z.words[0], borrow = bits.Add64(z.words[0], mdoubled_64_0, 0)
			z.words[1], borrow = bits.Add64(z.words[1], mdoubled_64_1, borrow)
			z.words[2], borrow = bits.Add64(z.words[2], mdoubled_64_2, borrow)
			z.words[3], _ = bits.Add64(z.words[3], mdoubled_64_3, borrow) // _ is one
			// Note: z might be > BaseFieldSize, but not by much. This is fine.
		}
	}
}

// Multiply 4x64 bit number by a 1x64 bit number. The result is 5x64 bits, split as 1x64 (low) + 4x64 (high), everything low-endian.
func mul_four_one_64(x *[4]uint64, y uint64) (low uint64, high [4]uint64) {
	var carry, mul_result_low uint64

	high[0], low = bits.Mul64(x[0], y)

	high[1], mul_result_low = bits.Mul64(x[1], y)
	high[0], carry = bits.Add64(high[0], mul_result_low, 0)

	high[2], mul_result_low = bits.Mul64(x[2], y)
	high[1], carry = bits.Add64(high[1], mul_result_low, carry)

	high[3], mul_result_low = bits.Mul64(x[3], y)
	high[2], carry = bits.Add64(high[2], mul_result_low, carry)

	high[3] += carry
	return
}

// This computes (target + x * y) >> 64, stores the result in target and return the uint64 shifted out (everything low-endian)
func add_mul_shift_64(target *[4]uint64, x *[4]uint64, y uint64) (low uint64) {

	// carry_mul_even resp. carry_mul_odd end up in target[even] resp. target[odd]
	// Could do with fewer carries, but that's more error-prone (and also this is more pipeline-friendly, not that it mattered much)

	var carry_mul_even uint64
	var carry_mul_odd uint64
	var carry_add_1 uint64
	var carry_add_2 uint64

	carry_mul_even, low = bits.Mul64(x[0], y)
	low, carry_add_2 = bits.Add64(low, target[0], 0)

	carry_mul_odd, target[0] = bits.Mul64(x[1], y)
	target[0], carry_add_1 = bits.Add64(target[0], carry_mul_even, 0)
	target[0], carry_add_2 = bits.Add64(target[0], target[1], carry_add_2)

	carry_mul_even, target[1] = bits.Mul64(x[2], y)
	target[1], carry_add_1 = bits.Add64(target[1], carry_mul_odd, carry_add_1)
	target[1], carry_add_2 = bits.Add64(target[1], target[2], carry_add_2)

	carry_mul_odd, target[2] = bits.Mul64(x[3], y)
	target[2], carry_add_1 = bits.Add64(target[2], carry_mul_even, carry_add_1)
	target[2], carry_add_2 = bits.Add64(target[2], target[3], carry_add_2)

	target[3] = carry_mul_odd + carry_add_1 + carry_add_2
	return
}

// This function computes t+= (q*BaseFieldSize)/2^64 + 1, assuming no overflow.
func montgomery_step_64(t *[4]uint64, q uint64) {
	var low, high, carry1, carry2 uint64

	high, _ = bits.Mul64(q, m_64_0)
	t[0], carry1 = bits.Add64(t[0], high, 1) // After this, carry1 needs to go in t[1]

	high, low = bits.Mul64(q, m_64_1)
	t[0], carry2 = bits.Add64(t[0], low, 0)       // After this, carry2 needs to go in t[1]
	t[1], carry2 = bits.Add64(t[1], high, carry2) // After this, carry2 needs to go in t[2]

	high, low = bits.Mul64(q, m_64_2)
	t[1], carry1 = bits.Add64(t[1], low, carry1)  // After this, carry1 needs to go in t[2]
	t[2], carry1 = bits.Add64(t[2], high, carry1) // After this, carry1 needs to go in t[3]

	high, low = bits.Mul64(q, m_64_3)
	t[2], carry2 = bits.Add64(t[2], low, carry2) // After this, carry2 needs to go in t[3]
	t[3], carry1 = bits.Add64(t[3], high+carry1, carry2)

	if carry1 != 0 {
		panic("Overflow in montgomery step")
	}

}

func (z *bsFieldElement_64) Mul(x, y *bsFieldElement_64) {

	/*
		We perform Montgomery multiplication, i.e. we need to find x*y / r^4 bmod BaseFieldSize with r==2^64
		To do so, note that x*y == x*(y[0] + ry[1]+r^2y[2]+r^3y[3]), so
		x*y / r^4 == 1/r^4 x*y[0] + 1/r^3 x*y[1] + 1/r^2 x*y[2] + 1/r x*y[3],
		which can be computed as ((((x*y[0]/r + x*y[1]) /r + x*y[1]) / r + x*y[2]) /r) + x*y[3]) /r
		i.e by interleaving adding x*y[i] and dividing by r (everything is mod BaseFieldSize).
		We store the intermediate results in temp

		Dividing by r modulo BaseFieldSize is done by adding a suitable multiple of BaseFieldSize
		(which we can always do mod BaseFieldSize) s.t. the result is divisible by r and just dividing by r.
		This has the effect of reducing the size of number, thereby performing a (partial) modular reduction (Montgomery's trick)
	*/

	// temp holds the result of computation so far. We only write into z at the end, because z might alias x or y.
	var temp [4]uint64

	// -1/Modulus mod r.
	const negativeInverseModulus = (0xFFFFFFFF_FFFFFFFF * 0x00000001_00000001) % (1 << 64)
	const negativeInverseModulus_uint uint64 = negativeInverseModulus

	var reducer uint64

	reducer, temp = mul_four_one_64(&x.words, y.words[0]) // NOTE: temp <= B - floor(B/r) - 1  <= B + floor(M/r), see overflow analysis below

	// If reducer == 0, then temp == x*y[0]/r.
	// Otherwise, we need to compute temp = ([temp, reducer] + BaseFieldSize * (reducer * negativeInverseModulus mod r)) / r
	// Note that we know exactly what happens in the least significant uint64 in the addition (result 0, carry 1). Be aware that carry 1 relies on reducer != 0, hence the if...
	if reducer != 0 {
		montgomery_step_64(&temp, reducer*negativeInverseModulus_uint)
	}

	reducer = add_mul_shift_64(&temp, &x.words, y.words[1])
	if reducer != 0 {
		montgomery_step_64(&temp, reducer*negativeInverseModulus_uint)
	}

	reducer = add_mul_shift_64(&temp, &x.words, y.words[2])
	if reducer != 0 {
		montgomery_step_64(&temp, reducer*negativeInverseModulus_uint)
	}

	reducer = add_mul_shift_64(&temp, &x.words, y.words[3])
	if reducer != 0 {
		// TODO: Store directly into z
		montgomery_step_64(&temp, reducer*negativeInverseModulus_uint)
	}

	/*
		Overflow analysis:
		Let B:= 2^256 - BaseFieldSize - 1. We know that 0<= x,y <= B and need to ensure that 0<=z<=B to maintain our invariants:

		(1) If temp <= B + M (which is 2^256 - 1, so this condition is somewhat vacuous) and x <= B, then after applying add_mul_shift_64(&temp, x, y), we have
		temp <= (B + M + B * (r-1)) / r <= B + floor(M/r)

		(2) If temp <= B + floor(M/r) is satisfied and we compute montgomery_step_64(&temp, something), we afterwards obtain
		temp <= B + floor(M/r) + floor(M*(r-1)/r) + 1 == B + M  (this implies there is no overflow inside montgomery_step_64)

		Since the end result might be bigger than B, we may need to reduce by M, but once is enough.
	*/

	z.words = temp
	z.maybe_reduce_once()
}

func (z *bsFieldElement_64) IsZero() bool {
	return (z.words[0]|z.words[1]|z.words[2]|z.words[3] == 0) || (*z == bsFieldElement_64_zero_alt)
}

func (z *bsFieldElement_64) IsOne() bool {
	return *z == bsFieldElement_64_one
}

func (z *bsFieldElement_64) SetOne() {
	z.words = bsFieldElement_64_one.words
}

func (z *bsFieldElement_64) SetZero() {
	z.words = bsFieldElement_64_zero.words
}

// converts a low-endian 4xuint64 array to Int, without any Montgomery conversions
func uintarrayToInt(z *[4]uint64) *big.Int {
	var big_endian_byte_slice [32]byte
	binary.BigEndian.PutUint64(big_endian_byte_slice[0:8], z[3])
	binary.BigEndian.PutUint64(big_endian_byte_slice[8:16], z[2])
	binary.BigEndian.PutUint64(big_endian_byte_slice[16:24], z[1])
	binary.BigEndian.PutUint64(big_endian_byte_slice[24:32], z[0])
	return new(big.Int).SetBytes(big_endian_byte_slice[:])
}

// converts a big.Int to a low-endian [4]uint64 array without Montgomery conversions.
// We assume 0 <= x < 2^256
func intTouintarray(x *big.Int) (result [4]uint64) {
	// As this is an internal function, panic is OK for error handling.
	if x.Sign() < 0 {
		panic("Trying to convert negative big Int")
	}
	if x.BitLen() > 256 {
		panic("big Int too large to fit.")
	}
	var big_endian_byte_slice [32]byte
	x.FillBytes(big_endian_byte_slice[:])
	result[0] = binary.BigEndian.Uint64(big_endian_byte_slice[24:32])
	result[1] = binary.BigEndian.Uint64(big_endian_byte_slice[16:24])
	result[2] = binary.BigEndian.Uint64(big_endian_byte_slice[8:16])
	result[3] = binary.BigEndian.Uint64(big_endian_byte_slice[0:8])
	return
}

func (z *bsFieldElement_64) ToInt() *big.Int {

	// This represents 1/2^256 in Montgomery form
	// temp := bsFieldElement_64{words: [4]uint64{1, 0, 0, 0}}

	// temp.words is now NOT in Montgomery form. This can be done more efficiently if needed.
	// temp.Mul(&temp, z)
	// temp.Normalize()

	t := uintarrayToInt(&z.words)

	var converter *big.Int = big.NewInt(1)
	converter.Lsh(converter, 256)
	converter.ModInverse(converter, BaseFieldSize)
	t.Mul(t, converter)
	t.Mod(t, BaseFieldSize)

	return t
}

func (z *bsFieldElement_64) SetInt(v *big.Int) {
	sign := v.Sign()
	w := new(big.Int).Set(v)
	w.Abs(w)

	// Can be done much more efficiently if desired, but we do not convert often.
	w.Lsh(w, 256)
	w.Mod(w, BaseFieldSize)
	if sign < 0 {
		w.Sub(BaseFieldSize, w)
	}
	z.words = intTouintarray(w)
}

// Generate uniformly random number. Note that this is not crypto-grade randomness. Testing only.
// We do NOT guarantee that the distribution is even close to uniform.
func (z *bsFieldElement_64) setRandomUnsafe(rnd *rand.Rand) {

	// Not the most efficient way, but for testing purposes we want the _64 and _8 variants to have the same output for given rnd
	var xInt *big.Int = new(big.Int).Rand(rnd, BaseFieldSize)
	z.SetInt(xInt)
}

// Multiplicative Inverse
func (z *bsFieldElement_64) Inv(x *bsFieldElement_64) {

	// Slow, but rarely used anyway (due to working in projective coordinates)
	t := x.ToInt()
	t.ModInverse(t, BaseFieldSize)
	z.SetInt(t)
}

// Checks whether z == x (mod BaseFieldSize)
func (z *bsFieldElement_64) Compare(x *bsFieldElement_64) bool {
	// There are at most 2 possible representations per field element and they differ by exactly BaseFieldSize.
	// So it is enough to reduce the larger one, provided it is much larger.
	switch {
	case z.words[3] == x.words[3]:
		return *z == *x
	case z.words[3] > x.words[3]:
		// Note that RHS cannot overflow due to invariant
		if z.words[3] > x.words[3]+(m_64_3-1) {
			z.Normalize()
			return *z == *x
		} else {
			return false
		}
	case z.words[3] < x.words[3]:
		// Note that RHS cannot overflow due to invariant
		if x.words[3] > z.words[3]+(m_64_3-1) {
			x.Normalize()
			return *z == *x
		} else {
			return false
		}
	// needed to make golang not complain about missing return in all branches.
	default:
		panic("This cannot happen")
	}
}
