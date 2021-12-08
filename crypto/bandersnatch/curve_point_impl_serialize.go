package bandersnatch

import (
	"encoding/binary"
	"errors"
	"io"
)

// Note: If X/Z is not on the curve, we might get either a "not on curve" or "not in subgroup" error.
var ErrXNotInSubgroup = errors.New("received affine X coordinate does not correspond to any point in the p253 subgroup of the Bandersnatch curve")
var ErrXNotOnCurve = errors.New("received affine X coordinate does not correspond to any (finite, rational) point of the Bandersnatch curve")

var ErrNotInSubgroup = errors.New("deserialization: received affine X and Y coordinates do not correspond to a point in the p253 subgroup of the Bandersnatch curve")
var ErrNotOnCurve = errors.New("deserialization: received affine X and Y corrdinates do not correspond to a point on the Bandersnatch curve")
var ErrWrongSignY = errors.New("deserialization: encountered affine Y coordinate with unexpected Sign bit")
var ErrUnrecognizedFormat = errors.New("deserialization: could not automatically detect serialization format")

func (p *Point_axtw) specialSerialzeXCoo_a() (ret FieldElement) {
	ret = p.x
	if p.y.Sign() < 0 {
		ret.NegEq()
	}
	return
}

func (p *Point_axtw) specialSerialzeYCoo_a() (ret FieldElement) {
	ret = p.y
	if ret.Sign() < 0 {
		ret.NegEq()
	}
	return
}

func mapToFieldElement(input CurvePointRead) (ret FieldElement) {
	ret = input.Y_projective()
	ret.InvEq()
	temp := input.X_projective()
	ret.MulEq(&temp)
	return
}

func (p *Point_axtw) SerializeShort(output io.Writer) (bytes_written int, err error) {
	xAbsy := p.specialSerialzeXCoo_a()
	bytes_written, err = xAbsy.Serialize(output, binary.BigEndian)
	return
}

func (p *Point_axtw) SerializeLong(output io.Writer) (bytes_written int, err error) {
	temp := p.specialSerialzeYCoo_a()
	bytes_written, err = temp.SerializeWithPrefix(output, PrefixBits(0b10), 2, binary.BigEndian)
	if err != nil {
		return
	}
	bytes_just_written, err := p.SerializeShort(output)
	bytes_written += bytes_just_written
	return
}

func affineFromXSignY(xSignY *FieldElement, trusted bool) (ret Point_axtw, err error) {
	ret.x = *xSignY // xSignY is x * Sign(y), which is correct for ret.x up to sign.

	// Note that recoverYFromXAffine only depends on the square of x, so the sign of xSignY does not matter.
	ret.y, err = recoverYFromXAffine(xSignY, !trusted)
	if err != nil {
		return
	}

	// p.x, p.y are now guaranteed to satisfy the curve equation (we we would set p.t := p.x * p.y, which we will do later).
	// The +/- ambiguity of both p.x and p.y corresponds to the set of 4 points of the form {P, -P, P+A, -P+A} for the affine 2-torsion point A.
	// Due to working mod A, we just need to fix the sign:
	if ret.y.Sign() < 0 {
		ret.y.NegEq() // p.x.NegEq() would work as well, giving a point that differs by +A
	}

	// Set t coordinate correctly:
	ret.t.Mul(xSignY, &ret.y)
	return
}

func affineFromXYSignY(xTemp *FieldElement, yTemp *FieldElement, trusted bool) (ret Point_axtw, err error) {
	ret.x = *xTemp
	ret.y = *yTemp
	ret.t.Mul(xTemp, yTemp)
	if !trusted {
		if yTemp.Sign() <= 0 {
			err = ErrWrongSignY
			return
		}

		// We compute 1-ax^2 - y^2 + dt^2, which is 0 iff the point is on the curve (and finite).
		// Observe that the subexpression 1-ax^2 is also used in the subgroup check, so we do that along the way.
		// We reuse xTemp and yTemp as temporaries, using yTemp as accumulator.
		yTemp.Square(xTemp) // x^2

		yTemp.multiply_by_five()      // 5x^2 == -ax^2
		yTemp.AddEq(&FieldElementOne) // 1+5x^2 == 1-ax^2

		if yTemp.Jacobi() < 0 {
			err = ErrNotInSubgroup
			// no return. This way, if we have both "not on curve" and "not in subgroup", we get "not on curve", which is more informative.
		}

		xTemp.Square(&ret.y) // y^2
		yTemp.SubEq(xTemp)   // 1-ax^2 - y^2

		xTemp.Square(&ret.t)             // t^2 == x^2y^2
		xTemp.MulEq(&TwistedEdwardsD_fe) // dt^2
		yTemp.AddEq(xTemp)               // 1 - ax^2 - y^2 + dt^2
		if !yTemp.IsZero() {
			err = ErrNotOnCurve
		}
	}
	return
}

func (p *Point_axtw) DeserializeShort(input io.Reader, trusted bool) (bytes_read int, err error) {
	var NonNormalized bool = false // special error flag for reading inputs that are not in the range 0<=. < BaseFieldSize. This error needs special treatment.

	var xTemp FieldElement
	// Read from input. Note that Deserialization gives x * Sign(y), so p.x is only correct up to sign.
	bytes_read, err = xTemp.DeserializeWithPrefix(input, PrefixBits(0), 1, binary.BigEndian)
	if err != nil {
		// If we get a ErrNonNormalizedDeserialization, we continue as if no error had occurred, but remember the error to return it in the end (if no other error happens).
		if err == ErrNonNormalizedDeserialization {
			NonNormalized = true
		} else {
			return
		}
	}

	// We write to temp instead of directly to p. This way, p is untouched on errors others than ErrNonNormalizedDeserialization.
	temp, err := affineFromXSignY(&xTemp, trusted)
	if err == nil {
		*p = temp
		if NonNormalized {
			err = ErrNonNormalizedDeserialization
		}
	}

	// If NonNormalized was set, we return ErrNonNormalizedDeserializtion as error, but the point is actually correct.
	return
}

func (p *Point_axtw) DeserializeLong(input io.Reader, trusted bool) (bytes_read int, err error) {
	var NonNormalized bool = false // special error flag for reading inputs that are not in the range 0<=. < BaseFieldSize. This error needs special treatment

	var ySignY, xSignY FieldElement
	bytes_read, err = ySignY.DeserializeWithPrefix(input, PrefixBits(0b10), 2, binary.BigEndian)

	// Abort if error was encountered, unless the error was NonNormalizedDeserialization.
	if err != nil {
		if err == ErrNonNormalizedDeserialization {
			NonNormalized = true
		} else {
			return
		}
	}

	bytes_just_read, err := xSignY.DeserializeWithPrefix(input, PrefixBits(0b0), 1, binary.BigEndian)
	bytes_read += bytes_just_read
	if err != nil {
		if err == ErrNonNormalizedDeserialization {
			NonNormalized = true
		} else {
			return
		}
	}

	// If we get here, we got no error other than ErrNonNormalizedDeserialization so far.
	// We write to temp instead of directly to p, since we only write if there is no error.
	temp, err := affineFromXYSignY(&xSignY, &ySignY, trusted)
	if err == nil {
		*p = temp
		if NonNormalized {
			err = ErrNonNormalizedDeserialization
		}
	}
	return
}

func (p *Point_axtw) DeserializeAuto(input io.Reader, trusted bool) (bytes_read int, err error) {
	var fieldElement_read FieldElement
	var prefix_read PrefixBits
	var temp Point_axtw
	bytes_read, prefix_read, err = fieldElement_read.deserializeAndGetPrefix(input, 1, binary.BigEndian)
	if err == ErrNonNormalizedDeserialization {
		err = ErrUnrecognizedFormat
	}
	if err != nil {
		return
	}
	if prefix_read == PrefixBits(0b0) {
		temp, err = affineFromXSignY(&fieldElement_read, trusted)
		if err != nil {
			*p = temp
		}
		return
	} else if prefix_read == PrefixBits(0b1) {
		if fieldElement_read.Sign() < 0 {
			err = ErrUnrecognizedFormat
			return
		}
		// Actually, the prefix must have beein 0b10, since otherwise we would either hit ErrNonNormalizedDeserialization or the Sign() < 0 above.
		var fieldElement2_read FieldElement
		var bytes_just_read int
		bytes_just_read, err = fieldElement2_read.DeserializeWithPrefix(input, PrefixBits(0b0), 1, binary.BigEndian)
		bytes_read += bytes_just_read
		if err == ErrNonNormalizedDeserialization {
			err = ErrUnrecognizedFormat
		}
		if err != nil {
			return
		}
		temp, err = affineFromXYSignY(&fieldElement2_read, &fieldElement_read, trusted)
		if err == nil {
			*p = temp
		}
		return
	} else {
		panic("This cannot happen")
	}
}

func recoverYFromXAffine(x *FieldElement, checkSubgroup bool) (y FieldElement, err error) {

	// We have y^2 = (1-ax^2) / (1-dx^2)
	// So, we first compute (1-ax^2) / 1-dx^2
	var num, denom FieldElement

	num.Square(x)                        // x^2, only compute this once
	denom.Mul(&num, &TwistedEdwardsD_fe) // dx^2
	num.multiply_by_five()               // 5x^2 = -ax^2
	num.AddEq(&FieldElementOne)          // 1 - ax^2
	denom.Sub(&FieldElementOne, &denom)  // 1 - dx^2
	// Note that x is in the correct subgroup iff *both* num and denom are squares
	if checkSubgroup {
		if num.Jacobi() < 0 {
			err = ErrXNotInSubgroup
			return
		}
	}
	num.DivideEq(&denom) // (1-ax^2)/(1-dx^2). Note that 1-dx^2 cannot be 0, as d is a non-square.
	if !y.SquareRoot(&num) {
		err = ErrXNotOnCurve
		return
	}
	err = nil // err is nil at this point anyway, but we prefer to be explicit.
	return
}

// isPointOnCurve checks whether the given point is actually on the curve.
// Note: This does NOT verify that the point is in the correct subgroup.
// Note2: On encountering singular values (0:0:0:0), we just return false *without* calling any error handler.
// Note3: This function is only provided for xtw
func (p *Point_xtw) isPointOnCurve() bool {

	// Singular points are not on the curve
	if p.IsSingular() {
		return false
	}

	// check whether x*y == t*z
	var u, v FieldElement
	u.Mul(&p.x, &p.y)
	v.Mul(&p.t, &p.z)
	if !u.IsEqual(&v) {
		return false
	}

	// We now check the main curve equation, i.e. whether ax^2 + y^2 == z^2 + dt^2
	u.Mul(&p.t, &p.t)
	u.MulEq(&TwistedEdwardsD_fe) // u = d*t^2
	v.Mul(&p.z, &p.z)
	u.AddEq(&v) // u= dt^2 + z^2
	v.Mul(&p.y, &p.y)
	u.SubEq(&v) // u = z^2 + dt^2 - y^2
	v.Mul(&p.x, &p.x)
	v.multiply_by_five()
	u.AddEq(&v) // u = z^2 + dt^2 - y^2 + 5x^2 ==  z^2 + dt^2 - y^2 - ax^2
	return u.IsZero()
}

// checkLegendreX(X/Z) checks whether the provided x=X/Z value may be the x-coordinate of a point in the subgroup spanned by p253 and A, assuming the curve equation has a rational solution for the given X/Z.
func checkLegendreX(x FieldElement) bool {
	// x is passed by value. We use it as a temporary.
	x.SquareEq()
	x.multiply_by_five()
	x.AddEq(&FieldElementOne) // 1 + 5x^2 = 1-ax^2
	return x.Jacobi() >= 0    // cannot be ==0, since a is a non-square
}

// checkLegendreX2(x) == checkLegendreX iff an rational y-coo satisfying the curve equation exists.
func checkLegendreX2(x FieldElement) bool {
	x.SquareEq()
	x.MulEq(&TwistedEdwardsD_fe)
	x.Sub(&FieldElementOne, &x) // 1 - dx^2
	return x.Jacobi() >= 0      // cannot be ==0, since d is a non-square
}

// This checks whether the X/Z coordinate may be in the subgroup spanned by p253 and A.
// Note that since this is called on a Point_xtw, we assume thay y is set correctly (we do not use y, but we need that y exists for the test to be sufficient)
func (p *Point_xtw) legendre_check_point() bool {
	var temp FieldElement
	/// p.MakeAffine()  -- removed in favour of homogenous formula
	temp.Square(&p.x)
	temp.multiply_by_five()
	var zz FieldElement
	zz.Square(&p.z)
	temp.AddEq(&zz) // temp = z^2 + 5x^2 = z^2-ax^2
	result := temp.Jacobi()
	if result == 0 {
		panic("Jacobi symbol of z^2-ax^2 is 0") // Cannot happen, because a is a non-square.
	}
	return result > 0
}
